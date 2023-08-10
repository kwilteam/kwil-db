package snapshots

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/kwilteam/kwil-db/pkg/snapshots"
	"github.com/kwilteam/kwil-db/pkg/utils"
)

// Receives snapshot chunks from the network and writes them to disk & restore the DB from the snapshot chunks
type Bootstrapper struct {
	tempDir       string
	dbDir         string
	activeSession *BootstrapSession
	dbRestored    bool
}

type BootstrapSession struct {
	ready            bool
	totalChunks      uint32
	chunksReceived   uint32
	snapshotMetadata *snapshots.Snapshot
	chunkInfo        map[uint32]bool
	refetchChunks    map[uint32]bool
	restoreFailed    bool
}

func NewBootstrapper(tempDir string, dbDir string) *Bootstrapper {
	return &Bootstrapper{
		tempDir:       tempDir,
		dbDir:         dbDir,
		activeSession: nil,
		dbRestored:    false,
	}
}

func (b *Bootstrapper) IsDBRestored() bool {
	if b.dbRestored || b.activeSession.restoreFailed {
		b.endBootstrapSession()
	}

	return b.dbRestored
}

func (b *Bootstrapper) OfferSnapshot(snapshot *snapshots.Snapshot) error {
	if b.validateSnapshot(snapshot) != nil {
		return fmt.Errorf("invalid snapshot")
	}
	return b.beginBootstrapSession(snapshot)
}

/*
Validates the chunk and writes it to disk & When all chunks are received, it restores the DB from the snapshot chunks
*/
func (b *Bootstrapper) ApplySnapshotChunk(chunk []byte, index uint32) ([]uint32, error) {
	b.clearRefetchChunks()
	if b.activeSession == nil {
		return nil, fmt.Errorf("no active bootstrap session")
	}

	// If chunk is already accepted or if in db restore process, return
	if b.activeSession.chunkInfo[index] || b.activeSession.ready {
		return nil, nil
	}

	format := b.activeSession.snapshotMetadata.Format
	err := b.validateChunk(chunk, index, format)
	if err != nil {
		b.activeSession.refetchChunks[index] = true
		return b.refetchChunks(), err
	}

	err = b.writeChunk(chunk, index, format)
	if err != nil {
		return nil, err
	}
	b.activeSession.chunksReceived++
	b.activeSession.chunkInfo[index] = true

	if !b.readyToBootstrap() {
		return nil, nil
	}

	return b.restoreDB()
}

func (b *Bootstrapper) validateSnapshot(snapshot *snapshots.Snapshot) error {
	// TODO: What's a valid snapshot?
	return nil
}

/*
Chunk Validation:
  - Chunk boundaries
  - Chunk hash
*/
func (b *Bootstrapper) validateChunk(chunk []byte, index uint32, format uint32) error {
	if (b.activeSession == nil) || (b.activeSession.snapshotMetadata == nil) {
		return fmt.Errorf("invalid bootstrap session")
	}

	if len(chunk) == 0 || len(chunk) < snapshots.BoundaryLen {
		return fmt.Errorf("invalid chunk length")
	}

	if !bytes.Equal(chunk[:snapshots.BeginLen], snapshots.ChunkBegin) {
		return fmt.Errorf("invalid chunk begin")
	}

	if !bytes.Equal(chunk[len(chunk)-snapshots.EndLen:], snapshots.ChunkEnd) {
		return fmt.Errorf("invalid chunk end")
	}

	hash := sha256.Sum256(chunk)
	chunkHash, ok := b.activeSession.snapshotMetadata.Metadata.ChunkHashes[index]
	if !ok {
		return fmt.Errorf("invalid chunk info")
	}
	if !bytes.Equal(hash[:], chunkHash) {
		return fmt.Errorf("invalid chunk hash")
	}

	return nil
}

func (b *Bootstrapper) beginBootstrapSession(snapshot *snapshots.Snapshot) error {
	if b.activeSession != nil {
		return fmt.Errorf("bootstrap session already active")
	}

	// create temp dir
	err := utils.CreateDirIfNeeded(b.tempDir)
	if err != nil {
		return err
	}

	b.activeSession = &BootstrapSession{
		ready:            false,
		totalChunks:      snapshot.ChunkCount,
		chunksReceived:   0,
		snapshotMetadata: snapshot,
		chunkInfo:        make(map[uint32]bool, snapshot.ChunkCount),
		refetchChunks:    make(map[uint32]bool, snapshot.ChunkCount),
		restoreFailed:    false,
	}

	for i := uint32(0); i < snapshot.ChunkCount; i++ {
		b.activeSession.chunkInfo[i] = false
	}

	return nil
}

func (b *Bootstrapper) endBootstrapSession() error {
	if b.activeSession == nil {
		return fmt.Errorf("no active bootstrap session")
	}
	b.activeSession = nil

	// delete temp dir
	err := os.RemoveAll(b.tempDir)
	if err != nil {
		return err
	}
	return nil
}

func (b *Bootstrapper) readyToBootstrap() bool {
	if b.activeSession == nil || b.activeSession.chunksReceived != b.activeSession.totalChunks {
		return false
	}
	b.activeSession.ready = true
	return true
}

// TODO: Maintain a list of chunks to be refetched: hash mismatch, chunk not received
func (b *Bootstrapper) restoreDB() ([]uint32, error) {
	// Go through each file in snapshot and read its chunks, restore the file, validate its hash
	fileInfo := b.activeSession.snapshotMetadata.Metadata.FileInfo
	var wg sync.WaitGroup
	for fileName := range fileInfo {
		wg.Add(1)
		go func(fileName string) {
			defer wg.Done()
			err := b.restoreDBFile(fileName)
			if err != nil {
				b.activeSession.restoreFailed = true
				os.Remove(filepath.Join(b.dbDir, fileName))
			}
		}(fileName)
	}
	wg.Wait()

	if b.activeSession.restoreFailed {
		return nil, fmt.Errorf("db restore failure")
	}
	b.dbRestored = true
	return nil, nil
}

func (b *Bootstrapper) restoreDBFile(fileName string) error {
	fileInfo := b.activeSession.snapshotMetadata.Metadata.FileInfo[fileName]
	beginIdx := fileInfo.BeginIdx
	endIdx := fileInfo.EndIdx

	for i := beginIdx; i <= endIdx; i++ {
		chunkfile := filepath.Join(b.tempDir, fmt.Sprintf("Chunk_%d_%d", b.activeSession.snapshotMetadata.Format, i))
		err := snapshots.CopyChunkFile(chunkfile, filepath.Join(b.dbDir, fileName))
		if err != nil {
			return err
		}
	}

	fileHash, err := utils.HashFile(filepath.Join(b.dbDir, fileName))
	if err != nil {
		return err
	}

	if !bytes.Equal(fileHash, fileInfo.Hash) {
		//b.refetchFileChunks(fileName)
		return fmt.Errorf("invalid file hash")
	}
	return nil
}

func (b *Bootstrapper) writeChunk(chunk []byte, index uint32, format uint32) error {
	chunkFile := filepath.Join(b.tempDir, fmt.Sprintf("Chunk_%d_%d", format, index))
	return utils.WriteFile(chunkFile, chunk)
}

func (b *Bootstrapper) refetchChunks() []uint32 {
	var chunks []uint32
	for chunk := range b.activeSession.refetchChunks {
		chunks = append(chunks, chunk)
	}
	return chunks
}

func (b *Bootstrapper) clearRefetchChunks() {
	b.activeSession.refetchChunks = make(map[uint32]bool)
}

func (b *Bootstrapper) refetchFileChunks(filename string) {
	fileInfo := b.activeSession.snapshotMetadata.Metadata.FileInfo[filename]
	beginIdx := fileInfo.BeginIdx
	endIdx := fileInfo.EndIdx

	for i := beginIdx; i <= endIdx; i++ {
		b.activeSession.refetchChunks[i] = true
	}
}

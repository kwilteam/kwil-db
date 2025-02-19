// This file implements a naive KV persistent solution using a file, changes made
// through the exposed functions will be written to files on every invoking.
// This state won't grow, for signer svc, this is good enough.

package signersvc

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

type voteRecord struct {
	RewardRoot  []byte `json:"reward_root"`
	BlockHeight int64  `json:"block_height"`
	BlockHash   string `json:"block_hash"`
	SafeNonce   uint64 `json:"safe_nonce"`
}

// State is a naive kv impl used by singer rewardSigner.
type State struct {
	path string

	mu sync.Mutex

	data map[string]*voteRecord // target => latest vote record
}

// _sync will write State on to disk if it's loaded from disk.
func (s *State) _sync() error {
	if s.path == "" {
		return nil
	}

	tmpPath := s.path + ".tmp"

	err := os.RemoveAll(tmpPath)
	if err != nil {
		return fmt.Errorf("ensure no tmp file: %w", err)
	}

	tmpFile, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer tmpFile.Close()

	err = json.NewEncoder(tmpFile).Encode(s.data)
	if err != nil {
		return fmt.Errorf("write state to file: %w", err)
	}

	err = tmpFile.Sync()
	if err != nil {
		return fmt.Errorf("file sync: %w", err)
	}

	err = os.Rename(tmpPath, s.path)
	if err != nil {
		return fmt.Errorf("")
	}

	return err
}

// UpdateLastVote updates the latest vote record, and syncs the changes to disk.
func (s *State) UpdateLastVote(target string, newVote *voteRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[target] = newVote

	return s._sync()
}

func (s *State) LastVote(target string) *voteRecord {
	s.mu.Lock()
	defer s.mu.Unlock()

	if record, ok := s.data[target]; ok {
		return record
	}

	return nil
}

// LoadStateFromFile load the state from a file.
func LoadStateFromFile(stateFile string) (*State, error) {
	s := &State{
		path: stateFile,
		data: make(map[string]*voteRecord),
	}

	data, err := os.ReadFile(stateFile)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return s, nil
	}

	err = json.Unmarshal(data, &s.data)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func NewMemState() *State {
	return &State{}
}

func NewTmpState() *State {
	return &State{
		path: "/tmp/erc20rw-signer-state.json",
	}
}

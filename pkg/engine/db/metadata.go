package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/engine/types"
)

type metadata struct {
	Identifier string `json:"identifier"`
	Type       string `json:"type"`
	Content    []byte `json:"content"`
}

type metadataType string

const (
	metadataTypeTable     metadataType = "table"
	metadataTypeProcedure metadataType = "procedure"
	metadataTypeExtension metadataType = "extension"
)

const (
	metadataTableName            = "metadata"
	createMetadataTableStatement = `
		CREATE TABLE IF NOT EXISTS ` + metadataTableName + ` (
			identifier TEXT NOT NULL,
			type TEXT NOT NULL,
			content BLOB NOT NULL,
			PRIMARY KEY (identifier, type)
		) WITHOUT ROWID, STRICT;
	`

	insertMetadataStatement = `
		INSERT INTO metadata (identifier, type, content)
		VALUES ($identifier, $type, $content);
	`

	selectMetadataStatement = `
		SELECT identifier, content
		FROM metadata
		WHERE type = $type;
	`
)

func (d *DB) initMetadataTable(ctx context.Context) error {
	exists, err := d.Sqldb.TableExists(ctx, metadataTableName)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	return d.Sqldb.Execute(ctx, createMetadataTableStatement, nil)
}

func (d *DB) storeMetadata(ctx context.Context, meta *metadata) error {
	return d.Sqldb.Execute(ctx, insertMetadataStatement, map[string]interface{}{
		"$identifier": meta.Identifier,
		"$type":       meta.Type,
		"$content":    meta.Content,
	})
}

func (d *DB) getMetadata(ctx context.Context, metaType metadataType) ([]*metadata, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	cached, ok := d.metadataCache[metaType]
	if ok {
		return cached, nil
	}

	results, err := d.Sqldb.QueryUnsafe(ctx, selectMetadataStatement, map[string]interface{}{
		"$type": metaType,
	})
	if err != nil {
		return nil, err
	}

	var metas []*metadata
	for _, result := range results {
		identAny, ok := result["identifier"]
		if !ok {
			return nil, fmt.Errorf("stored metadata missing identifier")
		}
		ident, ok := identAny.(string)
		if !ok {
			return nil, fmt.Errorf("stored metadata identifier is not a string")
		}

		contentAny, ok := result["content"]
		if !ok {
			return nil, fmt.Errorf("stored metadata missing content")
		}

		content, ok := contentAny.([]byte)
		if !ok {
			return nil, fmt.Errorf("stored metadata content is not a byte array")
		}

		metas = append(metas, &metadata{
			Identifier: ident,
			Type:       string(metaType),
			Content:    content,
		})
	}

	d.metadataCache[metaType] = metas

	return metas, nil
}

// VersionedMetadata is a generic that wraps a serializable type with a version
type VersionedMetadata struct {
	Version int    `json:"version"`
	Data    []byte `json:"data"`
}

func (d *DB) getVersionedMetadata(ctx context.Context, metaType metadataType) ([]*VersionedMetadata, error) {
	metas, err := d.getMetadata(ctx, metaType)
	if err != nil {
		return nil, err
	}

	var versionedMetas []*VersionedMetadata
	for _, meta := range metas {
		versionedMeta := &VersionedMetadata{}
		err = json.Unmarshal(meta.Content, versionedMeta)
		if err != nil {
			return nil, err
		}

		versionedMetas = append(versionedMetas, versionedMeta)
	}

	return versionedMetas, nil
}

func (d *DB) persistVersionedMetadata(ctx context.Context, identifier string, metaType metadataType, meta *VersionedMetadata) error {
	bts, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	return d.storeMetadata(ctx, &metadata{
		Identifier: identifier,
		Type:       string(metaType),
		Content:    bts,
	})
}

// serializable is an interface and generic that all serializable types must implement
type serializable interface {
	types.Table | types.Procedure | types.Extension
}

func decodeMetadata[T serializable](meta []*VersionedMetadata) ([]*T, error) {
	var decoded []*T

	for _, value := range meta {
		tbl := new(T)

		err := json.Unmarshal(value.Data, tbl)
		if err != nil {
			return nil, err
		}

		decoded = append(decoded, tbl)
	}

	return decoded, nil
}

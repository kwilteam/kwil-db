package db

import (
	"context"
	"encoding/base64"
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
	metadataTypeTable      metadataType = "table"
	metadataTypeProcedure  metadataType = "procedure"
	metadataTypeExtensions metadataType = "extensions"
)

func getMetadataType(val any) (metadataType, error) {
	var metaType metadataType
	switch val.(type) {
	case *types.Table:
		metaType = metadataTypeTable
	case *types.Procedure:
		metaType = metadataTypeProcedure
	default:
		return "", fmt.Errorf("unknown metadata type: %T", val)
	}

	return metaType, nil
}

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
	exists, err := d.sqldb.TableExists(ctx, metadataTableName)
	if err != nil {
		return err
	}

	if exists {
		return nil
	}

	return d.sqldb.Execute(createMetadataTableStatement)
}

func (d *DB) storeMetadata(ctx context.Context, meta *metadata) error {
	return d.sqldb.Execute(insertMetadataStatement, map[string]interface{}{
		"$identifier": meta.Identifier,
		"$type":       meta.Type,
		"$content":    meta.Content,
	})
}

func (d *DB) getMetadata(ctx context.Context, metaType metadataType) ([]*metadata, error) {
	reader, err := d.sqldb.Query(ctx, selectMetadataStatement, map[string]interface{}{
		"$type": metaType,
	})
	if err != nil {
		return nil, err
	}

	if reader == nil {
		return nil, nil
	}

	results, err := ResultsfromReader(reader)
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

		// decode content as base64 string and convert to byte array
		contentStr, ok := contentAny.(string)
		if !ok {
			return nil, fmt.Errorf("stored metadata content is not a byte array")
		}

		content, err := base64.StdEncoding.DecodeString(contentStr)
		if err != nil {
			return nil, err
		}

		metas = append(metas, &metadata{
			Identifier: ident,
			Type:       string(metaType),
			Content:    content,
		})
	}

	return metas, nil
}

// versionedMetadata is a generic that wraps a serializable type with a version
type versionedMetadata struct {
	Version int `json:"version"`
	Data    any `json:"data"`
}

func (d *DB) getVersionedMetadata(ctx context.Context, metaType metadataType) ([]*versionedMetadata, error) {
	metas, err := d.getMetadata(ctx, metaType)
	if err != nil {
		return nil, err
	}

	var versionedMetas []*versionedMetadata
	for _, meta := range metas {
		versionedMeta := &versionedMetadata{}
		err = json.Unmarshal(meta.Content, versionedMeta)
		if err != nil {
			return nil, err
		}

		versionedMetas = append(versionedMetas, versionedMeta)
	}

	return versionedMetas, nil
}

func (d *DB) persistVersionedMetadata(ctx context.Context, meta *versionedMetadata) error {
	bts, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	metaType, err := getMetadataType(meta.Data)
	if err != nil {
		return err
	}

	ident, err := getIdentifier(meta.Data)
	if err != nil {
		return err
	}

	return d.storeMetadata(ctx, &metadata{
		Identifier: ident,
		Type:       string(metaType),
		Content:    bts,
	})
}

// getIdentifier returns the identifier for a serializable type.
// this is opposed to using a method on the structs, which requires an extra
// method in the types package
func getIdentifier(data any) (string, error) {
	// we have to use any instead of serializable because you can't type-assert a generic
	var ident string

	switch dataType := data.(type) {
	case *types.Table:
		ident = dataType.Name
	case *types.Procedure:
		ident = dataType.Name
	default:
		return "", fmt.Errorf("invalid serializable type: %s", data)
	}

	return ident, nil
}

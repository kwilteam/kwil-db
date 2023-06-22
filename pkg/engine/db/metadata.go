package db

import (
	"context"
	"encoding/base64"
	"fmt"
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

package metadata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrMetadataVersionMismatch = errors.New("metadata version mismatch. the database needs to run migrations")
)

type metadataType string

const (
	// metadataTypeGeneral is the type for general metadata.
	metadataTypeGeneral metadataType = "general"

	// metadataTypeTable is the type for table metadata.
	metadataTypeTable metadataType = "table"

	// metadataTypeProcedure is the type for procedure metadata.
	metadataTypeProcedure metadataType = "procedure"

	// metadataTypeExtension is the type for extension metadata.
	metadataTypeExtension metadataType = "extension"
)

func (t metadataType) version() int {
	switch t {
	case metadataTypeGeneral:
		return 0
	case metadataTypeTable:
		return 0
	case metadataTypeProcedure:
		return 0
	case metadataTypeExtension:
		return 0
	default:
		panic(fmt.Sprintf("unknown metadata type: %s", t))
	}
}

// metadata represents a metadata entry.
type metadata struct {
	Type    metadataType `json:"type"`
	Content []byte       `json:"content"`
}

// versionedMetadata represents a metadata entry with a version.
type versionedMetadata struct {
	Version  int       `json:"version"`
	Metadata *metadata `json:"metadata"`
}

// storeMetadata stores a metadata entry in the given connection.
func storeMetadata(ctx context.Context, kv KV, meta *metadata) error {
	bts, err := json.Marshal(&versionedMetadata{
		Version:  meta.Type.version(),
		Metadata: meta,
	})
	if err != nil {
		return err
	}

	return kv.Set(ctx, []byte(meta.Type), bts)
}

// getMetadata returns serialized metadata of a certain type.
// if the metadata is not the correct version, it will be return an error.
func getMetadata(ctx context.Context, kv KV, metaType metadataType) ([]byte, error) {
	bts, err := kv.Get(ctx, []byte(metaType))
	if err != nil {
		return nil, err
	}

	var meta versionedMetadata
	err = json.Unmarshal(bts, &meta)
	if err != nil {
		return nil, err
	}

	if meta.Version != metaType.version() {
		return nil, fmt.Errorf(`%w: expected version %d for metadata "%s", got %d`, ErrMetadataVersionMismatch, metaType.version(), metaType, meta.Version)
	}

	return meta.Metadata.Content, nil
}

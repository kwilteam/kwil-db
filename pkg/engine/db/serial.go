package db

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/kwilteam/kwil-db/pkg/engine/types"
)

/*
	this file handles serializing and deserializing metadata to and from the database

	to add a new data type, you need to add it to the serializable generic, create a predefined type,
	version, metadata type, and identifier getter in the getIdentifier function.

	These are are all right below this comment to make it easier to find, besides metadataType, which is in metadata.go
*/

// serializable is an interface and generic that all serializable types must implement
type serializable interface {
	types.Table | types.Procedure
}

// predefined types that are serializable
var (
	tableType     = reflect.TypeOf(types.Table{})
	procedureType = reflect.TypeOf(types.Procedure{})
)

// metadataTypesMap contains the metadataType enum for each serializable type
var metadataTypesMap = map[reflect.Type]metadataType{
	tableType:     metadataTypeTable,
	procedureType: metadataTypeProcedure,
}

// metadataVersionMap contains the metadata version for each serializable type
var metadataVersionMap = map[reflect.Type]int{
	tableType:     1,
	procedureType: 1,
}

// getIdentifier returns the identifier for a serializable type.
// this is opposed to using a method on the structs, which requires an extra
// method in the types package
func getIdentifier(data any) (string, error) {
	// we have to use any instead of serializable because you can't type-assert a generic
	var ident string

	refType := reflect.TypeOf(data).Elem()
	switch refType {
	case tableType:
		ident = data.(*types.Table).Name
	case procedureType:
		ident = data.(*types.Procedure).Name
	default:
		return "", fmt.Errorf("invalid serializable type: %s", refType.String())
	}

	return ident, nil
}

// serialized is a generic that wraps a serializable type with a version
type serialized[T serializable] struct {
	Version int `json:"version"`
	Data    *T  `json:"data"`
}

// serdes is a generic that wraps a database and provides methods for serializing and deserializing serializable types
type serdes[T serializable] struct {
	db *DB
}

// getReflectType returns the reflect.Type for a serializable type
func (d serdes[T]) getReflectType() (ref reflect.Type, err error) {
	defer func() {
		if r := recover(); r != nil {
			typ := reflect.TypeOf(new(T))
			err = fmt.Errorf("serializable types must be pointers.  this is an internal implementation bug.  received: %s", typ.String())

			return
		}
	}()

	return reflect.TypeOf(new(T)).Elem(), nil
}

// getStructType returns the metadataType for a serializable type
func (d serdes[T]) getStructType() (metadataType, error) {
	ref, err := d.getReflectType()
	if err != nil {
		return "", err
	}

	meta, ok := metadataTypesMap[ref]
	if !ok {
		return "", fmt.Errorf("invalid metadata type: %s", ref.String())
	}

	return meta, nil
}

// getMetadataVersion returns the metadata version for a serializable type
func (d serdes[T]) getMetadataVersion() (int, error) {
	ref, err := d.getReflectType()
	if err != nil {
		return 0, err
	}

	version, ok := metadataVersionMap[ref]
	if !ok {
		return 0, fmt.Errorf("invalid metadata type: %s", ref.String())
	}

	return version, nil
}

// listDeserialized returns a list of deserialized serializable metadata from the database
func (d serdes[T]) listDeserialized(ctx context.Context) ([]*T, error) {
	structType, err := d.getStructType()
	if err != nil {
		return nil, err
	}

	serializedData, err := d.db.getMetadata(ctx, structType)
	if err != nil {
		return nil, err
	}

	var results []*T

	for _, ser := range serializedData {
		var s T

		err = json.Unmarshal(ser.Content, &s)
		if err != nil {
			return nil, err
		}

		results = append(results, &s)
	}

	return results, nil
}

func unmarshalAny(data []byte, v any) error {
	err := json.Unmarshal(data, v)
	if err != nil {
		return err
	}

	return nil
}

// persistSerializable persists serializable metadata to the database
func (d serdes[T]) persistSerializable(ctx context.Context, data *T) error {
	structType, err := d.getStructType()
	if err != nil {
		return err
	}

	version, err := d.getMetadataVersion()
	if err != nil {
		return err
	}

	serializedData, err := json.Marshal(&serialized[T]{
		Version: version,
		Data:    data,
	})
	if err != nil {
		return err
	}

	identifier, err := getIdentifier(data)
	if err != nil {
		return err
	}

	return d.db.storeMetadata(ctx, &metadata{
		Identifier: identifier,
		Type:       string(structType),
		Content:    serializedData,
	})
}

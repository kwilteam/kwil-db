package db_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/cstockton/go-conv"
	"github.com/kwilteam/kwil-db/pkg/engine/db"
)

type storedMetadata struct {
	Identifier string `json:"identifier"`
	Type       string `json:"type"`
	Content    []byte `json:"content"`
}

func getIdentType(valArr ...map[string]any) (string, error) {
	if len(valArr) < 1 {
		return "", fmt.Errorf("expected exactly one argument")
	}
	if len(valArr) > 1 {
		panic("cannot handle multiple arguments")
	}

	vals := valArr[0]

	typ, ok := vals["$type"]
	if !ok {
		return "", fmt.Errorf("metadata missing $type")
	}

	typStr, err := conv.String(typ)
	if err != nil {
		panic(fmt.Sprintf("type is not a string: %v", typ))
	}

	return typStr, nil
}

func getMetadata(valArr ...map[string]any) (storedMetadata, error) {
	if len(valArr) < 1 {
		return storedMetadata{}, fmt.Errorf("expected exactly one argument")
	}
	if len(valArr) > 1 {
		panic("cannot handle multiple arguments")
	}

	vals := valArr[0]

	ident, ok := vals["$identifier"]
	if !ok {
		return storedMetadata{}, fmt.Errorf("metadata missing $identifier")
	}

	identStr, ok := ident.(string)
	if !ok {
		panic(fmt.Sprintf("identifier is not a string: %v", ident))
	}

	typ, ok := vals["$type"]
	if !ok {
		return storedMetadata{}, fmt.Errorf("metadata missing $type")
	}

	typStr, err := conv.String(typ)
	if err != nil {
		panic(fmt.Sprintf("type is not a string: %v", typ))
	}

	content, ok := vals["$content"]
	if !ok {
		return storedMetadata{}, fmt.Errorf("metadata missing $content")
	}

	contentBts, ok := content.([]byte)
	if !ok {
		panic(fmt.Sprintf("content is not a []byte: %v", content))
	}

	return storedMetadata{
		Identifier: identStr,
		Type:       typStr,
		Content:    contentBts,
	}, nil
}

type mockDB struct {
	metadata []storedMetadata
}

func newMockDB() *mockDB {
	return &mockDB{
		metadata: []storedMetadata{},
	}
}

func (m *mockDB) Close() error {
	return nil
}

func (m *mockDB) Delete() error {
	return nil
}

func (m *mockDB) Execute(stmt string, args ...map[string]any) error {
	storeMeta, err := getMetadata(args...)
	if err != nil {
		// if an error is returned, it means that the query is not a metadata query, and we are only testing metadata queries
		return nil
	}

	m.metadata = append(m.metadata, storeMeta)

	return nil
}

func (m *mockDB) Prepare(stmt string) (db.Statement, error) {
	return &mockStatement{}, nil
}

func (m *mockDB) Query(ctx context.Context, query string, args ...map[string]any) (io.Reader, error) {
	identType, err := getIdentType(args...)
	if err != nil {
		// if an error is returned, it means that the query is not a metadata query, and we are only testing metadata queries
		return nil, nil
	}

	var metas []storedMetadata
	for _, meta := range m.metadata {
		if meta.Type == identType {
			metas = append(metas, meta)
		}
	}

	bts, err := json.Marshal(metas)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(bts), nil
}

func (m *mockDB) Savepoint() (db.Savepoint, error) {
	return &mockSavepoint{}, nil
}

func (m *mockDB) TableExists(ctx context.Context, table string) (bool, error) {
	return true, nil
}

type mockStatement struct{}

func (m *mockStatement) Close() error {
	return nil
}

func (m *mockStatement) Execute(args map[string]any) (io.Reader, error) {
	return nil, nil
}

type mockSavepoint struct{}

func (m *mockSavepoint) Commit() error {
	return nil
}

func (m *mockSavepoint) Rollback() error {
	return nil
}

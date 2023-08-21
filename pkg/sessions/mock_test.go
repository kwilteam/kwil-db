package sessions_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/kwilteam/kwil-db/pkg/sessions"
)

func newMockWal(startingReconds ...sessions.WalRecord) *mockWal {
	data := [][]byte{}

	for _, record := range startingReconds {
		bts, err := sessions.SerializeWalRecord(&record)
		if err != nil {
			panic(fmt.Errorf("error serializing wal record in mock wal: %w", err))
		}

		data = append(data, bts)
	}

	return &mockWal{
		index: 0,
		data:  data,
	}
}

type mockWal struct {
	index int
	data  [][]byte
}

func (m *mockWal) Append(ctx context.Context, data []byte) error {
	m.data = append(m.data, data)
	return nil
}

func (m *mockWal) ReadNext(ctx context.Context) ([]byte, error) {
	if m.index >= len(m.data) {
		return nil, io.EOF
	}

	data := m.data[m.index]
	m.index++

	return data, nil
}

func (m *mockWal) Truncate(ctx context.Context) error {
	m.data = [][]byte{}
	m.index = 0
	return nil
}

func (m *mockWal) isEmpty() bool {
	return len(m.data) == 0
}

type keyValue struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

func (k *keyValue) serialize() []byte {
	bts, err := json.Marshal(k)
	if err != nil {
		panic(fmt.Errorf("error serializing key value: %w", err))
	}

	return bts
}

func newMockCommittable(commitId string, data map[string]any) *mockCommittable {
	return &mockCommittable{
		isInCommit:   false,
		isInApply:    false,
		commitId:     []byte(commitId),
		dataToCommit: data,
		appliedData:  map[string]any{},
	}
}

func mockCommittable1() *mockCommittable {
	return newMockCommittable("commit1", map[string]any{
		"key1": "value1",
		"key2": "value2",
	})
}

func mockCommittable2() *mockCommittable {
	return newMockCommittable("commit2", map[string]any{
		"key3": "value3",
		"key4": "value4",
	})
}

type mockCommittable struct {
	isInCommit bool
	isInApply  bool

	canceled bool

	commitId []byte

	dataToCommit map[string]any
	appliedData  map[string]any
}

func (m *mockCommittable) BeginCommit(ctx context.Context) error {
	if m.isInCommit {
		return fmt.Errorf("already in commit")
	}
	if m.isInApply {
		return fmt.Errorf("cannot begin commit while in apply")
	}

	m.isInCommit = true

	return nil
}

func (m *mockCommittable) EndCommit(ctx context.Context, appender func([]byte) error) (err error) {
	if !m.isInCommit {
		return fmt.Errorf("not in commit")
	}
	if m.isInApply {
		return fmt.Errorf("cannot end commit while in apply")
	}

	m.isInCommit = false

	for key, value := range m.dataToCommit {
		bts, err := json.Marshal(keyValue{
			Key:   key,
			Value: value,
		})
		if err != nil {
			return err
		}

		err = appender(bts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *mockCommittable) BeginApply(ctx context.Context) error {
	if m.isInCommit {
		return fmt.Errorf("cannot begin apply while in commit")
	}
	if m.isInApply {
		return fmt.Errorf("already in apply")
	}

	m.isInApply = true

	return nil
}

func (m *mockCommittable) Apply(ctx context.Context, changes []byte) error {
	if !m.isInApply {
		return fmt.Errorf("not in apply")
	}

	var kv keyValue
	err := json.Unmarshal(changes, &kv)
	if err != nil {
		return err
	}

	m.appliedData[kv.Key] = kv.Value

	return nil
}

func (m *mockCommittable) EndApply(ctx context.Context) error {
	if !m.isInApply {
		return fmt.Errorf("not in apply")
	}
	if m.isInCommit {
		return fmt.Errorf("cannot end apply while in commit")
	}

	m.isInApply = false

	return nil
}

func (m *mockCommittable) Cancel(ctx context.Context) {
	m.isInCommit = false
	m.isInApply = false

	m.appliedData = map[string]any{}

	m.canceled = true
}

func (m *mockCommittable) ID(ctx context.Context) ([]byte, error) {
	return m.commitId, nil
}

type mockCommittableWithErrors struct {
	*mockCommittable

	errInBeginCommit bool
	errInEndCommit   bool
	errInBeginApply  bool
	errInApply       bool
	errInEndApply    bool
}

func (m *mockCommittableWithErrors) BeginCommit(ctx context.Context) error {
	err := m.mockCommittable.BeginCommit(ctx)
	if err != nil {
		return err
	}

	if m.errInBeginCommit {
		return fmt.Errorf("error in BeginCommit")
	}

	return nil
}

func (m *mockCommittableWithErrors) EndCommit(ctx context.Context, appender func([]byte) error) (err error) {
	err = m.mockCommittable.EndCommit(ctx, appender)
	if err != nil {
		return err
	}

	if m.errInEndCommit {
		return fmt.Errorf("error in EndCommit")
	}

	return nil
}

func (m *mockCommittableWithErrors) BeginApply(ctx context.Context) error {
	err := m.mockCommittable.BeginApply(ctx)
	if err != nil {
		return err
	}

	if m.errInBeginApply {
		return fmt.Errorf("error in BeginApply")
	}

	return nil
}

func (m *mockCommittableWithErrors) Apply(ctx context.Context, changes []byte) error {
	err := m.mockCommittable.Apply(ctx, changes)
	if err != nil {
		return err
	}

	if m.errInApply {
		return fmt.Errorf("error in Apply")
	}

	return nil
}

func (m *mockCommittableWithErrors) EndApply(ctx context.Context) error {
	err := m.mockCommittable.EndApply(ctx)
	if err != nil {
		return err
	}

	if m.errInEndApply {
		return fmt.Errorf("error in EndApply")
	}

	return nil
}

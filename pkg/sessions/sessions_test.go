package sessions_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"sort"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/sessions"
	"github.com/stretchr/testify/assert"
)

// TODO: test calling things like begin, end, apply, etc. out of order / multiple times
// TODO: test register and unregister; and try calling them at different times
// e.g. calling register after begin, or unregister after end, etc.

func Test_Session(t *testing.T) {
	type fields struct {
		wal          sessions.Wal
		committables map[string]sessions.Committable
	}

	type testCase struct {
		name   string
		fields fields
		err    error
	}

	tests := []testCase{
		{
			name: "basic atomic commit",
			fields: fields{
				wal: newMockWal(),
				committables: map[string]sessions.Committable{
					"c1": mockCommittable1(),
					"c2": mockCommittable2(),
				},
			},
		},
		{
			name: "commit id order",
			fields: fields{
				wal: newMockWal(),
				committables: map[string]sessions.Committable{
					"c2": mockCommittable2(),
					"c1": mockCommittable1(),
				},
			},
		},
		{
			name: "error in BeginCommit",
			fields: fields{
				wal: newMockWal(),
				committables: map[string]sessions.Committable{
					"c1": mockCommittable1(),
					"c2": &mockCommittableWithErrors{
						mockCommittable:  mockCommittable2(),
						errInBeginCommit: true,
					},
				},
			},
			err: sessions.ErrBeginCommit,
		},
		{
			name: "error in EndCommit",
			fields: fields{
				wal: newMockWal(),
				committables: map[string]sessions.Committable{
					"c1": mockCommittable1(),
					"c2": &mockCommittableWithErrors{
						mockCommittable: mockCommittable2(),
						errInEndCommit:  true,
					},
				},
			},
			err: sessions.ErrEndCommit,
		},
		{
			name: "error in BeginApply",
			fields: fields{
				wal: newMockWal(),
				committables: map[string]sessions.Committable{
					"c1": mockCommittable1(),
					"c2": &mockCommittableWithErrors{
						mockCommittable: mockCommittable2(),
						errInBeginApply: true,
					},
				},
			},
			err: sessions.ErrBeginApply,
		},
		{
			name: "error in Apply",
			fields: fields{
				wal: newMockWal(),
				committables: map[string]sessions.Committable{
					"c1": mockCommittable1(),
					"c2": &mockCommittableWithErrors{
						mockCommittable: mockCommittable2(),
						errInApply:      true,
					},
				},
			},
			err: sessions.ErrApply,
		},
		{
			name: "error in EndApply",
			fields: fields{
				wal: newMockWal(),
				committables: map[string]sessions.Committable{
					"c1": mockCommittable1(),
					"c2": &mockCommittableWithErrors{
						mockCommittable: mockCommittable2(),
						errInEndApply:   true,
					},
				},
			},
			err: sessions.ErrEndApply,
		},
		// TODO: test wal erroring out
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var id []byte
			outErr := func() error {
				ctx := context.Background()
				committer := sessions.NewAtomicCommitter(ctx, tt.fields.wal)

				for id, committable := range tt.fields.committables {
					err := committer.Register(ctx, id, committable)
					if err != nil {
						return err
					}
				}

				defer committer.Close()
				err := committer.ClearWal(ctx)
				if err != nil {
					return err
				}

				if !tt.fields.wal.(*mockWal).isEmpty() {
					return fmt.Errorf("expected wal to be empty after startup")
				}

				err = committer.Begin(ctx)
				if err != nil {
					return err
				}

				id, err = committer.ID(ctx)
				if err != nil {
					return err
				}

				applyErr := make(chan error)
				err = committer.Commit(ctx, func(err error) {
					applyErr <- err
				})
				if err != nil {
					return err
				}

				err = <-applyErr
				if err != nil {
					return err
				}

				return nil
			}()
			assertError(t, outErr, tt.err)
			if tt.err != nil {
				assertAllCanceled(t, tt.fields.committables)
			}
			if outErr != nil {
				return
			}

			for _, committable := range tt.fields.committables {
				for key, value := range committable.(*mockCommittable).appliedData {
					if value != committable.(*mockCommittable).dataToCommit[key] {
						t.Fatalf("expected value %v, got %v", committable.(*mockCommittable).dataToCommit[key], value)
					}
				}
			}

			orderedCommittables := orderAlphabetically(tt.fields.committables)
			expectedHash := sha256.New()

			for _, committable := range orderedCommittables {
				_, err := expectedHash.Write(committable.(*mockCommittable).commitId)
				if err != nil {
					t.Fatal(err)
				}
			}

			expectedId := expectedHash.Sum(nil)

			assert.Equal(t, expectedId, id, fmt.Sprintf("expected id %v, got %v", expectedId, id))

		})
	}
}

// orderAlphabetically orders the committables alphabetically by their unique identifier.
func orderAlphabetically(committables map[string]sessions.Committable) []sessions.Committable {
	// Extracting keys
	keys := make([]string, 0, len(committables))
	for k := range committables {
		keys = append(keys, k)
	}

	// Sorting keys
	sort.Strings(keys)

	// Creating sorted slice
	sorted := make([]sessions.Committable, 0, len(committables))
	for _, key := range keys {
		sorted = append(sorted, committables[key])
	}

	return sorted
}

var (
	walRecordBegin = sessions.WalRecord{
		Type: sessions.WalRecordTypeBegin,
	}
	walRecordCommit = sessions.WalRecord{
		Type: sessions.WalRecordTypeCommit,
	}
	walRecordCs1 = sessions.WalRecord{
		Type:          sessions.WalRecordTypeChangeset,
		CommittableId: sessions.CommittableId("c1"),
		Data: (&keyValue{
			Key:   "key1",
			Value: "c1_changeset_1",
		}).serialize(),
	}
	walRecordCs2 = sessions.WalRecord{
		Type:          sessions.WalRecordTypeChangeset,
		CommittableId: sessions.CommittableId("c2"),
		Data: (&keyValue{
			Key:   "key1",
			Value: "c2_changeset_1",
		}).serialize(),
	}
)

// tests for when data already exists in the wal
func Test_ExistingWal(t *testing.T) {
	type fields struct {
		wal sessions.Wal
	}

	type committableData struct {
		id          string
		committable sessions.Committable
		resultData  map[string]any
	}

	type testCase struct {
		name           string
		fields         fields
		commitableData []committableData
		err            error
	}

	tests := []testCase{
		{
			name: "starting with uncommitted records in wal, should not apply them",
			fields: fields{
				wal: newMockWal([]sessions.WalRecord{
					walRecordBegin,
					walRecordCs1,
					walRecordCs2,
				}...),
			},
			commitableData: []committableData{
				{
					id:          "c1",
					committable: mockCommittable1(),
					resultData:  map[string]any{},
				},
				{
					id:          "c2",
					committable: mockCommittable2(),
					resultData:  map[string]any{},
				},
			},
		},
		{
			name: "starting with committed records in wal, should apply them",
			fields: fields{
				wal: newMockWal([]sessions.WalRecord{
					walRecordBegin,
					walRecordCs1,
					walRecordCs2,
					walRecordCommit,
				}...),
			},
			commitableData: []committableData{
				{
					id:          "c1",
					committable: mockCommittable1(),
					resultData: map[string]any{
						"key1": "c1_changeset_1",
					},
				},
				{
					id:          "c2",
					committable: mockCommittable2(),
					resultData: map[string]any{
						"key1": "c2_changeset_1",
					},
				},
			},
		},
		{
			name: "starting with records in wal, but no begin, should truncate",
			fields: fields{
				wal: newMockWal([]sessions.WalRecord{
					walRecordCs1,
					walRecordCs2,
					walRecordCommit,
				}...),
			},
			commitableData: []committableData{
				{
					id:          "c1",
					committable: mockCommittable1(),
					resultData:  map[string]any{},
				},
				{
					id:          "c2",
					committable: mockCommittable2(),
					resultData:  map[string]any{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			commitableMap := map[string]sessions.Committable{}
			for _, data := range tt.commitableData {
				commitableMap[data.id] = data.committable
			}

			ctx := context.Background()

			committer := sessions.NewAtomicCommitter(ctx, tt.fields.wal)
			for id, committable := range commitableMap {
				err := committer.Register(ctx, id, committable)
				if err != nil {
					t.Fatal(err)
				}
			}
			err := committer.ClearWal(ctx)
			assertError(t, err, tt.err)
			if tt.err != nil {
				assertAllCanceled(t, commitableMap)
			}
			if err != nil {
				return
			}

			if !tt.fields.wal.(*mockWal).isEmpty() {
				t.Fatalf("expected wal to be empty after startup")
			}

			for _, data := range tt.commitableData {
				for key, value := range data.committable.(*mockCommittable).appliedData {
					if value != data.resultData[key] {
						t.Fatalf("expected value %v, got %v", data.resultData[key], value)
					}
				}
			}
		})
	}
}

func assertError(t *testing.T, returned, expected error) {
	if expected == nil {
		if returned != nil {
			t.Fatalf("expected no error, got %v", returned)
		}

		return
	}

	if returned == nil {
		t.Fatalf("expected error %v, got nil", expected)
	}

	if !errors.Is(returned, expected) {
		t.Fatalf("expected error %v, got %v", expected, returned)
	}
}

func assertAllCanceled(t *testing.T, committables map[string]sessions.Committable) {
	for _, committable := range committables {
		canceled := false
		switch v := committable.(type) {
		case *mockCommittable:
			canceled = v.canceled
		case *mockCommittableWithErrors:
			canceled = v.canceled
		}

		if !canceled {
			t.Fatalf("expected committable to be canceled")
		}
	}
}

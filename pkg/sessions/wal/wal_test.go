package wal_test

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/kwilteam/kwil-db/pkg/sessions/wal"
	"github.com/stretchr/testify/assert"
)

const directory = "./tmp"
const path = directory + "/wal"

func Test_Wal(t *testing.T) {
	type testCase struct {
		name     string
		testFunc func(t *testing.T, w *wal.Wal)
	}

	testCases := []testCase{
		{
			name: "intended use of Wal",
			testFunc: func(t *testing.T, w *wal.Wal) {
				ctx := context.Background()
				err := w.Append(ctx, []byte("hello"))
				assert.NoError(t, err)

				err = w.Append(ctx, []byte("world"))
				assert.NoError(t, err)

				hello, err := w.ReadNext(ctx)
				assert.NoError(t, err)
				assert.Equal(t, []byte("hello"), hello)

				world, err := w.ReadNext(ctx)
				assert.NoError(t, err)
				assert.Equal(t, []byte("world"), world)

				_, err = w.ReadNext(ctx)
				assert.ErrorIs(t, err, io.EOF)

				err = w.Truncate(ctx)
				assert.NoError(t, err)
			},
		},
		{
			name: "test total truncation",
			// write two records, truncate, then check that the file is empty
			// test that truncating new wal does nothing
			testFunc: func(t *testing.T, w *wal.Wal) {
				ctx := context.Background()
				err := w.Append(ctx, []byte("hello"))
				assert.NoError(t, err)

				err = w.Append(ctx, []byte("world"))
				assert.NoError(t, err)

				hello, err := w.ReadNext(ctx)
				assert.NoError(t, err)
				assert.Equal(t, []byte("hello"), hello)

				err = w.Truncate(ctx)
				assert.NoError(t, err)

				_, err = w.ReadNext(ctx)
				assert.ErrorIs(t, err, io.EOF)

				err = w.Truncate(ctx)
				assert.NoError(t, err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w, err := wal.OpenWal(path)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				err := errors.Join(
					w.Close(),
					os.RemoveAll(directory),
				)
				if err != nil {
					t.Fatal(err)
				}
			}()

			tc.testFunc(t, w)
		})
	}
}

package pg

// This file provides the binary encoding and decoding functions for
// sql.Statistics groups. All types used within the Statistic and
// ColumnStatistics fields must be registered with the encoding/gob package.

import (
	"encoding/gob"
	"errors"
	"io"
	"reflect"

	"github.com/kwilteam/kwil-db/common/sql"
)

// EncStats encodes a set of table statistics to an io.Writer.
func EncStats(w io.Writer, statsGroup map[sql.TableRef]*sql.Statistics) error {
	enc := gob.NewEncoder(w)
	// TableRef1,Statistics1,TableRef2,Statistics2,...
	for tblRef, stats := range statsGroup {
		err := enc.Encode(tblRef)
		if err != nil {
			return err
		}
		err = enc.Encode(stats)
		if err != nil {
			return err
		}
	}
	return nil
}

// DecStats decodes a set of table statistics from an io.Reader. Any type used
// must be registered with the gob package.
func DecStats(r io.Reader) (map[sql.TableRef]*sql.Statistics, error) {
	dec := gob.NewDecoder(r)

	// Read until EOF
	out := map[sql.TableRef]*sql.Statistics{}
	for {
		var tblRef sql.TableRef
		if err := dec.Decode(&tblRef); err != nil {
			if errors.Is(err, io.EOF) {
				break // return out, nil
			}
			return nil, err
		}

		stats := new(sql.Statistics)
		if err := dec.Decode(&stats); err != nil {
			return nil, err
		}

		if stats.RowCount > 0 {
			// gob will leave empty slice as nil, which is wrong for most fields
			for i := range stats.ColumnStatistics {
				cs := &stats.ColumnStatistics[i]
				cs.Min = nilToEmptySlice(cs.Min)
				cs.Max = nilToEmptySlice(cs.Max)
				for i := range cs.MCVals {
					cs.MCVals[i] = nilToEmptySlice(cs.MCVals[i])
				}
			}
		}

		out[tblRef] = stats
	}

	return out, nil
}

func nilToEmptySlice(s any) any {
	rt := reflect.TypeOf(s)
	if rt == nil {
		return nil
	}
	// Any custom types will decode as pointer, so recognize and dereference any
	// that we want as a value.
	// switch st := s.(type) {
	// case *types.UUID:
	// 	return *st
	// }
	if rt.Kind() == reflect.Slice &&
		reflect.ValueOf(s).IsNil() {
		st := reflect.SliceOf(rt.Elem())
		return reflect.MakeSlice(st, 0, 0).Interface()
	}
	return s
}

package datasource

import (
	"context"
	"strconv"

	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

// ScanData read the data source, return selected columns.
func ScanData(ctx context.Context, dsSchema *datatypes.Schema, dsRecords []Row, projection []string) *Result {
	if len(projection) == 0 {
		return ResultFromRaw(dsSchema, dsRecords)
	}

	// panic if projection field not found
	//for _, name := range projection {
	//	found := false
	//	for _, field := range ds.schema.Fields {
	//		if field.Name == name {
	//			found = true
	//			break
	//		}
	//	}
	//	if !found {
	//		panic(fmt.Sprintf("projection field %s not found", name))
	//	}
	//}

	fieldIndex := dsSchema.MapProjection(projection)
	newSchema := dsSchema.Project(projection...)

	// adjust column order based on projection
	projectedRecords := make([]Row, 0, len(dsRecords))
	for _, _row := range dsRecords {
		newRow := make(Row, len(projection))
		for j, idx := range fieldIndex {
			newRow[j] = _row[idx]
		}
		projectedRecords = append(projectedRecords, newRow)
	}

	out := StreamTap(ctx, projectedRecords)
	return ResultFromStream(newSchema, out)
}

// colTypeCast try to cast the raw column string to int, if failed, return the raw string.
// NOTE: we only support int/string for simplicity.
func colTypeCast(raw string) (kind string, value any) {
	v, err := strconv.ParseInt(raw, 10, 64)
	if err == nil {
		return "int64", v
	} else {
		return "string", raw
	}
}

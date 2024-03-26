package source

import (
	"strconv"

	"github.com/kwilteam/kwil-db/internal/engine/cost/datasource"
	"github.com/kwilteam/kwil-db/internal/engine/cost/datatypes"
)

// dsScan read the data source, return selected columns.
func dsScan(dsSchema *datatypes.Schema, dsRecords []datasource.Row, projection []string) *datasource.Result {
	if len(projection) == 0 {
		return datasource.ResultFromRaw(dsSchema, dsRecords)
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
	//newFields := make([]datatypes.Field, len(projection))
	//for i, idx := range fieldIndex {
	//	newFields[i] = dsSchema.Fields[idx]
	//}
	//newSchema := datatypes.NewSchema(newFields...)

	out := make(datasource.RowPipeline)
	go func() {
		defer close(out)

		for _, _row := range dsRecords {
			newRow := make(datasource.Row, len(projection))
			for j, idx := range fieldIndex {
				newRow[j] = _row[idx]
			}
			out <- newRow
		}
	}()

	return datasource.ResultFromStream(newSchema, out)
}

// colTypeCast try to cast the raw column string to int, if failed, return the raw string.
// NOTE: we only support int/string for simplicity.
func colTypeCast(raw string) (kind string, value any) {
	v, err := strconv.Atoi(raw)
	if err == nil {
		return "int", v
	} else {
		return "string", raw
	}
}

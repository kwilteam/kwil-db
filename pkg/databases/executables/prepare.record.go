package executables

import (
	"fmt"

	"github.com/doug-martin/goqu/v9"
)

type record map[string]interface{}

func (r *record) asGoqu() goqu.Record {
	return goqu.Record(*r)
}

// getRecords converts the user inputs and params into a goqu.Record.
// this is used for values that are being inserted or updated
func (p *preparer) getRecords() (record, error) {
	record := make(record)
	for _, param := range p.executable.Query.Params {
		val, err := p.prepareInput(param)
		if err != nil {
			return nil, fmt.Errorf(`failed to prepare input "%s": %w`, param.Name, err)
		}

		record[param.Column] = val.Value()
	}

	return record, nil
}

package database

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/olekukonko/tablewriter"

	"github.com/kwilteam/kwil-db/core/types"
)

// NOTE: I feel those types are better to be defined in the core/client package
// but also not sure, because how to display the response is a cli thing
//
// A possible way to do this is to define actual response types in core/client package
// and wrap them in cli package?

// respDBList represent databases belong to an owner in cli
type respDBList struct {
	Info []*types.DatasetIdentifier
	// owner is the owner configured in cli
	owner []byte
}

func (d *respDBList) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Info)
}

func (d *respDBList) MarshalText() ([]byte, error) {
	if len(d.Info) == 0 {
		return []byte(fmt.Sprintf("No databases found for '%x'.", d.owner)), nil
	}

	var msg bytes.Buffer
	if len(d.owner) == 0 {
		msg.WriteString("Databases:\n")
	} else {
		msg.WriteString(fmt.Sprintf("Databases belonging to '%x':\n", d.owner))
	}
	for i, db := range d.Info {
		msg.WriteString(fmt.Sprintf("  DBID: %s\n", db.DBID))
		msg.WriteString(fmt.Sprintf("    Name: %s\n", db.Name))
		msg.WriteString(fmt.Sprintf("    Owner: %x", db.Owner))
		if i != len(d.Info)-1 {
			msg.WriteString("\n")
		}
	}

	return msg.Bytes(), nil
}

// respRelations is a slice of maps that represent the relations(from set theory)
// of a database in cli
type respRelations struct {
	// to avoid recursive call of MarshalJSON
	Data *types.QueryResult
}

func (r *respRelations) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Data)
}

func (r *respRelations) MarshalText() ([]byte, error) {
	return recordsToTable(r.Data.ExportToStringMap()), nil
}

// recordsToTable converts records to a formatted table structure
// that can be printed
func recordsToTable(data []map[string]string) []byte {
	if len(data) == 0 {
		return []byte("No data to display.")
	}

	// collect headers
	headers := make([]string, 0, len(data[0]))
	for k := range data[0] {
		headers = append(headers, k)
	}

	// keep the headers in a sorted order
	sort.Strings(headers)

	var buf bytes.Buffer
	table := tablewriter.NewWriter(&buf)
	table.SetHeader(headers)
	table.SetAutoFormatHeaders(false)
	table.SetBorders(
		tablewriter.Border{Left: true, Top: false, Right: true, Bottom: false})

	for _, row := range data {
		rs := make([]string, 0, len(headers))
		for _, h := range headers {
			rs = append(rs, row[h])
		}
		table.Append(rs)
	}

	table.Render()
	return buf.Bytes()
}

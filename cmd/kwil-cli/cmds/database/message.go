package database

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/olekukonko/tablewriter"

	clientType "github.com/kwilteam/kwil-db/core/client/types"
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
	Data *clientType.Records
}

func (r *respRelations) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Data.Export())
}

func (r *respRelations) MarshalText() ([]byte, error) {
	return recordsToTable(r.Data), nil
}

// recordsToTable converts records to a formatted table structure
// that can be printed
func recordsToTable(r *clientType.Records) []byte {
	data := r.ExportString()

	if len(data) == 0 {
		return []byte("No data to display.")
	}

	// collect headers
	headers := make([]string, 0)
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
		rs := make([]string, 0)
		for _, h := range headers {
			rs = append(rs, row[h])
		}
		table.Append(rs)
	}

	table.Render()
	return buf.Bytes()
}

// respSchema is used to represent a database schema in cli
type respSchema struct {
	Schema *types.Schema
}

func (s *respSchema) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Schema)
}

func (s *respSchema) MarshalText() ([]byte, error) {
	// TODO: make output more readable
	var msg bytes.Buffer

	// now we print the metadata
	msg.WriteString("Tables:\n")
	for _, t := range s.Schema.Tables {
		msg.WriteString(fmt.Sprintf("  %s\n", t.Name))
		msg.WriteString("    Columns:\n")
		for _, c := range t.Columns {
			msg.WriteString(fmt.Sprintf("    %s\n", c.Name))
			msg.WriteString(fmt.Sprintf("      Type: %s\n", c.Type.String()))

			for _, a := range c.Attributes {
				msg.WriteString(fmt.Sprintf("      %s\n", a.Type))
				if a.Value != "" {
					msg.WriteString(fmt.Sprintf("        %s\n", fmt.Sprint(a.Value)))
				}
			}
		}
	}

	// print queries
	msg.WriteString("Actions:\n")
	for _, q := range s.Schema.Actions {
		public := "private"
		if q.Public {
			public = "public"
		}

		msg.WriteString(fmt.Sprintf("  %s (%s)\n", q.Name, public))
		msg.WriteString(fmt.Sprintf("    Inputs: %s\n", q.Parameters))
	}

	// print procedures
	msg.WriteString("Procedures:\n")
	for _, p := range s.Schema.Procedures {
		public := "private"
		if p.Public {
			public = "public"
		}

		msg.WriteString(fmt.Sprintf("  %s (%s)\n", p.Name, public))
		for _, param := range p.Parameters {
			msg.WriteString(fmt.Sprintf("    %s: %s\n", param.Name, param.Type.String()))
		}
	}

	return msg.Bytes(), nil
}

package database

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/kwilteam/kwil-db/core/client"
	"github.com/kwilteam/kwil-db/core/types/transactions"
	"github.com/kwilteam/kwil-db/core/utils"

	"github.com/olekukonko/tablewriter"
)

// NOTE: I feel those types are better to be defined in the client package
// but also not sure, because how to display the response is a cli thing
//
// A possible way to do this is define actual response types in client package
// and wrap them in cli package?

// respDBList represent databases belong to an owner in cli
type respDBList struct {
	Databases []string `json:"databases"`
	Owner     []byte   `json:"owner"`
}

type dbInfo struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

func (d *respDBList) MarshalJSON() ([]byte, error) {
	dbs := make([]dbInfo, len(d.Databases))

	for i, db := range d.Databases {
		dbs[i] = dbInfo{
			Name: db,
			Id:   utils.GenerateDBID(db, d.Owner),
		}
	}

	return json.Marshal(struct {
		Databases []dbInfo `json:"databases"`
		Owner     string   `json:"owner"`
	}{
		Databases: dbs,
		Owner:     fmt.Sprintf("%x", d.Owner),
	})
}

func (d *respDBList) MarshalText() ([]byte, error) {
	if len(d.Databases) == 0 {
		return []byte(fmt.Sprintf("No databases found for '%x'.", d.Owner)), nil
	}

	msg := fmt.Sprintf("Databases belonging to '%x':\n", d.Owner)
	for _, db := range d.Databases {
		msg += fmt.Sprintf(" - %s   (dbid:%s)\n", db, utils.GenerateDBID(db, d.Owner))
	}

	return []byte(msg), nil
}

// respRelations is a slice of maps that represent the relations(from set theory)
// of a database in cli
type respRelations struct {
	// to avoid recursive call of MarshalJSON
	Data *client.Records
}

func (r *respRelations) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Data.ExportString())
}

func (r *respRelations) MarshalText() ([]byte, error) {
	data := r.Data.ExportString()

	if len(data) == 0 {
		return []byte("No data to display."), nil
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
	return buf.Bytes(), nil
}

// respSchema is used to represent a database schema in cli
type respSchema struct {
	Schema *transactions.Schema
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
			msg.WriteString(fmt.Sprintf("      Type: %s\n", c.Type))

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
		msg.WriteString(fmt.Sprintf("  %s\n", q.Name))
		msg.WriteString(fmt.Sprintf("    Inputs: %s\n", q.Inputs))
	}

	return msg.Bytes(), nil
}

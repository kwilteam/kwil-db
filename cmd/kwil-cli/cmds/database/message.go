package database

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/kwilteam/kwil-db/pkg/client"
	"github.com/kwilteam/kwil-db/pkg/engine/utils"
	"github.com/kwilteam/kwil-db/pkg/transactions"
	"github.com/olekukonko/tablewriter"
)

// NOTE: I feel those types are better to be defined in the client package
// but also not sure, because how to display the response is a cli thing
//
// A possible way to do this is define actual response types in client package
// and wrap them in cli package?

// respTxHash is used to represent a transaction hash in cli
// NOTE: it's different from transactions.TxHash, this is for display purpose.
type respTxHash []byte

func (h respTxHash) Hex() string {
	return strings.ToUpper(fmt.Sprintf("%x", h))
}

func (h respTxHash) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		TxHash string `json:"tx_hash"`
	}{
		TxHash: h.Hex(),
	})
}

func (h respTxHash) MarshalText() (string, error) {
	return fmt.Sprintf("TxHash: %s", h.Hex()), nil
}

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

func (d *respDBList) MarshalText() (string, error) {
	if len(d.Databases) == 0 {
		return fmt.Sprintf("No databases found for '%x'.", d.Owner), nil
	}

	msg := fmt.Sprintf("Databases belonging to '%x':\n", d.Owner)
	for _, db := range d.Databases {
		msg += fmt.Sprintf(" - %s   (dbid:%s)\n", db, utils.GenerateDBID(db, d.Owner))
	}

	return msg, nil
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

func (r *respRelations) MarshalText() (string, error) {
	data := r.Data.ExportString()

	if len(data) == 0 {
		return "No data to display.", nil
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

	// TODO: fix the order
	for _, row := range data {
		rs := make([]string, 0)
		for _, h := range headers {
			rs = append(rs, row[h])
		}
		table.Append(rs)
	}

	table.Render()

	bs, err := io.ReadAll(&buf)
	if err != nil {
		return "", err
	}

	return string(bs), nil
}

// respSchema is used to represent a database schema in cli
type respSchema struct {
	Schema *transactions.Schema
}

func (s respSchema) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Schema)
}

func (s respSchema) MarshalText() (string, error) {
	// TODO: make it more readable
	msg := make([]string, 0)

	// now we print the metadata
	msg = append(msg, fmt.Sprintln("Tables:"))
	for _, t := range s.Schema.Tables {
		msg = append(msg, fmt.Sprintf("  %s\n", t.Name))
		msg = append(msg, "    Columns:\n")
		for _, c := range t.Columns {
			msg = append(msg, fmt.Sprintf("    %s\n", c.Name))
			msg = append(msg, fmt.Sprintf("      Type: %s\n", c.Type))
			for _, a := range c.Attributes {
				msg = append(msg, fmt.Sprintf("      %s\n", a.Type))
				if a.Value != "" {
					msg = append(msg, fmt.Sprintf("        %s\n", fmt.Sprint(a.Value)))
				}
			}
		}
	}

	// print queries
	msg = append(msg, "Actions:\n")
	for _, q := range s.Schema.Actions {
		msg = append(msg, fmt.Sprintf("  %s\n", q.Name))
		msg = append(msg, fmt.Sprintf("    Inputs: %s\n", q.Inputs))
	}

	return strings.Join(msg, ""), nil
}

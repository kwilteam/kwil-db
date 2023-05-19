package serialize

import (
	"encoding/json"
	"fmt"

	"github.com/kwilteam/kwil-db/pkg/engine2/dto"
)

func (s *Serializable) Table() (*dto.Table, error) {
	if s.Type != IdentifierTable {
		return nil, fmt.Errorf("cannot unserialize type '%s' as table", s.Type)
	}

	switch s.Version {
	case tableVersion:
		return tableVersion1ToTable(s.Data)
	default:
		return nil, fmt.Errorf("unknown table version: %d", s.Version)
	}
}

func SerializeTable(table *dto.Table) (*Serializable, error) {
	ser, err := convertTableToCurrentVersion(table)
	if err != nil {
		return nil, err
	}

	data, err := ser.Serialize()
	if err != nil {
		return nil, err
	}

	return &Serializable{
		Name:    table.Name,
		Type:    IdentifierTable,
		Version: tableVersion,
		Data:    data,
	}, nil
}

func convertTableToCurrentVersion(table *dto.Table) (serializer, error) {
	return &tableVersion1{
		Name: table.Name,
		Columns: func() []*columnVersion1 {
			var columns []*columnVersion1
			for _, col := range table.Columns {
				columns = append(columns, &columnVersion1{
					Name: col.Name,
					Type: col.Type.String(),
					Attributes: func() []*attributeVersion1 {
						var attrs []*attributeVersion1
						for _, attr := range col.Attributes {
							attrs = append(attrs, &attributeVersion1{
								Type:  attr.Type.String(),
								Value: attr.Value,
							})
						}
						return attrs
					}(),
				})
			}
			return columns
		}(),
		Indexes: func() []*indexVersion1 {
			var indexes []*indexVersion1
			for _, index := range table.Indexes {
				indexes = append(indexes, &indexVersion1{
					Name:    index.Name,
					Columns: index.Columns,
					Type:    index.Type.String(),
				})
			}
			return indexes
		}(),
	}, nil
}

func tableVersion1ToTable(data []byte) (*dto.Table, error) {
	tbl := &tableVersion1{}
	if err := json.Unmarshal(data, tbl); err != nil {
		return nil, err
	}

	return &dto.Table{
		Name: tbl.Name,
		Columns: func() []*dto.Column {
			var columns []*dto.Column
			for _, col := range tbl.Columns {
				columns = append(columns, &dto.Column{
					Name: col.Name,
					Type: dto.DataType(col.Type),
					Attributes: func() []*dto.Attribute {
						var attrs []*dto.Attribute
						for _, attr := range col.Attributes {
							attrs = append(attrs, &dto.Attribute{
								Type:  dto.AttributeType(attr.Type),
								Value: attr.Value,
							})
						}
						return attrs
					}(),
				})
			}
			return columns
		}(),
	}, nil
}

type tableVersion1 struct {
	Name    string            `json:"name"`
	Columns []*columnVersion1 `json:"columns"`
	Indexes []*indexVersion1  `json:"indexes"`
}

func (t *tableVersion1) Serialize() ([]byte, error) {
	return json.Marshal(t)
}

type columnVersion1 struct {
	Name       string               `json:"name"`
	Type       string               `json:"type"`
	Attributes []*attributeVersion1 `json:"attributes"`
}

type attributeVersion1 struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

type indexVersion1 struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Type    string   `json:"type"`
}

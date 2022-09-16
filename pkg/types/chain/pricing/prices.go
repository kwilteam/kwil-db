package pricing

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"
)

type operations struct {
	Database cruds `json:"database" yaml:"database" mapstructure:"database"`
	Table    cruds `json:"table" yaml:"table" mapstructure:"table"`
	Role     cruds `json:"role" yaml:"role" mapstructure:"role"`
	Query    cruds `json:"query" yaml:"query" mapstructure:"query"`
}

type cruds struct {
	Create string `json:"create" yaml:"create" mapstructure:"create"`
	Modify string `json:"modify" yaml:"modify" mapstructure:"modify"`
	Delete string `json:"delete" yaml:"delete" mapstructure:"delete"`
}

//type operations interface{}

type priceBuilder struct {
	prices *map[int16]int64
	op     byte
	cr     byte
}

func New(pbytes []byte) (*priceBuilder, error) {
	prices, err := readPrices(pbytes)
	if err != nil {
		return nil, err
	}

	return &priceBuilder{
		prices: parsePrices(prices),
	}, nil
}

func readPrices(p []byte) (*operations, error) {

	var prices operations
	err := json.Unmarshal(p, &prices)
	if err != nil {
		return nil, err
	}

	return &prices, nil
}

// initPrices converts the prices from the config file into a map
func parsePrices(p *operations) *map[int16]int64 {
	prices := make(map[int16]int64)

	// Loop through all keys in p and run determineOp
	// Then loop through all keys in the value and run determineCRUD
	// Then add the price to the map

	// get the operations
	ops := getFieldNames(*p)
	for _, op := range ops {
		// op is the operation (e.g. "table")

		// get the crud struct corresponding to the operation
		cstruct := getField(p, op).(cruds)

		// get the list of crud names (e.g. "create", "modify", "delete")
		cn := getFieldNames(cstruct)

		// loop through subfields
		for _, c := range cn {
			// cype is the CRUD (e.g. "create")
			// get the prices
			rint := getField(&cstruct, c).(string)

			// convert rint to int64
			r, err := strconv.ParseInt(rint, 10, 64)
			if err != nil {
				continue
			}

			// convert operation name to int8
			ob, err := determineOp(strings.ToLower(op))
			if err != nil {
				continue
			}

			// convert crud name to int8
			cb, err := determineCRUD(strings.ToLower(c))
			if err != nil {
				continue
			}

			// find final type
			id := make([]byte, 2)
			id[0] = byte(ob)
			id[1] = byte(cb)

			// turn id into int16
			id16 := byte2int16(id)

			// check if r is positive
			if r > 0 {
				prices[id16] = r
			}
		}
	}
	return &prices
}

var ErrOperationNotSupported = errors.New("operation not supported")

// determineOp converts the operation string to a byte
func determineOp(op string) (byte, error) {
	switch op {
	case "database":
		return 0, nil
	case "table":
		return 1, nil
	case "role":
		return 2, nil
	case "query":
		return 3, nil
	default:
		return 0, ErrOperationNotSupported
	}
}

var ErrCRUDNotSupported = errors.New("crud operation not supported")

// determineCRUD converts the crud string to a byte
func determineCRUD(crud string) (byte, error) {
	switch crud {
	case "create":
		return 0, nil
	case "modify":
		return 1, nil
	case "delete":
		return 2, nil
	default:
		return 0, ErrCRUDNotSupported
	}
}

type PriceBuilder interface {
	Operation(byte) PriceBuilder
	Crud(byte) PriceBuilder
	GetID(int16) int64
	Build() int64
}

func (p *priceBuilder) GetID(id int16) int64 {
	return (*p.prices)[id]
}

func (p *priceBuilder) Build() int64 {
	id := make([]byte, 2)
	id[0] = byte(p.op)
	id[1] = byte(p.cr)

	id16 := byte2int16(id)

	return (*p.prices)[id16]
}

func (p *priceBuilder) Operation(o byte) PriceBuilder {
	p.op = o
	return p
}

func (p *priceBuilder) Crud(c byte) PriceBuilder {
	p.cr = c
	return p
}

/*
	Conversion of operations to bytes

	// TODO: we should probably rename these, since operations is a better name for Cruds

	Operations: uint8
		- Database: 0
		- Table: 1
		- Role: 2
		- Query: 3

	Cruds:
		- Create: 0
		- Modify: 1
		- Delete: 2

	Example:
		- Database Create: 00
		- Table Delete: 12
*/

// utils

// interface for choosing either operation or cruds struct
type st interface {
	operations | cruds
}

func getFieldNames[S st](i S) []string {
	tt := reflect.TypeOf(i)
	ff := reflect.VisibleFields(tt)

	var r []string
	for _, f := range ff {
		r = append(r, f.Name)
	}

	return r
}

// getField gets the value of a field in a struct.  Must be type converted after
func getField(v any, field string) any {
	r := reflect.ValueOf(v)
	f := reflect.Indirect(r).FieldByName(field)
	return f.Interface()
}

func byte2int16(b []byte) int16 {
	var i int16
	_ = binary.Read(bytes.NewReader(b), binary.LittleEndian, &i)
	return i
}

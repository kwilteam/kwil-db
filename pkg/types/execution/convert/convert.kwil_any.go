package convert

import (
	"kwil/pkg/types/data_types/any_type"
	execution2 "kwil/pkg/types/execution"
)

type kwilAnyConversion struct{}

func (kwilAnyConversion) InputToBytes(v *execution2.UserInput[anytype.KwilAny]) *execution2.UserInput[[]byte] {
	return &execution2.UserInput[[]byte]{
		Name:  v.Name,
		Value: v.Value.Bytes(),
	}
}

func (kwilAnyConversion) BodyToBytes(v *execution2.ExecutionBody[anytype.KwilAny]) *execution2.ExecutionBody[[]byte] {
	var inputs []*execution2.UserInput[[]byte]
	for _, i := range v.Inputs {
		inputs = append(inputs, kwilAnyConversion{}.InputToBytes(i))
	}

	return &execution2.ExecutionBody[[]byte]{
		Database: v.Database,
		Query:    v.Query,
		Inputs:   inputs,
	}
}

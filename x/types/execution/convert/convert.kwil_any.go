package convert

import (
	anytype "kwil/x/types/data_types/any_type"
	"kwil/x/types/execution"
)

type kwilAnyConversion struct{}

func (kwilAnyConversion) InputToBytes(v *execution.UserInput[anytype.KwilAny]) *execution.UserInput[[]byte] {
	return &execution.UserInput[[]byte]{
		Name:  v.Name,
		Value: v.Value.Bytes(),
	}
}

func (kwilAnyConversion) BodyToBytes(v *execution.ExecutionBody[anytype.KwilAny]) *execution.ExecutionBody[[]byte] {
	var inputs []*execution.UserInput[[]byte]
	for _, i := range v.Inputs {
		inputs = append(inputs, kwilAnyConversion{}.InputToBytes(i))
	}

	return &execution.ExecutionBody[[]byte]{
		Database: v.Database,
		Query:    v.Query,
		Inputs:   inputs,
	}
}

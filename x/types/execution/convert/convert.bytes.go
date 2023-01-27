package convert

import (
	anytype "kwil/x/types/data_types/any_type"
	"kwil/x/types/execution"
)

// Byte conversions that will return an error if they fail.
type bytesConversion struct {
	Must bytesMustConversion
}

func (bytesConversion) InputToKwilAny(v *execution.UserInput[[]byte]) (*execution.UserInput[anytype.KwilAny], error) {
	val, err := anytype.NewFromSerial(v.Value)
	if err != nil {
		return nil, err
	}

	return &execution.UserInput[anytype.KwilAny]{
		Name:  v.Name,
		Value: val,
	}, nil
}

func (bytesConversion) BodyToKwilAny(v *execution.ExecutionBody[[]byte]) (*execution.ExecutionBody[anytype.KwilAny], error) {
	var inputs []*execution.UserInput[anytype.KwilAny]
	for _, i := range v.Inputs {
		input, err := bytesConversion{}.InputToKwilAny(i)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, input)
	}

	return &execution.ExecutionBody[anytype.KwilAny]{
		Database: v.Database,
		Query:    v.Query,
		Inputs:   inputs,
	}, nil
}

// Byte conversions that will panic if they fail.
type bytesMustConversion struct{}

func (m *bytesMustConversion) InputToKwilAny(v *execution.UserInput[[]byte]) *execution.UserInput[anytype.KwilAny] {
	val, err := anytype.NewFromSerial(v.Value)
	if err != nil {
		panic(err)
	}

	return &execution.UserInput[anytype.KwilAny]{
		Name:  v.Name,
		Value: val,
	}
}

// useful for tests
func (m *bytesMustConversion) SeveralInputToKwilAny(v []*execution.UserInput[[]byte]) []*execution.UserInput[anytype.KwilAny] {
	var ret []*execution.UserInput[anytype.KwilAny]
	for _, i := range v {
		ret = append(ret, m.InputToKwilAny(i))
	}
	return ret
}

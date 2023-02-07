package convert

import (
	"kwil/pkg/types/data_types/any_type"
	execution2 "kwil/pkg/types/execution"
)

// Byte conversions that will return an error if they fail.
type bytesConversion struct {
	Must bytesMustConversion
}

func (bytesConversion) InputToKwilAny(v *execution2.UserInput[[]byte]) (*execution2.UserInput[anytype.KwilAny], error) {
	val, err := anytype.NewFromSerial(v.Value)
	if err != nil {
		return nil, err
	}

	return &execution2.UserInput[anytype.KwilAny]{
		Name:  v.Name,
		Value: val,
	}, nil
}

func (bytesConversion) BodyToKwilAny(v *execution2.ExecutionBody[[]byte]) (*execution2.ExecutionBody[anytype.KwilAny], error) {
	var inputs []*execution2.UserInput[anytype.KwilAny]
	for _, i := range v.Inputs {
		input, err := bytesConversion{}.InputToKwilAny(i)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, input)
	}

	return &execution2.ExecutionBody[anytype.KwilAny]{
		Database: v.Database,
		Query:    v.Query,
		Inputs:   inputs,
	}, nil
}

// Byte conversions that will panic if they fail.
type bytesMustConversion struct{}

func (m *bytesMustConversion) InputToKwilAny(v *execution2.UserInput[[]byte]) *execution2.UserInput[anytype.KwilAny] {
	val, err := anytype.NewFromSerial(v.Value)
	if err != nil {
		panic(err)
	}

	return &execution2.UserInput[anytype.KwilAny]{
		Name:  v.Name,
		Value: val,
	}
}

// useful for tests
func (m *bytesMustConversion) SeveralInputToKwilAny(v []*execution2.UserInput[[]byte]) []*execution2.UserInput[anytype.KwilAny] {
	var ret []*execution2.UserInput[anytype.KwilAny]
	for _, i := range v {
		ret = append(ret, m.InputToKwilAny(i))
	}
	return ret
}

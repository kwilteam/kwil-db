package tree

import (
	"errors"
	"reflect"
)

func run(errs ...error) error {
	return errors.Join(errs...)
}

// AstWalker represents an AST node that can be walked.
type AstWalker interface {
	// Walk walks through itself using AstListener.
	Walk(AstListener) error
}

func isNil(input interface{}) bool {
	if input == nil {
		return true
	}
	kind := reflect.ValueOf(input).Kind()
	switch kind {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Chan:
		return reflect.ValueOf(input).IsNil()
	default:
		return false
	}
}

func walk(v AstListener, a AstWalker) error {
	if isNil(a) {
		return nil
	}

	return a.Walk(v)
}

func walkMany[T AstWalker](v AstListener, as []T) error {
	for _, a := range as {
		err := walk(v, a)
		if err != nil {
			return err
		}
	}

	return nil
}

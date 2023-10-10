package extensions

import (
	"context"
	"fmt"
	"reflect"

	"github.com/cstockton/go-conv"
)

type extensionBuilder struct {
	extension *Extension
}

// ExtensionBuilder is the interface for creating an extension server
type ExtensionBuilder interface {
	// WithMethods specifies the methods that should be provided
	// by the extension
	WithMethods(map[string]MethodFunc) ExtensionBuilder
	// WithInitializer is a function that initializes a new extension instance.
	WithInitializer(InitializeFunc) ExtensionBuilder
	// Named specifies the name of the extensions.
	Named(string) ExtensionBuilder

	// Build creates the extensions
	Build() (*Extension, error)
}

func Builder() ExtensionBuilder {
	return &extensionBuilder{
		extension: &Extension{
			methods: make(map[string]MethodFunc),
			initializeFunc: func(ctx context.Context, metadata map[string]string) (map[string]string, error) {
				return metadata, nil
			},
		},
	}
}

func (b *extensionBuilder) Named(name string) ExtensionBuilder {
	b.extension.name = name
	return b
}

func (b *extensionBuilder) WithMethods(methods map[string]MethodFunc) ExtensionBuilder {
	b.extension.methods = methods
	return b
}

func (b *extensionBuilder) WithInitializer(fn InitializeFunc) ExtensionBuilder {
	b.extension.initializeFunc = fn
	return b
}

func (b *extensionBuilder) Build() (*Extension, error) {
	return b.extension, nil
}

type ExecutionContext struct {
	Ctx      context.Context
	Metadata map[string]string
}

// MethodFunc is a function that executes a method
type MethodFunc func(ctx *ExecutionContext, inputs ...*ScalarValue) ([]*ScalarValue, error)

// InitializeFunc is a function that creates a new instance of an extension.
// In most cases, this should just validate the metadata being sent.
type InitializeFunc func(ctx context.Context, metadata map[string]string) (map[string]string, error)

// WithInputsCheck checks the number of inputs.
// If the number of inputs is not equal to numInputs, it returns an error.
func WithInputsCheck(fn MethodFunc, numInputs int) MethodFunc {
	return func(ctx *ExecutionContext, inputs ...*ScalarValue) ([]*ScalarValue, error) {
		if len(inputs) != numInputs {
			return nil, fmt.Errorf("expected %d args, got %d", numInputs, len(inputs))
		}
		return fn(ctx, inputs...)
	}
}

// WithOutputsCheck checks the number of outputs.
// If the number of outputs is not equal to numOutputs, it returns an error.
func WithOutputsCheck(fn MethodFunc, numOutputs int) MethodFunc {
	return func(ctx *ExecutionContext, inputs ...*ScalarValue) ([]*ScalarValue, error) {
		res, err := fn(ctx, inputs...)
		if err != nil {
			return nil, err
		}

		if len(res) != numOutputs {
			return nil, fmt.Errorf("expected %d returns, got %d", numOutputs, len(res))
		}

		return res, nil
	}
}

type ScalarValue struct {
	Value any
}

func NewScalarValue(v any) (*ScalarValue, error) {
	valueType := reflect.TypeOf(v)
	switch valueType.Kind() {
	case reflect.String, reflect.Float32, reflect.Float64:
		return &ScalarValue{
			Value: v,
		}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &ScalarValue{
			Value: v,
		}, nil
	default:
		return nil, fmt.Errorf("invalid scalar type: %s", valueType.Kind())
	}
}

// String returns the string representation of the value.
func (s *ScalarValue) String() (string, error) {
	return conv.String(s.Value)
}

// Int returns the int representation of the value.
func (s *ScalarValue) Int() (int64, error) {
	return conv.Int64(s.Value)
}

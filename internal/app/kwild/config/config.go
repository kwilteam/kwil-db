package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/cstockton/go-conv"
	"github.com/spf13/viper"
)

func setupViper() {
	// Allow reading environment variables with a prefix
	viper.SetEnvPrefix(EnvPrefix) // This will look for environment variables with the "MYAPP_" prefix

	// Replace '.' in the key with '_' when looking up environment variables
	//viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Automatically read environment variables
	viper.AutomaticEnv()
}

type cfgVar struct {
	// EnvName is the name of the environment variable to use
	EnvName string

	// Required is true if the variable is required.
	// This will cause an error if the variable is not set.
	Required bool

	// Default is the default value to use if the variable is not set.
	Default any

	// Setter is a function that will be called to set the value of the variable.
	Setter func(any) (any, error)

	// Field is the name of the field in the config struct to set.
	// It can be nested using dot notation.
	// For example, "Server.Port" will set the "Port" field of the "Server" field.
	Field string
}

func LoadConfig() (*KwildConfig, error) {
	setupViper()

	c := &KwildConfig{}

	for _, v := range RegisteredVariables {
		if err := insertVar(c, v); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func insertVar(c *KwildConfig, v cfgVar) error {
	var value any
	if viper.IsSet(v.EnvName) {
		value = viper.Get(v.EnvName)
	}

	if value == nil && v.Default != nil {
		value = v.Default
	}

	if v.Required && value == nil {
		return fmt.Errorf("missing required environment variable %s", v.EnvName)
	}

	field, err := locateField(c, v.Field)
	if err != nil {
		return fmt.Errorf("error locating field on variable '%s': %s", v.EnvName, err)
	}

	if v.Setter != nil {
		value, err = v.Setter(value)
		if err != nil {
			return err
		}

		field.Set(reflect.ValueOf(value))
	} else {
		if value != nil {
			if err := convertVal(field, value); err != nil {
				return err
			}
		}
	}

	return nil
}

// locateField will locate and return the field in the given struct.
// The field can be nested using dot notation.
// For example, "Server.Port" will return the "Port" field of the "Server" field.
func locateField(obj any, field string) (*reflect.Value, error) {
	// Get the value of the struct
	structValue := reflect.ValueOf(obj)
	if structValue.Kind() != reflect.Ptr || structValue.IsNil() {
		return nil, fmt.Errorf("obj must be a non-nil pointer to a struct")
	}
	// Dereference the pointer to get the struct
	rv := structValue.Elem()
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("obj must be a pointer to a struct")
	}

	fields := strings.Split(field, ".")

	for _, f := range fields {
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}

		if rv.Kind() != reflect.Struct {
			return nil, fmt.Errorf("config field '%s' is of type '%s', must be a struct", f, rv.Kind())
		}

		rv = rv.FieldByName(f)
		if !rv.IsValid() {
			return nil, fmt.Errorf("field '%s' not found", f)
		}
	}

	return &rv, nil
}

// convertVal converts a value to the given type.
// this can then be used to set the value of a struct field.
// this is helpful with native types like strings, ints, etc.
// where you don't want to write a full setter function.
func convertVal(field *reflect.Value, val any) error {
	switch field.Type().Kind() {
	case reflect.String:
		strVal, err := conv.String(val)
		if err != nil {
			return err
		}

		field.SetString(strVal)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := conv.Int64(val)
		if err != nil {
			return err
		}

		field.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := conv.Uint64(val)
		if err != nil {
			return err
		}

		field.SetUint(uintVal)
	case reflect.Bool:
		boolVal, err := conv.Bool(val)
		if err != nil {
			return err
		}

		field.SetBool(boolVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := conv.Float64(val)
		if err != nil {
			return err
		}

		field.SetFloat(floatVal)
	default:
		return fmt.Errorf("unsupported type %s", field.Type().Kind())
	}

	return nil
}

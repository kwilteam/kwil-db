package config

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/cstockton/go-conv"
	"github.com/spf13/viper"
)

func setupViper(envPrefix string) {
	// Allow reading environment variables with a prefix

	if envPrefix != "" {
		viper.SetEnvPrefix(envPrefix) // This will look for environment variables with the "MYAPP_" prefix
	}

	// Replace '.' in the key with '_' when looking up environment variables
	//viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Automatically read environment variables
	viper.AutomaticEnv()
}

type CfgVar struct {
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

	// Flag is the flag to use for this variable.
	Flag Flag
}

// LoadConfig will load the configuration from environment variables.
// It can take a default configuration struct to use as a base.
// The value's order of precedence is:
// 1. Environment variable
// 2. Default config
// 3. CfgVar default value
// 4. Zero value
// If the CfgVar is required, but the value is zeroed, an error will be returned.
func LoadConfig[T any](vars []CfgVar, envPrefix string, defaultConfig *T) error {
	setupViper(envPrefix)

	if defaultConfig == nil {
		return fmt.Errorf("defaultConfig must be non-nil")
	}

	for _, v := range vars {
		if err := insertVar(defaultConfig, v); err != nil {
			return err
		}
	}

	return nil
}

func insertVar(configStruct any, v CfgVar) error {
	field, err := locateField(configStruct, v.Field)
	if err != nil {
		return fmt.Errorf("error locating field on variable '%s': %s", v.EnvName, err)
	}

	// dictating value precedence
	var value any
	if viper.IsSet(v.EnvName) {
		fmt.Println("setting value from env var", v.EnvName)
		value = viper.Get(v.EnvName)
	} else if !field.IsZero() {
		// do nothing, leave the value as is
		return nil
	} else if v.Default != nil {
		fmt.Println("setting value from default")
		value = v.Default
	} else if v.Required && value == nil {
		return fmt.Errorf("missing required environment variable %s", v.EnvName)
	}

	if v.Setter != nil {
		value, err = v.Setter(value)
		if err != nil {
			return err
		}

		if value != nil {
			field.Set(reflect.ValueOf(value))
		}
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

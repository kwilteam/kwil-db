package clean

import (
	"kwil/pkg/engine/models/clean/traverse"
	"reflect"
	"strings"
)

const (
	tagName = "clean"

	// Lower cleans the string to lower case
	Lower = "lower"

	// IsEnum checks if an int is a valid enum value
	// It should have an additional parameter with the enum name
	IsEnum = "is_enum"
)

// Clean cleans the struct.  The struct must have a tag with the name "clean"
// It also must be passed as a pointer
func Clean(val interface{}) {
	if val == nil {
		return
	}

	traverser := traverse.New(tagName, func(v reflect.Value, tags []string) {
		switch tags[0] {
		case Lower:
			cleanString(v)
		case IsEnum:
			cleanEnum(v, tags)
		}
	})

	traverser.Traverse(val)
}

func cleanString(v reflect.Value) {
	if v.Kind() == reflect.Slice || v.Kind() == reflect.Array {
		for i := 0; i < v.Len(); i++ {
			cleanString(v.Index(i))
		}
	} else if v.Kind() == reflect.String {
		v.SetString(strings.ToLower(v.String()))
	}
}

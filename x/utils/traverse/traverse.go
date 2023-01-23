package traverse

import (
	"reflect"
	"strings"
)

const (
	// the traverse tag specifies the name of the tag to use for traversal instructions
	// by default, all fields are traversed recursively (deep)
	traverseTagName = "traverse"

	// shallow indicates that this field should not be traversed recursively
	Shallow = "shallow"

	// skip indicates that this field should not be traversed at all
	Skip = "skip"
)

/*
	Traverse will recursively traverse through a struct and pass the reflect.Value, and tags (split by ,) to the callback function

	For example, if you have a struct like this:

	type User struct {
		Name string `tag1:"value1" tag2:"value2"`
		Age int    `tag1:"value3" tag2:"value4"`
	}

	You can traverse through it like this:

	traverse.Traverse(&User{}, tagName, func(v reflect.Value, tags []string) {
		fmt.Println(v, tags)
	}

*/

type Traverser interface {
	Traverse(val interface{})
}

type traverser struct {
	tag      string
	callback func(v reflect.Value, tags []string)
}

func New(tag string, callback func(v reflect.Value, tags []string)) Traverser {
	return &traverser{
		tag:      tag,
		callback: callback,
	}
}

func (t *traverser) Traverse(input interface{}) {
	if input == nil {
		return
	}

	v := reflect.ValueOf(input)

	t.traverseReflection(v)
}

func (t *traverser) traverseReflection(v reflect.Value) {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		v = v.Elem()
	}

	t.forEachTraversableField(v)
}

// will call the callback function for each field of "v" that is traversable
func (t *traverser) forEachTraversableField(v reflect.Value) {
	for i := 0; i < v.NumField(); i++ {
		value := v.Field(i)
		traversal := v.Type().Field(i).Tag.Get(traverseTagName)

		if !value.CanSet() {
			continue
		}

		field := v.Type().Field(i)

		// map all tag values to their respective tag names
		tag := field.Tag.Get(t.tag)
		tags := strings.Split(tag, ",")

		switch traversal {
		case Skip:
			continue
		case Shallow:
			t.callback(value, tags)
			continue
		default:
			t.deepTraverse(value, tags)
		}
	}
}

// deep traverse will traverse a field recursively
// it will traverse through each value of arrays / slices,
// and will traverse through each field of structs
// if the value is neither of these, it will call the callback function
func (t *traverser) deepTraverse(v reflect.Value, tags []string) {
	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		t.deepTraverseArray(v, tags)
	case reflect.Struct:
		t.traverseReflection(v)
	default:
		t.callback(v, tags)
	}
}

func (t *traverser) deepTraverseArray(v reflect.Value, tags []string) {
	length := v.Len()
	if length == 0 {
		return
	}

	switch v.Index(0).Kind() {
	case reflect.Struct:
		for i := 0; i < length; i++ {
			t.traverseReflection(v.Index(i))
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < length; i++ {
			t.deepTraverseArray(v.Index(i), tags)
		}
	case reflect.Pointer:
		for i := 0; i < length; i++ {
			t.traverseReflection(v.Index(i).Elem())
		}
	default:
		t.callback(v, tags)
	}
}

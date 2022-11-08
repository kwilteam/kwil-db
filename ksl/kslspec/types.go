package kslspec

import (
	"fmt"
	"strconv"
)

type Type interface{ typ() }
type Value interface{ val() }

type ConcreteType struct {
	Type  string
	Attrs map[string]*Attr
}

func (c *ConcreteType) AddAttr(name string, value Value) {
	if c.Attrs == nil {
		c.Attrs = make(map[string]*Attr, 1)
	}
	c.Attrs[name] = &Attr{Name: name, Value: value}
}

func (c *ConcreteType) AddIntAttr(name string, value int) {
	c.AddAttr(name, &LiteralValue{Value: strconv.Itoa(value)})
}

func (c *ConcreteType) AddStringAttr(name, value string) {
	c.AddAttr(name, &LiteralValue{Value: value})
}

func (c *ConcreteType) AddBoolAttr(name string, value bool) {
	c.AddAttr(name, &LiteralValue{Value: strconv.FormatBool(value)})
}

type Attr struct {
	Name  string
	Value Value
}

type UnsupportedType struct {
	T string
}

func (UnsupportedType) typ() {}

func (a *Attr) Bool() (bool, error) {
	lit, ok := a.Value.(*LiteralValue)
	if !ok {
		return false, fmt.Errorf("schema: cannot read attribute %q as literal", a.Name)
	}
	b, err := strconv.ParseBool(lit.Value)
	if err != nil {
		return false, fmt.Errorf("schema: cannot read attribute %q as bool: %w", a.Name, err)
	}
	return b, nil
}

func (a *Attr) Int() (int, error) {
	i, err := a.Int64()
	return int(i), err
}

func (a *Attr) Int64() (int64, error) {
	lit, ok := a.Value.(*LiteralValue)
	if !ok {
		return 0, fmt.Errorf("schema: cannot read attribute %q as literal", a.Name)
	}
	i, err := strconv.Atoi(lit.Value)
	if err != nil {
		return 0, fmt.Errorf("schema: cannot read attribute %q as int: %w", a.Name, err)
	}
	return int64(i), nil
}

func (a *Attr) String() (string, error) {
	lit, ok := a.Value.(*LiteralValue)
	if !ok {
		return "", fmt.Errorf("schema: cannot read attribute %q as literal", a.Name)
	}
	return lit.Value, nil
}

type LiteralValue struct{ Value string }
type ListValue struct{ Values []Value }

func (LiteralValue) val() {}
func (ListValue) val()    {}

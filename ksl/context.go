package ksl

import (
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

type Expression interface {
	Eval(ctx *Context) (cty.Value, Diagnostics)
	Range() Range
}

type Context struct {
	Variables map[string]cty.Value
	Functions map[string]function.Function
	parent    *Context
}

func (ctx *Context) NewChild() *Context {
	return &Context{parent: ctx}
}

func (ctx *Context) Parent() *Context {
	return ctx.parent
}

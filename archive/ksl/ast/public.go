package ast

import (
	"ksl"
	"ksl/ast/pdb"
	"ksl/syntax"
	"ksl/syntax/nodes"
	"reflect"
	"unsafe"
)

func ParseFiles(files ...string) *SchemaAst {
	fs, diags := syntax.ParseFiles(files...)
	return New(pdb.New(fs, ksl.NewContext(), diags))
}

func ParseString(src string, filename string) *SchemaAst {
	hdr := (*reflect.StringHeader)(unsafe.Pointer(&src))
	data := (*[1<<31 - 1]byte)(unsafe.Pointer(hdr.Data))[:hdr.Len]
	return Parse(data, filename)
}

func Parse(src []byte, filename string) *SchemaAst {
	fs, diags := syntax.Parse(src, filename, ksl.InitialPos)
	return New(pdb.New([]*nodes.File{fs}, ksl.NewContext(), diags))
}

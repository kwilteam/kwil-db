package schema

import (
	"ksl"
	"ksl/pdb"
	"ksl/syntax"
	"ksl/syntax/ast"
	"reflect"
	"unsafe"
)

func ParseFiles(files ...string) *KwilSchema {
	fs, diags := syntax.ParseFiles(files...)
	return New(pdb.New(fs, ksl.NewContext(), diags))
}

func ParseString(src string, filename string) *KwilSchema {
	hdr := (*reflect.StringHeader)(unsafe.Pointer(&src))
	data := (*[1<<31 - 1]byte)(unsafe.Pointer(hdr.Data))[:hdr.Len]
	return Parse(data, filename)
}

func Parse(src []byte, filename string) *KwilSchema {
	fs, diags := syntax.Parse(src, filename, ksl.InitialPos)
	return New(pdb.New([]*ast.File{fs}, ksl.NewContext(), diags))
}

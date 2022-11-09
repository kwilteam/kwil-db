package kslformat

import (
	"reflect"
	"unsafe"

	"ksl"
	"ksl/kslsyntax"
	"ksl/kslsyntax/ast"
)

func Format(src []byte) ([]byte, error) {
	doc, diags := kslsyntax.Parse(src, "", ksl.InitialPos)
	if diags.HasErrors() {
		return nil, diags
	}
	data := FormatDoc(doc)
	hdr := (*reflect.StringHeader)(unsafe.Pointer(&data))
	b := (*[1<<31 - 1]byte)(unsafe.Pointer(hdr.Data))[:hdr.Len]
	return b, nil
}

func FormatString(src string) (string, error) {
	hdr := (*reflect.StringHeader)(unsafe.Pointer(&src))
	b := (*[1<<31 - 1]byte)(unsafe.Pointer(hdr.Data))[:hdr.Len]
	data, err := Format(b)
	if err != nil {
		return "", err
	}
	return *(*string)(unsafe.Pointer(&data)), nil
}

func FormatDoc(src *ast.Document) string {
	return formatDocument(src)
}

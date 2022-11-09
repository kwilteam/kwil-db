package kslformat

import (
	"strings"

	"ksl/kslsyntax/ast"
)

func formatDocument(doc *ast.Document) string {
	var builder strings.Builder

	nodes := make([]ast.Node, 0, len(doc.Directives)+len(doc.Blocks))
	for _, d := range doc.Directives {
		nodes = append(nodes, d)
	}
	for _, b := range doc.Blocks {
		nodes = append(nodes, b)
	}
	groups := groupNodes(nodes)
	_ = groups

	return builder.String()
}

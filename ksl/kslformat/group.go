package kslformat

import (
	"math"
	"reflect"
	"sort"

	"ksl"
	"ksl/kslsyntax/ast"
)

type byOffset struct {
	Nodes []ast.Node
}

func (s byOffset) Len() int {
	return len(s.Nodes)
}

func (s byOffset) Less(i, j int) bool {
	rangeI := s.Nodes[i].Range()
	rangeJ := s.Nodes[j].Range()
	return rangeI.Start.Offset < rangeJ.Start.Offset
}

func (s byOffset) Swap(i, j int) {
	s.Nodes[i], s.Nodes[j] = s.Nodes[j], s.Nodes[i]
}

func groupNodes(nodes []ast.Node) [][]ast.Node {
	sort.Sort(byOffset{nodes})

	var groups [][]ast.Node
	var group []ast.Node
	var lastRange ksl.Range
	var lastType reflect.Type

	for _, node := range nodes {
		curRange := node.Range()
		switch {
		case lastType == nil:
			group = append(group, node)
		case math.Abs(float64(curRange.Start.Line-lastRange.End.Line)) > 1 || reflect.TypeOf(node) != lastType:
			groups = append(groups, group)
			group = []ast.Node{node}
		default:
			group = append(group, node)
		}
		lastRange = curRange
		lastType = reflect.TypeOf(node)
	}
	if len(group) > 0 {
		groups = append(groups, group)
	}
	return groups
}

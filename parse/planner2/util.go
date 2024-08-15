package planner2

import (
	"fmt"
	"strings"
)

// visitAllExprs visits all nodes in a tree, calling the given function on each expression.
func visitAllNodes(node LogicalNode, f func(Traversable)) error {
	// this uses the rewrite but doesn't actually rewrite anything
	_, err := Rewrite(node, &RewriteConfig{
		ExprCallback: func(e LogicalExpr) (LogicalExpr, bool, error) {
			f(e)
			return e, true, nil
		},
		PlanCallback: func(p LogicalPlan) (LogicalPlan, bool, error) {
			f(p)
			return p, true, nil
		},
		ScanSourceCallback: func(s ScanSource) (ScanSource, bool, error) {
			f(s)
			return s, true, nil
		},
	})
	return err
}

// traverse traverses a logical plan in preorder.
// It will call the callback function for each node in the plan.
// If the callback function returns false, the traversal will not
// continue to the children of the node.
func traverse(node Traversable, callback func(node Traversable) bool) {
	if !callback(node) {
		return
	}
	for _, child := range node.Children() {
		if child == nil {
			fmt.Println("nil child")
		}
		traverse(child, callback)
	}
}

func Format(plan LogicalNode) string {
	str := strings.Builder{}
	inner, topLevel := innerFormat(plan, 0, []bool{})
	str.WriteString(inner)

	printSubplans(&str, topLevel)

	return str.String()
}

// printSubplans is a recursive function that prints the subplans
func printSubplans(str *strings.Builder, subplans []*Subplan) {
	for _, sub := range subplans {
		str.WriteString(sub.String())
		str.WriteString("\n")
		strs, subs := innerFormat(sub.Plan, 1, []bool{false})
		str.WriteString(strs)
		printSubplans(str, subs)
	}
}

// innerFormat is a function that allows us to give more complex
// formatting logic.
// It returns subplans that should be added to the top level.
func innerFormat(plan LogicalNode, count int, printLong []bool) (string, []*Subplan) {
	if sub, ok := plan.(*Subplan); ok {
		return "", []*Subplan{sub}
	}

	var msg strings.Builder
	for i := 0; i < count; i++ {
		if i == count-1 && len(printLong) > i && !printLong[i] {
			msg.WriteString("└─")
		} else if i == count-1 && len(printLong) > i && printLong[i] {
			msg.WriteString("├─")
		} else if len(printLong) > i && printLong[i] {
			msg.WriteString("│ ")
		} else {
			msg.WriteString("  ")
		}
	}
	msg.WriteString(plan.String())
	msg.WriteString("\n")
	var topLevel []*Subplan
	plans := plan.Plans()
	for i, child := range plans {
		showLong := true
		// if it is the last plan, or if the next plan is a subplan,
		// we should not show the long line
		if i == len(plans)-1 {
			showLong = false
		} else if _, ok := plans[i+1].(*Subplan); ok {
			showLong = false
		}

		str, children := innerFormat(child, count+1, append(printLong, showLong))
		msg.WriteString(str)
		topLevel = append(topLevel, children...)
	}
	return msg.String(), topLevel
}

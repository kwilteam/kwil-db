package ast

import (
	"ksl"
	"strings"
)

type CommentGroup struct {
	Comments []*Comment
	Span     ksl.Range
}

func (c *CommentGroup) Range() ksl.Range { return c.Span }

func (c *CommentGroup) String() string {
	if c == nil {
		return ""
	}

	lines := make([]string, len(c.Comments))
	for i, comment := range c.Comments {
		lines[i] = comment.String()
	}
	return strings.Join(lines, " ")
}

type Comment struct {
	Text string
	Span ksl.Range
}

func (c *Comment) String() string {
	if c == nil {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(c.Text, "///"))
}

func (c Comment) Range() ksl.Range { return c.Span }

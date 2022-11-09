package ast

import "ksl"

type Node interface {
	Range() ksl.Range
}

func (n *Document) Range() ksl.Range     { return n.SrcRange }
func (n *Block) Range() ksl.Range        { return n.SrcRange }
func (n *Directive) Range() ksl.Range    { return n.SrcRange }
func (n *Body) Range() ksl.Range         { return n.SrcRange }
func (n *Definition) Range() ksl.Range   { return n.SrcRange }
func (n *Attribute) Range() ksl.Range    { return n.SrcRange }
func (n *Type) Range() ksl.Range         { return n.SrcRange }
func (n *Var) Range() ksl.Range          { return n.SrcRange }
func (n *Float) Range() ksl.Range        { return n.SrcRange }
func (n *Str) Range() ksl.Range          { return n.SrcRange }
func (n *QuotedStr) Range() ksl.Range    { return n.SrcRange }
func (n *Heredoc) Range() ksl.Range      { return n.SrcRange }
func (n *Bool) Range() ksl.Range         { return n.SrcRange }
func (n *FunctionCall) Range() ksl.Range { return n.SrcRange }
func (n *List) Range() ksl.Range         { return n.SrcRange }
func (n *Null) Range() ksl.Range         { return n.SrcRange }
func (n *Int) Range() ksl.Range          { return n.SrcRange }
func (n *Number) Range() ksl.Range       { return n.SrcRange }
func (n *Object) Range() ksl.Range       { return n.SrcRange }
func (n *BlockLabels) Range() ksl.Range  { return n.SrcRange }
func (n *ArgList) Range() ksl.Range      { return n.SrcRange }
func (n *Annotation) Range() ksl.Range   { return n.SrcRange }

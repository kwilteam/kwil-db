package syntax

import (
	"ksl"
	"ksl/syntax/lex"
	"ksl/syntax/nodes"
	"strings"

	"golang.org/x/exp/slices"
)

type parser struct {
	*peeker
	recovery bool
}

func (p *parser) parseFile(src []byte, filename string) (*nodes.File, ksl.Diagnostics) {
	var diags ksl.Diagnostics
	var entries []nodes.TopLevel
	var pendingComment *nodes.CommentGroup
	cancelPendingComment := func() {
		if pendingComment != nil {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagWarning,
				Summary:  DiagFloatingDocComment,
				Detail:   "This doc comment is floating and isn't attached to any node.",
				Subject:  pendingComment.Span.Ptr(),
			})
		}
		pendingComment = nil
	}

	startRange := p.NextRange()

Token:
	for {
		next := p.Peek()

		var entry nodes.TopLevel
		var entryDiags ksl.Diagnostics

		switch next.Type {
		case lex.TokenDocComment:
			pendingComment = p.parseComments()
			continue

		case lex.TokenModel:
			entry, entryDiags = p.parseModel(pendingComment)

		case lex.TokenEnum:
			entry, entryDiags = p.parseEnum(pendingComment)

		case lex.TokenIdent:
			entry, entryDiags = p.parseBlock(pendingComment)

		case lex.TokenAt:
			entry, entryDiags = p.parseDirective()
			cancelPendingComment()

		case lex.TokenNewline:
			p.Read()
			cancelPendingComment()
			continue

		case lex.TokenEOF:
			p.Read()
			pendingComment = nil
			break Token

		default:
			bad := p.Read()
			recovering := p.recovery
			eol := p.recover(lex.TokenNewline)
			if !recovering {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  DiagInvalidLine,
					Detail:   "This line is invalid. A directive, model, or block defintion was expected here.",
					Subject:  ksl.RangeBetween(bad.Range, eol.Range).Ptr(),
				})
			}
			cancelPendingComment()
			continue
		}

		entries = append(entries, entry)
		diags = append(diags, entryDiags...)
		pendingComment = nil
	}

	return &nodes.File{
		Name:     filename,
		Entries:  entries,
		Contents: src,
		Span:     ksl.RangeBetween(startRange, p.PrevRange()),
	}, diags
}

func (p *parser) parseDirective() (*nodes.Annotation, ksl.Diagnostics) {
	return p.parseFieldAnnotation()
}

func (p *parser) parseModel(doc *nodes.CommentGroup) (*nodes.Model, ksl.Diagnostics) {
	var diags ksl.Diagnostics
	var name *nodes.Name
	var fields nodes.Fields
	var annotations nodes.Annotations
	var pendingComment *nodes.CommentGroup

	cancelPendingComment := func() {
		if pendingComment != nil {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagWarning,
				Summary:  DiagFloatingDocComment,
				Detail:   "This doc comment is floating and isn't attached to any node.",
				Subject:  pendingComment.Span.Ptr(),
			})
		}
		pendingComment = nil
	}

	p.PushIncludeNewlines(true)
	defer p.PopIncludeNewlines()

	typeTok := p.Peek()
	if typeTok.Type != lex.TokenModel {
		return nil, append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagExpectedModelDefinition,
			Detail:   "This line is invalid. A model definition was expected here.",
			Subject:  typeTok.Range.Ptr(),
		})
	}
	p.Read()

	nameTok := p.Peek()
	if !nameTok.IsAny(lex.TokenIdent, lex.TokenQualifiedIdent) {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagExpectedModelDefinition,
			Detail:   "Expected a model name here.",
			Subject:  nameTok.Range.Ptr(),
		})
		p.recoverTo(lex.TokenNewline, lex.TokenLBrace)
	} else {
		p.Read()
		name = &nodes.Name{Name: nameTok.Value, Span: nameTok.Range}
	}

	next := p.Peek()
	if next.Type == lex.TokenLBrace {
		p.Read()

	Token:
		for {
			first := p.Peek()

			switch first.Type {
			case lex.TokenDocComment:
				pendingComment = p.parseComments()
				continue

			case lex.TokenRBrace:
				p.Read()
				cancelPendingComment()
				break Token

			case lex.TokenNewline:
				p.Read()
				cancelPendingComment()
				continue

			case lex.TokenAt:
				annot, annotDiags := p.parseBlockAnnotation()
				annotations = append(annotations, annot)
				diags = append(diags, annotDiags...)
				if p.recovery && annotDiags.HasErrors() {
					p.recover(lex.TokenNewline)
				}
				cancelPendingComment()

			case lex.TokenIdent:
				fld, fldDiags := p.parseField(pendingComment)
				fields = append(fields, fld)
				diags = append(diags, fldDiags...)
				if p.recovery && fldDiags.HasErrors() {
					p.recover(lex.TokenNewline)
				}
				pendingComment = nil

			default:
				if !p.recovery {
					diags = append(diags, &ksl.Diagnostic{
						Severity: ksl.DiagError,
						Summary:  DiagUnexpectedToken,
						Detail:   "Expected an enum value or a block annotation.",
						Subject:  &first.Range,
					})
				}
				cancelPendingComment()
				p.recover(lex.TokenRBrace)
				break Token
			}
		}
	} else {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagInvalidModelDefinition,
			Detail:   "A model must have a body.",
			Subject:  next.Range.Ptr(),
		})
	}

	return &nodes.Model{
		Name:        name,
		Fields:      fields,
		Annotations: annotations,
		Comment:     doc,
		Span:        ksl.RangeBetween(typeTok.Range, p.PrevRange()),
	}, diags
}

func (p *parser) parseField(doc *nodes.CommentGroup) (*nodes.Field, ksl.Diagnostics) {
	var diags ksl.Diagnostics
	var name, typ *nodes.Name
	var annotations nodes.Annotations
	var arity nodes.FieldArity
	var fieldType *nodes.FieldType

	nameTok := p.Peek()
	if nameTok.Type != lex.TokenIdent {
		endTok := p.recover(lex.TokenNewline)
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagInvalidFieldDefinition,
			Detail:   "This line is invalid. A field definition was expected here.",
			Subject:  ksl.RangeBetween(nameTok.Range, endTok.Range).Ptr(),
		})
		return nil, diags
	}

	p.Read()
	name = &nodes.Name{Name: nameTok.Value, Span: nameTok.Range}

	typTok := p.Peek()
	if !typTok.IsAny(lex.TokenIdent, lex.TokenQualifiedIdent) {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagInvalidDeclarationName,
			Detail:   "A declaration name must be an identifier.",
			Subject:  typTok.Range.Ptr(),
		})
		p.recover(lex.TokenNewline)
	} else {
		p.Read()
		typ = &nodes.Name{Name: typTok.Value, Span: typTok.Range}
	}

	tok, next := p.Peek2()
	if tok.Type == lex.TokenLBrack && next.Type == lex.TokenRBrack {
		p.ReadN(2)
		arity = nodes.Repeated
		tok = p.Peek()
	}

	if tok.Type == lex.TokenQuestion {
		p.Read()
		if arity == nodes.Repeated {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagInvalidFieldDefinition,
				Detail:   "A repeated field cannot be optional.",
				Subject:  tok.Range.Ptr(),
			})
		} else {
			arity = nodes.Optional
		}
	}

	fieldType = &nodes.FieldType{Name: typ, Arity: arity, Span: ksl.RangeBetween(typTok.Range, p.PrevRange())}

	tok, next = p.Peek2()
	if tok.Type == lex.TokenLBrack && next.Type == lex.TokenRBrack {
		p.ReadN(2)
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagInvalidFieldDefinition,
			Detail:   "Unexpected brackets.",
			Subject:  tok.Range.Ptr(),
		})
	}

Token:
	for {
		switch tok, next := p.Peek2(); {
		case tok.Type == lex.TokenNewline:
			p.Read()
		case tok.Type == lex.TokenAt && next.Type != lex.TokenAt:
			annot, annotDiags := p.parseFieldAnnotation()
			diags = append(diags, annotDiags...)
			if annot != nil {
				annotations = append(annotations, annot)
			}
		default:
			break Token
		}
	}

	return &nodes.Field{
		Name:        name,
		Type:        fieldType,
		Comment:     doc,
		Annotations: annotations,
		Span:        ksl.RangeBetween(nameTok.Range, p.PrevRange()),
	}, diags
}

func (p *parser) parseBlock(doc *nodes.CommentGroup) (*nodes.Block, ksl.Diagnostics) {
	var diags ksl.Diagnostics
	var name *nodes.Name
	var blkType *nodes.TypeName
	var props nodes.Properties

	typeTok := p.Peek()
	if typeTok.Type != lex.TokenIdent {
		return nil, append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagExpectedBlockDefinition,
			Detail:   "This line is invalid. A block definition was expected here.",
			Subject:  typeTok.Range.Ptr(),
		})
	}
	p.Read()
	blkType = &nodes.TypeName{Name: typeTok.Value, Span: typeTok.Range}

	nameTok := p.Peek()
	if !nameTok.IsAny(lex.TokenIdent, lex.TokenQualifiedIdent) {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagExpectedBlockDefinition,
			Detail:   "Expected a block name here.",
			Subject:  nameTok.Range.Ptr(),
		})
		p.recoverTo(lex.TokenNewline, lex.TokenLBrace)
	} else {
		p.Read()
		name = &nodes.Name{Name: nameTok.Value, Span: nameTok.Range}
	}

	next := p.Peek()
	if next.Type == lex.TokenLBrace {
		p.Read()

	Token:
		for {
			next := p.Peek()
			switch next.Type {
			case lex.TokenNewline:
				p.Read()
				continue
			case lex.TokenIdent:
				prop, propDiags := p.parseProperty()
				props = append(props, prop)
				diags = append(diags, propDiags...)
			default:
				break Token
			}
		}
	} else {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagInvalidBlockDefinition,
			Detail:   "An block must have a body.",
			Subject:  next.Range.Ptr(),
		})
	}

	endTok := p.Peek()
	if endTok.Type != lex.TokenRBrace {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagExpectedBlockClose,
			Detail:   "This line is invalid. A closing brace was expected here.",
			Subject:  typeTok.Range.Ptr(),
		})
		endTok = p.recover(lex.TokenRBrace)
	} else {
		p.Read()
	}

	return &nodes.Block{
		Type:       blkType,
		Name:       name,
		Properties: props,
		Comment:    doc,
		Span:       ksl.RangeBetween(typeTok.Range, p.PrevRange()),
	}, diags
}

func (p *parser) parseEnum(doc *nodes.CommentGroup) (*nodes.Enum, ksl.Diagnostics) {
	var diags ksl.Diagnostics
	var name *nodes.Name
	var values nodes.EnumValues
	var annotations nodes.Annotations
	var pendingComment *nodes.CommentGroup
	cancelPendingComment := func() {
		if pendingComment != nil {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagWarning,
				Summary:  DiagFloatingDocComment,
				Detail:   "This doc comment is floating and isn't attached to any node.",
				Subject:  pendingComment.Span.Ptr(),
			})
		}
		pendingComment = nil
	}

	typeTok := p.Peek()
	if typeTok.Type != lex.TokenEnum {
		return nil, append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagExpectedBlockDefinition,
			Detail:   "This line is invalid. An enum definition was expected here.",
			Subject:  typeTok.Range.Ptr(),
		})
	}
	p.Read()

	p.PushIncludeNewlines(true)
	defer p.PopIncludeNewlines()

	nameTok := p.Peek()
	if !nameTok.IsAny(lex.TokenIdent, lex.TokenQualifiedIdent) {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagInvalidEnumName,
			Detail:   "An enum must have a name.",
			Subject:  nameTok.Range.Ptr(),
		})
		p.recoverTo(lex.TokenNewline, lex.TokenLBrace)
	} else {
		p.Read()
		name = &nodes.Name{Name: nameTok.Value, Span: nameTok.Range}
	}

	next := p.Peek()
	if next.Type == lex.TokenLBrace {
		p.Read()

	Token:
		for {
			first := p.Peek()

			switch first.Type {
			case lex.TokenDocComment:
				pendingComment = p.parseComments()
				continue

			case lex.TokenRBrace:
				p.Read()
				cancelPendingComment()
				break Token

			case lex.TokenNewline:
				p.Read()
				cancelPendingComment()
				continue

			case lex.TokenAt:
				annot, annotDiags := p.parseBlockAnnotation()
				annotations = append(annotations, annot)
				diags = append(diags, annotDiags...)
				if p.recovery && annotDiags.HasErrors() {
					p.recover(lex.TokenNewline)
				}
				cancelPendingComment()

			case lex.TokenIdent:
				val, valDiags := p.parseEnumValue(pendingComment)
				values = append(values, val)
				diags = append(diags, valDiags...)
				if p.recovery && valDiags.HasErrors() {
					p.recover(lex.TokenNewline)
				}
				pendingComment = nil

			default:
				if !p.recovery {
					diags = append(diags, &ksl.Diagnostic{
						Severity: ksl.DiagError,
						Summary:  DiagUnexpectedToken,
						Detail:   "Expected an enum value or a block annotation.",
						Subject:  &first.Range,
					})
				}
				p.recover(lex.TokenRBrace)
				cancelPendingComment()
				break Token
			}
		}
	} else {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagInvalidEnumDefinition,
			Detail:   "An enum must have a body.",
			Subject:  next.Range.Ptr(),
		})
	}

	return &nodes.Enum{
		Name:        name,
		Values:      values,
		Annotations: annotations,
		Comment:     doc,
		Span:        ksl.RangeBetween(typeTok.Range, p.PrevRange()),
	}, diags
}

func (p *parser) parseEnumValue(doc *nodes.CommentGroup) (*nodes.EnumValue, ksl.Diagnostics) {
	var diags ksl.Diagnostics
	var name *nodes.Name
	var annotations nodes.Annotations

	nameTok := p.Peek()
	if !nameTok.IsAny(lex.TokenIdent, lex.TokenQualifiedIdent) {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagInvalidEnumValue,
			Detail:   "An enum value must be an identifier.",
			Subject:  nameTok.Range.Ptr(),
		})
	} else {
		p.Read()
		name = &nodes.Name{Name: nameTok.Value, Span: nameTok.Range}
	}

Token:
	for {
		switch tok, next := p.Peek2(); {
		case tok.Type == lex.TokenNewline:
			p.Read()
		case tok.Type == lex.TokenAt && next.Type != lex.TokenAt:
			annot, annotDiags := p.parseFieldAnnotation()
			diags = append(diags, annotDiags...)
			if annot != nil {
				annotations = append(annotations, annot)
			}
		default:
			break Token
		}
	}

	return &nodes.EnumValue{
		Name:        name,
		Comment:     doc,
		Annotations: annotations,
		Span:        ksl.RangeBetween(nameTok.Range, p.PrevRange()),
	}, diags
}

func (p *parser) parseBlockAnnotation() (*nodes.Annotation, ksl.Diagnostics) {
	var diags ksl.Diagnostics
	var name *nodes.Name
	var args *nodes.ArgumentList

	atTok1 := p.Peek()
	if atTok1.Type != lex.TokenAt {
		return nil, ksl.Diagnostics{{
			Severity: ksl.DiagError,
			Summary:  DiagUnexpectedToken,
			Detail:   "Expected an atmark here.",
			Subject:  &atTok1.Range,
		}}
	}
	p.Read()

	atTok2 := p.Peek()
	if atTok2.Type != lex.TokenAt {
		return nil, ksl.Diagnostics{{
			Severity: ksl.DiagError,
			Summary:  DiagUnexpectedToken,
			Detail:   "Expected an atmark here.",
			Subject:  &atTok2.Range,
		}}
	}
	p.Read()

	nameTok := p.Peek()
	if !nameTok.IsAny(lex.TokenIdent, lex.TokenQualifiedIdent) {
		return nil, ksl.Diagnostics{{
			Severity: ksl.DiagError,
			Summary:  DiagInvalidFunctionName,
			Detail:   "A block annotation name is required here.",
			Subject:  &nameTok.Range,
		}}
	}
	p.Read()
	name = &nodes.Name{Name: nameTok.Value, Span: nameTok.Range}

	tok := p.Peek()
	if tok.Type == lex.TokenLParen {
		var argDiags ksl.Diagnostics
		args, argDiags = p.parseArgumentList()
		diags = append(diags, argDiags...)
	}

	return &nodes.Annotation{
		Name: name,
		Args: args,
		Span: ksl.RangeBetween(atTok1.Range, p.PrevRange()),
	}, diags
}

func (p *parser) parseFieldAnnotation() (*nodes.Annotation, ksl.Diagnostics) {
	var diags ksl.Diagnostics
	var name *nodes.Name
	var args *nodes.ArgumentList

	atTok := p.Peek()
	if atTok.Type != lex.TokenAt {
		return nil, ksl.Diagnostics{{
			Severity: ksl.DiagError,
			Summary:  DiagUnexpectedToken,
			Detail:   "Expected an opening parenthesis here.",
			Subject:  &atTok.Range,
		}}
	}
	p.Read()

	nameTok := p.Peek()
	if !nameTok.IsAny(lex.TokenIdent, lex.TokenQualifiedIdent) {
		return nil, ksl.Diagnostics{{
			Severity: ksl.DiagError,
			Summary:  DiagInvalidFunctionName,
			Detail:   "An annotation name is required here.",
			Subject:  &nameTok.Range,
		}}
	}
	p.Read()
	name = &nodes.Name{Name: nameTok.Value, Span: nameTok.Range}

	tok := p.Peek()
	if tok.Type == lex.TokenLParen {
		var argDiags ksl.Diagnostics
		args, argDiags = p.parseArgumentList()
		diags = append(diags, argDiags...)
	}

	return &nodes.Annotation{
		Name: name,
		Args: args,
		Span: ksl.RangeBetween(atTok.Range, p.PrevRange()),
	}, diags
}

func (p *parser) parseArgumentList() (*nodes.ArgumentList, ksl.Diagnostics) {
	var diags ksl.Diagnostics
	var args nodes.Arguments
	var closeTok lex.Token

	openTok := p.Peek()
	if openTok.Type != lex.TokenLParen {
		return nil, ksl.Diagnostics{{
			Severity: ksl.DiagError,
			Summary:  DiagUnexpectedToken,
			Detail:   "Expected an opening parenthesis here.",
			Subject:  &openTok.Range,
		}}
	}
	p.Read()

	p.PushIncludeNewlines(false)
	defer p.PopIncludeNewlines()

Token:
	for {
		next := p.Peek()
		if next.Type == lex.TokenRParen {
			closeTok = p.Read()
			break Token
		}

		arg, argDiags := p.parseArgument()
		diags = append(diags, argDiags...)

		if p.recovery && argDiags.HasErrors() {
			closeTok = p.recover(lex.TokenRParen)
			break Token
		}
		args = append(args, arg)

		sep := p.Read()
		if sep.Type == lex.TokenRParen {
			closeTok = sep
			break Token
		}

		if sep.Type != lex.TokenComma {
			switch sep.Type {
			case lex.TokenEOF:
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  DiagUnterminatedList,
					Detail:   "There is no closing parenthesis for this argument list before the end of the file.",
					Subject:  ksl.RangeBetween(openTok.Range, sep.Range).Ptr(),
				})
			default:
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  DiagMissingSeparator,
					Detail:   "A comma is required to separate each argument from the next.",
					Subject:  &sep.Range,
					Context:  ksl.RangeBetween(openTok.Range, sep.Range).Ptr(),
				})
			}
			closeTok = p.recover(lex.TokenRParen)
			break Token
		}

		if p.Peek().Type == lex.TokenRParen {
			closeTok = p.Read()
			break Token
		}
	}

	return &nodes.ArgumentList{
		Arguments: args,
		Span:      ksl.RangeBetween(openTok.Range, closeTok.Range),
	}, diags
}

func (p *parser) parseArgument() (*nodes.Argument, ksl.Diagnostics) {
	var diags ksl.Diagnostics
	var name *nodes.Name
	var value nodes.Expression

	startRange := p.NextRange()

	first, second := p.Peek2()
	if first.Type == lex.TokenIdent && second.Type == lex.TokenColon {
		name = &nodes.Name{Name: first.Value, Span: first.Range}
		p.ReadN(2)
	}

	var exprDiags ksl.Diagnostics
	value, exprDiags = p.parseExpression()
	diags = append(diags, exprDiags...)

	return &nodes.Argument{
		Name:  name,
		Value: value,
		Span:  ksl.RangeBetween(startRange, p.PrevRange()),
	}, diags
}

func (p *parser) parseExpression() (nodes.Expression, ksl.Diagnostics) {
	first, second := p.Peek2()

	switch first.Type {
	case lex.TokenIdent, lex.TokenQualifiedIdent:
		switch second.Type {
		case lex.TokenLParen:
			return p.parseFunction()
		case lex.TokenEqual:
			return nil, ksl.Diagnostics{
				{
					Severity: ksl.DiagError,
					Summary:  DiagInvalidExpression,
					Detail:   "Expected an expression here, but found a key-value pair.",
					Subject:  &first.Range,
				},
			}
		default:
			tok := p.Read()
			return &nodes.Literal{Value: tok.Value, Span: tok.Range}, nil
		}
	case lex.TokenStringLit:
		p.Read()
		val, diags := ParseStringLiteralToken(first)
		return &nodes.Literal{Value: val, Span: first.Range}, diags
	case lex.TokenQuotedLit:
		tok := p.Read()
		val, diags := ParseStringLiteralToken(tok)
		return &nodes.String{Value: val[1 : len(val)-1], Span: tok.Range}, diags
	case lex.TokenHeredocBegin:
		return p.parseHeredoc()
	case lex.TokenNumberLit, lex.TokenIntegerLit, lex.TokenFloatLit:
		return p.parseNumber()
	case lex.TokenBoolLit:
		p.Read()
		return &nodes.Literal{Value: first.Value, Span: first.Range}, nil
	case lex.TokenNullLit:
		p.Read()
		return &nodes.Literal{Value: first.Value, Span: first.Range}, nil
	case lex.TokenLBrack:
		return p.parseList()
	case lex.TokenDollar:
		return p.parseVariable()
	case lex.TokenLBrace:
		return p.parseObject()
	default:
		return nil, ksl.Diagnostics{{
			Severity: ksl.DiagError,
			Summary:  DiagInvalidExpression,
			Detail:   "Expected an expression here.",
			Subject:  &first.Range,
		}}
	}
}

func (p *parser) parseObject() (*nodes.Object, ksl.Diagnostics) {
	var diags ksl.Diagnostics
	var props nodes.Properties
	var closeTok lex.Token

	openTok := p.Read()
	if openTok.Type != lex.TokenLBrace {
		return nil, ksl.Diagnostics{
			{
				Severity: ksl.DiagError,
				Summary:  DiagUnexpectedToken,
				Detail:   "Expected a left curly brace here.",
				Subject:  &openTok.Range,
			},
		}
	}

	p.PushIncludeNewlines(false)
	defer p.PopIncludeNewlines()

Token:
	for {
		next := p.Peek()
		if next.Type == lex.TokenRBrace {
			closeTok = p.Read()
			break Token
		}

		prop, propDiags := p.parseProperty()
		diags = append(diags, propDiags...)
		if p.recovery && propDiags.HasErrors() {
			closeTok = p.recover(lex.TokenRBrace)
			break Token
		}
		props = append(props, prop)

		sep := p.Read()
		if sep.Type == lex.TokenRBrace {
			closeTok = sep
			break Token
		}

		if sep.Type != lex.TokenNewline && sep.Type != lex.TokenComma {
			switch sep.Type {
			case lex.TokenEOF:
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  DiagUnterminatedObject,
					Detail:   "There is no closing brace for this object before the end of the file.",
					Subject:  ksl.RangeBetween(openTok.Range, sep.Range).Ptr(),
				})
			default:
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  DiagMissingSeparator,
					Detail:   "Each object property must be on a separate line.",
					Subject:  &sep.Range,
					Context:  ksl.RangeBetween(openTok.Range, sep.Range).Ptr(),
				})
			}
			closeTok = p.recover(lex.TokenRBrace)
			break Token
		}

		if p.Peek().Type == lex.TokenRBrace {
			closeTok = p.Read()
			break Token
		}
	}

	return &nodes.Object{
		Properties: props,
		Span:       ksl.RangeBetween(openTok.Range, closeTok.Range),
	}, diags
}

func (p *parser) parseProperty() (*nodes.Property, ksl.Diagnostics) {
	var diags ksl.Diagnostics
	var name *nodes.Name
	var value nodes.Expression

	p.PushIncludeNewlines(true)
	defer p.PopIncludeNewlines()

	var endRange ksl.Range

	nameTok, eqTok := p.Peek2()
	if nameTok.Type != lex.TokenIdent {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagInvalidPropertyName,
			Detail:   "Expected a property name here.",
			Subject:  &nameTok.Range,
		})
		p.recover(lex.TokenNewline)
		return nil, diags
	}
	p.Read()
	name = &nodes.Name{Name: nameTok.Value, Span: nameTok.Range}

	if eqTok.Type != lex.TokenEqual {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagInvalidPropertyAssignment,
			Detail:   "Expected a property assignment here.",
			Subject:  &eqTok.Range,
		})
		endRange = nameTok.Range
		p.recover(lex.TokenNewline)
	} else {
		p.Read()
		var valueDiags ksl.Diagnostics
		value, valueDiags = p.parseExpression()
		diags = append(diags, valueDiags...)
		endRange = p.PrevRange()
	}

	return &nodes.Property{
		Name:  name,
		Value: value,
		Span:  ksl.RangeBetween(nameTok.Range, endRange),
	}, diags
}

func (p *parser) parseList() (*nodes.List, ksl.Diagnostics) {
	var diags ksl.Diagnostics
	var elems []nodes.Expression
	var closeTok lex.Token

	openTok := p.Read()
	if openTok.Type != lex.TokenLBrack {
		return nil, ksl.Diagnostics{{
			Severity: ksl.DiagError,
			Summary:  DiagUnexpectedToken,
			Detail:   "An left bracket was expected here.",
			Subject:  &openTok.Range,
		}}
	}

	p.PushIncludeNewlines(false)
	defer p.PopIncludeNewlines()

	for {
		next := p.Peek()
		if next.Type == lex.TokenRBrack {
			closeTok = p.Read()
			break
		}

		val, valueDiags := p.parseExpression()
		elems = append(elems, val)
		diags = append(diags, valueDiags...)

		if p.recovery && valueDiags.HasErrors() {
			closeTok = p.recover(lex.TokenRBrack)
			break
		}

		next = p.Peek()
		if next.Type == lex.TokenRBrack {
			closeTok = p.Read()
			break
		}
		if next.Type != lex.TokenComma {
			if !p.recovery {
				switch next.Type {
				case lex.TokenEOF:
					diags = append(diags, &ksl.Diagnostic{
						Severity: ksl.DiagError,
						Summary:  DiagUnterminatedList,
						Detail:   "There is no corresponding closing bracket before the end of the file. This may be caused by incorrect bracket nesting elsewhere in this file.",
						Subject:  openTok.Range.Ptr(),
					})
				default:
					diags = append(diags, &ksl.Diagnostic{
						Severity: ksl.DiagError,
						Summary:  DiagMissingSeparator,
						Detail:   "Expected a comma to mark the beginning of the next item.",
						Subject:  &next.Range,
						Context:  ksl.RangeBetween(openTok.Range, next.Range).Ptr(),
					})
				}
			}
			closeTok = p.recover(lex.TokenRBrack)
			break
		}

		p.Read()
	}

	return &nodes.List{
		Elements: elems,
		Span:     ksl.RangeBetween(openTok.Range, closeTok.Range),
	}, diags
}

func (p *parser) parseNumber() (*nodes.Number, ksl.Diagnostics) {
	var diags ksl.Diagnostics
	var val string

	tok := p.Peek()
	if !tok.IsAny(lex.TokenNumberLit, lex.TokenFloatLit, lex.TokenIntegerLit) {
		return nil, ksl.Diagnostics{{
			Severity: ksl.DiagError,
			Summary:  DiagUnexpectedToken,
			Detail:   "A number was expected here.",
			Subject:  &tok.Range,
		}}
	}
	p.Read()
	val = tok.Value

	return &nodes.Number{
		Value: val,
		Span:  tok.Range,
	}, diags
}

func (p *parser) parseVariable() (*nodes.Variable, ksl.Diagnostics) {
	var diags ksl.Diagnostics

	dollarTok := p.Read()
	if dollarTok.Type != lex.TokenDollar {
		return nil, ksl.Diagnostics{{
			Severity: ksl.DiagError,
			Summary:  DiagUnexpectedToken,
			Detail:   "Expected a variable reference here.",
			Subject:  &dollarTok.Range,
		}}
	}

	ident := p.Peek()
	if !ident.IsAny(lex.TokenIdent, lex.TokenQualifiedIdent) {
		return nil, ksl.Diagnostics{{
			Severity: ksl.DiagError,
			Summary:  "Invalid variable reference",
			Detail:   "A valid identifier is required here.",
			Subject:  &ident.Range,
		}}
	}

	p.Read()

	return &nodes.Variable{
		Name: &nodes.Name{Name: ident.Value, Span: ident.Range},
		Span: ksl.RangeBetween(dollarTok.Range, ident.Range),
	}, diags
}

func (p *parser) parseFunction() (*nodes.Function, ksl.Diagnostics) {
	var diags ksl.Diagnostics
	var name *nodes.Name
	var args *nodes.ArgumentList

	nameTok := p.Read()
	if !nameTok.IsAny(lex.TokenIdent, lex.TokenQualifiedIdent) {
		return nil, ksl.Diagnostics{{
			Severity: ksl.DiagError,
			Summary:  DiagInvalidFunctionName,
			Detail:   "A function name is required here.",
			Subject:  &nameTok.Range,
		}}
	}
	name = &nodes.Name{Name: nameTok.Value, Span: nameTok.Range}

	tok := p.Peek()
	if tok.Type != lex.TokenLParen {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagMissingParenthesis,
			Detail:   "A function call must contain parenthesis.",
			Subject:  &tok.Range,
		})
	} else {
		var argDiags ksl.Diagnostics
		args, argDiags = p.parseArgumentList()
		diags = append(diags, argDiags...)
	}

	return &nodes.Function{
		Name:      name,
		Arguments: args,
		Span:      ksl.RangeBetween(nameTok.Range, p.PrevRange()),
	}, diags
}

func (p *parser) parseHeredoc() (*nodes.Heredoc, ksl.Diagnostics) {
	var diags ksl.Diagnostics
	var lines []string
	var marker string
	var strip bool
	var closeTok lex.Token

	beginTok := p.Read()
	if beginTok.Type != lex.TokenHeredocBegin {
		return nil, ksl.Diagnostics{{
			Severity: ksl.DiagError,
			Summary:  DiagUnexpectedToken,
			Detail:   "Expected a heredoc here.",
			Subject:  &beginTok.Range,
		}}
	}

	p.PushIncludeNewlines(false)
	defer p.PopIncludeNewlines()

	marker = strings.TrimSpace(beginTok.Value)
	marker = marker[2:]
	if marker[0] == '-' {
		strip = true
		marker = marker[1:]
	}

Token:
	for {
		switch next := p.Peek(); next.Type {
		case lex.TokenStringLit:
			p.Read()
			lines = append(lines, next.Value)
		case lex.TokenHeredocEnd:
			closeTok = p.Read()
			if len(lines) == 0 {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  "Empty heredoc",
					Detail:   "Heredocs must contain at least one line of text.",
					Subject:  &next.Range,
				})
			}
			break Token
		default:
			closeTok = p.recover(lex.TokenHeredocEnd)
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Invalid heredoc",
				Detail:   "Expected heredoc marker.",
				Subject:  &next.Range,
			})
			break Token
		}
	}

	return &nodes.Heredoc{
		Marker:      marker,
		Values:      lines,
		StripIndent: strip,
		Span:        ksl.RangeBetween(beginTok.Range, closeTok.Range),
	}, diags
}

func (p *parser) parseComments() *nodes.CommentGroup {
	var comments []*nodes.Comment

	startRange := p.NextRange()
	for {
		next := p.Peek()
		if next.Type != lex.TokenDocComment {
			break
		}
		comments = append(comments, &nodes.Comment{Text: next.Value, Span: next.Range})
		p.Read()
	}

	return &nodes.CommentGroup{Comments: comments, Span: ksl.RangeBetween(startRange, p.PrevRange())}
}

func (p *parser) recoverTo(end ...lex.TokenType) lex.Token {
	p.recovery = true
	for {
		tok := p.Peek()
		if slices.Contains(end, tok.Type) || tok.Type == lex.TokenEOF {
			return tok
		}
		p.Read()
	}
}

func (p *parser) recover(end lex.TokenType) lex.Token {
	start := p.oppositeBracket(end)
	p.recovery = true

	nest := 0
	for {
		tok := p.Read()
		ty := tok.Type
		switch ty {
		case start:
			nest++
		case end:
			if nest < 1 {
				return tok
			}

			nest--
		case lex.TokenEOF:
			return tok
		}
	}
}

func (p *parser) oppositeBracket(ty lex.TokenType) lex.TokenType {
	switch ty {

	case lex.TokenLBrace:
		return lex.TokenRBrace
	case lex.TokenLBrack:
		return lex.TokenRBrack
	case lex.TokenLParen:
		return lex.TokenRParen
	case lex.TokenHeredocBegin:
		return lex.TokenHeredocEnd

	case lex.TokenRBrace:
		return lex.TokenLBrace
	case lex.TokenRBrack:
		return lex.TokenLBrack
	case lex.TokenRParen:
		return lex.TokenLParen
	case lex.TokenHeredocEnd:
		return lex.TokenHeredocBegin

	default:
		return lex.TokenNil
	}
}

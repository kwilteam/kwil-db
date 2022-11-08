package kslsyntax

import (
	"math/big"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
	"ksl"
	"ksl/kslsyntax/ast"
	"ksl/kslsyntax/lex"
)

type parser struct {
	*peeker
	recovery bool
}

func (p *parser) parseDocument() (*ast.Document, ksl.Diagnostics) {
	directives := ast.Directives{}
	blocks := ast.Blocks{}
	var diags ksl.Diagnostics

	startRange := p.PrevRange()
	var endRange ksl.Range

Token:
	for {
		next := p.Peek()

		switch next.Type {
		case lex.TokenNewline:
			p.Read()
			continue

		case lex.TokenEOF:
			endRange = p.NextRange()
			p.Read()
			break Token

		case lex.TokenIdent:
			blk, blockDiags := p.parseBlock()
			diags = append(diags, blockDiags...)
			blocks = append(blocks, blk)

		case lex.TokenAt:
			directive, directiveDiags := p.parseDirective()
			diags = append(diags, directiveDiags...)
			directives = append(directives, directive)

		default:
			bad := p.Read()
			if !p.recovery {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  DiagExpectedDirectiveOrBlock,
					Detail:   "A directive or block definition is required here.",
					Subject:  &bad.Range,
				})
			}

			endRange = p.PrevRange()
			p.recover(lex.TokenEOF)
			break Token
		}
	}

	return &ast.Document{
		Directives: directives,
		Blocks:     blocks,
		SrcRange:   ksl.RangeBetween(startRange, endRange),
	}, diags
}

func (p *parser) parseDirective() (*ast.Directive, ksl.Diagnostics) {
	at := p.Read()
	if at.Type != lex.TokenAt {
		return &ast.Directive{SrcRange: ksl.Range{Start: at.Range.Start, End: at.Range.Start}}, ksl.Diagnostics{
			&ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagExpectedDirective,
				Detail:   "A directive is required here.",
				Subject:  &at.Range,
			},
		}
	}

	typ := p.Peek()
	if typ.Type != lex.TokenIdent {
		p.recover(lex.TokenNewline)
		return nil, ksl.Diagnostics{
			{
				Severity: ksl.DiagError,
				Summary:  DiagInvalidDirectiveName,
				Detail:   "A directive name is required here.",
				Subject:  &typ.Range,
			},
		}
	}
	p.Read()

	var diags ksl.Diagnostics
	var endRange ksl.Range
	var key *ast.Str
	var value ast.Expr

	if first, second := p.Peek2(); first.IsAny(lex.TokenIdent, lex.TokenQualifiedIdent) && second.Type == lex.TokenEqual {
		p.ReadN(2)
		key = &ast.Str{Value: first.Value, SrcRange: first.Range}
	}

	next := p.Peek()

	if next.Type == lex.TokenEqual {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagNoKeyBeforeAssignment,
			Detail:   "A directive must have a key before an assignment.",
			Subject:  &next.Range,
		})
		p.Read()
		next = p.Peek()
	}

	if next.Type == lex.TokenNewline || next.Type == lex.TokenEOF {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagMissingDirectiveValue,
			Detail:   "A directive must have a value.",
			Subject:  &next.Range,
		})
	} else {
		val, valueDiags := p.parseExpression()
		diags = append(diags, valueDiags...)
		value = val
	}

	endRange = p.PrevRange()
	if p.recovery && diags.HasErrors() {
		p.recover(lex.TokenNewline)
	} else if tok := p.Peek(); tok.Type != lex.TokenNewline && tok.Type != lex.TokenEOF {
		if !p.recovery {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagMissingNewlineAfterDirective,
				Detail:   "A newline is required after a directive.",
				Subject:  &tok.Range,
			})
		}
	} else {
		p.Read()
	}

	return &ast.Directive{
		At:    at,
		Name:  &ast.Str{Value: string(typ.Value), SrcRange: typ.Range},
		Key:   key,
		Value: value,

		SrcRange: ksl.RangeBetween(at.Range, endRange),
	}, diags
}

func (p *parser) parseBlock() (*ast.Block, ksl.Diagnostics) {
	var typeNode, nameNode, keywordNode, targetNode *ast.Str
	var labels *ast.BlockLabels
	var body *ast.Body

	var diags ksl.Diagnostics

	typ := p.Peek()
	if typ.Type != lex.TokenIdent {
		return &ast.Block{SrcRange: ksl.Range{Start: typ.Range.Start, End: typ.Range.Start}}, ksl.Diagnostics{
			{
				Severity: ksl.DiagError,
				Summary:  DiagExpectedBlockDefinition,
				Detail:   "A block definition was expected here.",
				Subject:  &typ.Range,
			},
		}
	}
	p.Read()
	typeNode = &ast.Str{Value: typ.Value, SrcRange: typ.Range}

	next := p.Peek()
	if next.IsAny(lex.TokenIdent, lex.TokenQualifiedIdent) {
		nameNode = &ast.Str{Value: next.Value, SrcRange: next.Range}
		p.Read()
		next = p.Peek()

		if next.Type == lex.TokenIdent {
			keywordNode = &ast.Str{Value: next.Value, SrcRange: next.Range}
			p.Read()
			next = p.Peek()

			if next.IsAny(lex.TokenIdent, lex.TokenQualifiedIdent) {
				targetNode = &ast.Str{Value: next.Value, SrcRange: next.Range}
				p.Read()
				next = p.Peek()
			} else {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  DiagBlockInvalidModifierTarget,
					Detail:   "Expected a modifier target here. A modifier target must be an identifier.",
					Subject:  &next.Range,
				})
				next = p.recoverTo(lex.TokenNewline, lex.TokenLBrack, lex.TokenLBrace)
			}
		}
	}

	if next.Type == lex.TokenLBrack {
		var labelDiags ksl.Diagnostics
		labels, labelDiags = p.parseLabels()
		diags = append(diags, labelDiags...)
		next = p.Peek()
	}

	if next.Type == lex.TokenNewline {
		if !p.recovery {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagInvalidBlockDefinition,
				Detail:   "A block definition must be on a single line.",
				Subject:  &next.Range,
			})
		}
		p.Read()
		next = p.Peek()
	}

	if next.Type != lex.TokenLBrace {
		if !p.recovery {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagInvalidBlockDefinition,
				Detail:   "A block definition must be followed by a body.",
				Subject:  &next.Range,
			})
		}
	} else {
		var bodyDiags ksl.Diagnostics
		body, bodyDiags = p.parseBlockBody()
		diags = append(diags, bodyDiags...)
	}

	eol := p.Peek()
	if eol.Type == lex.TokenNewline || eol.Type == lex.TokenEOF {
		p.Read()
	} else {
		if !p.recovery {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagMissingNewlineAfterBlock,
				Detail:   "A block definition must end with a newline.",
				Subject:  &eol.Range,
				Context:  ksl.RangeBetween(typ.Range, eol.Range).Ptr(),
			})
		}
	}

	// We must never produce a nil body, because that would be a panic
	if body == nil && diags.HasErrors() {
		body = &ast.Body{
			SrcRange: ksl.RangeBetween(next.Range, p.PrevRange()),
		}
	}

	return &ast.Block{
		Type:     typeNode,
		Name:     nameNode,
		Modifier: keywordNode,
		Target:   targetNode,
		Labels:   labels,
		Body:     body,
		SrcRange: ksl.RangeBetween(typ.Range, p.PrevRange()),
	}, diags
}

func (p *parser) parseBlockBody() (*ast.Body, ksl.Diagnostics) {
	annots := ast.Annotations{}
	blocks := ast.Blocks{}
	props := ast.Attributes{}
	enums := []*ast.Str{}
	decls := ast.Definitions{}

	var diags ksl.Diagnostics
	var closeTok lex.Token

	openTok := p.Read()
	if openTok.Type != lex.TokenLBrace {
		return &ast.Body{SrcRange: ksl.Range{Start: openTok.Range.Start, End: openTok.Range.Start}}, ksl.Diagnostics{
			{
				Severity: ksl.DiagError,
				Summary:  DiagInvalidBlockBody,
				Detail:   "A block body must start with a left curly brace.",
				Subject:  &openTok.Range,
			},
		}
	}

	p.PushIncludeNewlines(true)
	defer p.PopIncludeNewlines()

Token:
	for {
		first, second := p.Peek2()

		switch first.Type {
		case lex.TokenRBrace:
			closeTok = p.Read()
			break Token
		case lex.TokenNewline:
			p.Read()
			continue
		case lex.TokenAt:
			annot, annotDiags := p.parseBlockAnnotation()
			annots = append(annots, annot)
			diags = append(diags, annotDiags...)
		case lex.TokenIdent:
			switch second.Type {
			case lex.TokenEqual:
				val, valueDiags := p.parseKeyValue()
				diags = append(diags, valueDiags...)
				props = append(props, val)
			case lex.TokenColon:
				decl, declDiags := p.parseDeclaration()
				decls = append(decls, decl)
				diags = append(diags, declDiags...)
			case lex.TokenNewline:
				p.Read()
				enums = append(enums, &ast.Str{Value: first.Value, SrcRange: first.Range})
			default:
				block, blockDiags := p.parseBlock()
				blocks = append(blocks, block)
				diags = append(diags, blockDiags...)

				if blockDiags.HasErrors() {
					p.recoverTo(lex.TokenNewline, lex.TokenRBrace)
				}
				// diags = append(diags, &ksl.Diagnostic{
				// 	Severity: ksl.DiagError,
				// 	Summary:  DiagUnexpectedToken,
				// 	Detail:   "Expected ':', '=', '{', or new line.",
				// 	Subject:  &second.Range,
				// })
				// closeTok = p.recover(lex.TokenRBrace)
				// break Token
			}
		default:
			if !p.recovery {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  DiagUnexpectedToken,
					Detail:   "Expected an identifier or a block annotation.",
					Subject:  &first.Range,
				})
			}
			closeTok = p.recover(lex.TokenRBrace)
			break Token
		}
	}

	return &ast.Body{
		LBrace:      openTok,
		Annotations: annots,
		Blocks:      blocks,
		Attributes:  props,
		EnumValues:  enums,
		Definitions: decls,
		RBrace:      closeTok,

		SrcRange: ksl.RangeBetween(openTok.Range, closeTok.Range),
	}, diags
}

func (p *parser) parseDeclaration() (*ast.Definition, ksl.Diagnostics) {
	var name *ast.Str
	var typ *ast.Type
	annotations := ast.Annotations{}
	var diags ksl.Diagnostics

	nameTok := p.Read()
	if nameTok.Type != lex.TokenIdent {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagInvalidDeclarationName,
			Detail:   "A declaration name must be an identifier.",
			Subject:  &nameTok.Range,
		})
	} else {
		name = &ast.Str{Value: nameTok.Value, SrcRange: nameTok.Range}
		colonTok := p.Read()
		if colonTok.Type != lex.TokenColon {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagInvalidDeclaration,
				Detail:   "A declaration must have a name followed by a colon.",
				Subject:  &colonTok.Range,
			})
		} else {
			var typDiags ksl.Diagnostics
			typ, typDiags = p.parseType()
			diags = append(diags, typDiags...)

			if !p.recovery && typDiags.HasErrors() {
				p.recover(lex.TokenNewline)
			}

			canHaveBlockAnnotation := false
		Token:
			for {
				first, second := p.Peek2()
				switch first.Type {
				case lex.TokenAt:
					if second.Type == lex.TokenAt {
						// This is a block annotation, so we'll stop here
						if !canHaveBlockAnnotation {
							diags = append(diags, &ksl.Diagnostic{
								Severity: ksl.DiagError,
								Summary:  DiagInvalidAnnotation,
								Detail:   "Block annotations are denoted by '@@' and must be on a new line.",
								Subject:  &first.Range,
							})
						}
						break Token
					}
					annot, annotDiags := p.parseAnnotation()
					annotations = append(annotations, annot)
					diags = append(diags, annotDiags...)
					canHaveBlockAnnotation = false
				case lex.TokenNewline:
					p.Read()
					canHaveBlockAnnotation = true
				default:
					break Token
				}
			}
		}
	}

	return &ast.Definition{
		Name:        name,
		Type:        typ,
		Annotations: annotations,

		SrcRange: ksl.RangeBetween(nameTok.Range, p.PrevRange()),
	}, diags
}

func (p *parser) parseType() (*ast.Type, ksl.Diagnostics) {
	var isArray bool
	var nullable bool
	var typeName *ast.Str
	var diags ksl.Diagnostics

	nameTok := p.Peek()
	switch nameTok.Type {
	case lex.TokenIdent, lex.TokenQualifiedIdent:
		typeName = &ast.Str{Value: nameTok.Value, SrcRange: nameTok.Range}
		p.Read()
	default:
		if !p.recovery {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagInvalidTypeName,
				Detail:   "A type name must be an identifier.",
				Subject:  &nameTok.Range,
			})
		}
		return &ast.Type{IsArray: isArray, Name: typeName, SrcRange: ksl.RangeBetween(nameTok.Range, p.PrevRange())}, diags
	}

	if openBrack := p.Peek(); openBrack.Type == lex.TokenLBrack {
		p.Read()
		switch tok := p.Peek(); tok.Type {
		case lex.TokenRBrack:
			isArray = true
			p.Read()
		case lex.TokenIntegerLit:
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagInvalidTypeSpecifier,
				Detail:   "Array types cannot have a constant size.",
				Subject:  &tok.Range,
			})
			p.Read()
			if p.Peek().Type == lex.TokenRBrack {
				p.Read()
			} else {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  DiagUnexpectedToken,
					Detail:   "Expected a closing bracket.",
					Subject:  &tok.Range,
				})
				p.recover(lex.TokenRBrack)
			}
		default:
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagUnexpectedToken,
				Detail:   "Expected a closing bracket.",
				Subject:  &tok.Range,
			})
			p.recover(lex.TokenRBrack)
		}
	}

	if tok := p.Peek(); tok.Type == lex.TokenQuestion {
		nullable = true
		p.Read()
	}

	return &ast.Type{
		IsArray:  isArray,
		Nullable: nullable,
		Name:     typeName,
		SrcRange: ksl.RangeBetween(nameTok.Range, p.PrevRange()),
	}, diags
}

func (p *parser) parseLabels() (*ast.BlockLabels, ksl.Diagnostics) {
	openTok := p.Read()
	if openTok.Type != lex.TokenLBrack {
		return &ast.BlockLabels{SrcRange: ksl.Range{Start: openTok.Range.Start, End: openTok.Range.Start}}, ksl.Diagnostics{
			{
				Severity: ksl.DiagError,
				Summary:  DiagUnexpectedToken,
				Detail:   "Expected an opening bracket here.",
				Subject:  &openTok.Range,
			},
		}
	}

	p.PushIncludeNewlines(false)
	defer p.PopIncludeNewlines()

	var closeTok lex.Token
	var diags ksl.Diagnostics
	values := []*ast.Attribute{}

Token:
	for {
		tok := p.Peek()
		if tok.Type == lex.TokenRBrack {
			closeTok = p.Read()
			break Token
		}

		if tok.Type != lex.TokenIdent {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagInvalidLabelName,
				Detail:   "Label names must be identifiers.",
				Subject:  &tok.Range,
			})
			closeTok = p.recover(lex.TokenRBrack)
			break Token
		}

		key := &ast.Str{Value: tok.Value, SrcRange: tok.Range}
		var value ast.Expr
		p.Read()

		if tok := p.Peek(); tok.Type == lex.TokenEqual {
			p.Read()
			var valueDiags ksl.Diagnostics
			value, valueDiags = p.parseExpression()
			diags = append(diags, valueDiags...)
		}

		if p.recovery && diags.HasErrors() {
			closeTok = p.recover(lex.TokenRBrack)
			break Token
		}

		label := &ast.Attribute{Name: key, Value: value, SrcRange: ksl.RangeOver(key.Range(), p.PrevRange())}
		values = append(values, label)

		next := p.Peek()
		if next.Type == lex.TokenRBrack {
			closeTok = p.Read()
			break
		}

		if next.Type != lex.TokenComma {
			if !p.recovery {
				switch next.Type {
				case lex.TokenEOF, lex.TokenLBrace:
					diags = append(diags, &ksl.Diagnostic{
						Severity: ksl.DiagError,
						Summary:  DiagUnexpectedToken,
						Detail:   "Expected a closing bracket here.",
						Subject:  openTok.Range.Ptr(),
					})
				default:
					diags = append(diags, &ksl.Diagnostic{
						Severity: ksl.DiagError,
						Summary:  DiagUnexpectedToken,
						Detail:   "Expected a comma to mark the beginning of the next item.",
						Subject:  &next.Range,
						Context:  ksl.RangeBetween(openTok.Range, next.Range).Ptr(),
					})
				}
			}
			closeTok = p.recover(lex.TokenRBrack)
			break Token
		}

		p.Read()
	}

	return &ast.BlockLabels{
		LBrack:   openTok,
		Values:   values,
		RBrack:   closeTok,
		SrcRange: ksl.RangeBetween(openTok.Range, closeTok.Range),
	}, diags
}

func (p *parser) parseAnnotation() (*ast.Annotation, ksl.Diagnostics) {
	at := p.Read()
	if at.Type != lex.TokenAt {
		return &ast.Annotation{SrcRange: ksl.Range{Start: at.Range.Start, End: at.Range.Start}}, ksl.Diagnostics{
			{
				Severity: ksl.DiagError,
				Summary:  DiagInvalidAnnotation,
				Detail:   "Expected an annotation here.",
				Subject:  &at.Range,
			},
		}
	}

	args := []ast.Expr{}
	kwargs := []*ast.Attribute{}
	var diags ksl.Diagnostics
	var name *ast.Str

	if tok := p.Peek(); !tok.IsAny(lex.TokenIdent, lex.TokenQualifiedIdent) {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagInvalidFunctionName,
			Detail:   "Annotation names must be identifiers.",
			Subject:  &tok.Range,
		})
	} else {
		nameTok := p.Read()
		name = &ast.Str{Value: nameTok.Value, SrcRange: nameTok.Range}

		if tok := p.Peek(); tok.Type == lex.TokenLParen {
			argList, argDiags := p.parseArgumentList()
			args, kwargs = argList.Args, argList.Kwargs
			diags = append(diags, argDiags...)
		}
	}

	return &ast.Annotation{
		Marker: &ast.Str{Value: at.Value, SrcRange: at.Range},
		Name:   name,
		ArgList: &ast.ArgList{
			Args:   args,
			Kwargs: kwargs,
		},
		SrcRange: ksl.RangeBetween(at.Range, p.PrevRange()),
	}, diags
}

func (p *parser) parseBlockAnnotation() (*ast.Annotation, ksl.Diagnostics) {
	first, second := p.Peek2()
	if first.Type != lex.TokenAt || second.Type != lex.TokenAt {
		return &ast.Annotation{SrcRange: first.Range}, ksl.Diagnostics{
			{
				Severity: ksl.DiagError,
				Summary:  DiagInvalidAnnotation,
				Detail:   "Expected a block annotation here. A block annotation is denoted by '@@'.",
				Subject:  &first.Range,
			},
		}
	}
	p.Read()

	annot, diags := p.parseAnnotation()
	annot.Marker = &ast.Str{Value: "@@", SrcRange: ksl.RangeBetween(first.Range, second.Range)}
	return annot, diags
}

func (p *parser) parseKeyValue() (*ast.Attribute, ksl.Diagnostics) {
	var diags ksl.Diagnostics
	var key *ast.Str
	var value ast.Expr

	first, second := p.Peek2()
	if !first.IsAny(lex.TokenIdent, lex.TokenQualifiedIdent) || second.Type != lex.TokenEqual {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagInvalidKeyValue,
			Detail:   "Expected a key-value pair here.",
			Subject:  &first.Range,
		})
		return &ast.Attribute{SrcRange: ksl.Range{Start: first.Range.Start, End: first.Range.Start}}, diags
	}

	keyTok := p.Read()
	key = &ast.Str{Value: keyTok.Value, SrcRange: keyTok.Range}

	eqTok := p.Read()

	var valueDiags ksl.Diagnostics
	value, valueDiags = p.parseExpression()
	diags = append(diags, valueDiags...)

	return &ast.Attribute{
		Name:     key,
		Eq:       eqTok,
		Value:    value,
		SrcRange: ksl.RangeOver(key.Range(), p.PrevRange()),
	}, diags
}

func (p *parser) parseExpression() (ast.Expr, ksl.Diagnostics) {
	first, second := p.Peek2()

	switch first.Type {
	case lex.TokenIdent, lex.TokenQualifiedIdent:
		switch second.Type {
		case lex.TokenLParen:
			return p.parseFunctionCall()
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
			return &ast.Str{Value: tok.Value, SrcRange: tok.Range}, nil
		}
	case lex.TokenStringLit:
		tok := p.Read()
		val, diags := ParseStringLiteralToken(tok)
		return &ast.Str{Value: val, SrcRange: tok.Range}, diags
	case lex.TokenQuotedLit:
		tok := p.Read()
		val, diags := ParseStringLiteralToken(tok)
		return &ast.QuotedStr{Char: []rune(val)[0], Value: val[1 : len(val)-1], SrcRange: tok.Range}, diags
	case lex.TokenHeredocBegin:
		return p.parseHeredoc()
	case lex.TokenNumberLit, lex.TokenIntegerLit, lex.TokenFloatLit:
		return p.parseNumberLit()
	case lex.TokenBoolLit:
		p.Read()
		return &ast.Bool{Value: first.Value == "true", SrcRange: first.Range}, nil
	case lex.TokenNullLit:
		p.Read()
		return &ast.Null{SrcRange: first.Range}, nil
	case lex.TokenLBrack:
		return p.parseArray()
	case lex.TokenDollar:
		return p.parseVarRef()
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

func (p *parser) parseObject() (*ast.Object, ksl.Diagnostics) {
	openTok := p.Read()
	if openTok.Type != lex.TokenLBrace {
		return &ast.Object{SrcRange: ksl.Range{Start: openTok.Range.Start, End: openTok.Range.Start}}, ksl.Diagnostics{
			{
				Severity: ksl.DiagError,
				Summary:  DiagUnexpectedToken,
				Detail:   "Expected a left curly brace here.",
				Subject:  &openTok.Range,
			},
		}
	}

	props := ast.Attributes{}
	var diags ksl.Diagnostics
	var closeTok lex.Token

	p.PushIncludeNewlines(false)
	defer p.PopIncludeNewlines()

Token:
	for {
		first, second := p.Peek2()
		if first.Type == lex.TokenRBrace {
			closeTok = p.Read()
			break Token
		}

		if first.Type != lex.TokenIdent {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagInvalidPropertyName,
				Detail:   "Expected a property name here.",
				Subject:  &first.Range,
			})
			closeTok = p.recover(lex.TokenRBrace)
			break Token
		}

		if second.Type != lex.TokenEqual {
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  DiagInvalidPropertyAssignment,
				Detail:   "Expected a property assignment here.",
				Subject:  &second.Range,
			})
			closeTok = p.recover(lex.TokenRBrace)
			break Token
		}

		prop, kvDiags := p.parseKeyValue()
		props = append(props, prop)
		diags = append(diags, kvDiags...)
		if p.recovery && kvDiags.HasErrors() {
			closeTok = p.recover(lex.TokenRBrace)
			break Token
		}

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

	return &ast.Object{
		LBrace:     openTok,
		Attributes: props,
		RBrace:     closeTok,
		SrcRange:   ksl.RangeBetween(openTok.Range, closeTok.Range),
	}, diags
}

func (p *parser) parseNumberLit() (ast.Expr, ksl.Diagnostics) {
	tok := p.Read()
	if !tok.IsAny(lex.TokenNumberLit, lex.TokenIntegerLit, lex.TokenFloatLit) {
		return &ast.Null{}, ksl.Diagnostics{{
			Severity: ksl.DiagError,
			Summary:  DiagExpectedNumberLit,
			Detail:   "Expected a number literal here.",
			Subject:  &tok.Range,
		}}
	}

	switch tok.Type {
	case lex.TokenIntegerLit:
		i, _ := strconv.ParseInt(tok.Value, 0, 64)
		val := &ast.Int{Value: int(i), SrcRange: tok.Range}
		return val, nil
	case lex.TokenFloatLit:
		f, _ := strconv.ParseFloat(tok.Value, 64)
		val := &ast.Float{Value: f, SrcRange: tok.Range}
		return val, nil
	case lex.TokenNumberLit:
		f, _, _ := big.ParseFloat(tok.Value, 10, 512, big.ToNearestEven)
		n := &ast.Number{Value: f, SrcRange: tok.Range}
		return n, nil
	default:
		return nil, ksl.Diagnostics{{
			Severity: ksl.DiagError,
			Summary:  "Invalid number literal",
			Detail:   "A number literal is required here.",
			Subject:  &tok.Range,
		}}
	}
}

func (p *parser) parseVarRef() (*ast.Var, ksl.Diagnostics) {
	tok := p.Read()
	if tok.Type != lex.TokenDollar {
		return &ast.Var{SrcRange: ksl.Range{Start: tok.Range.Start, End: tok.Range.Start}}, ksl.Diagnostics{{
			Severity: ksl.DiagError,
			Summary:  DiagUnexpectedToken,
			Detail:   "Expected a variable reference here.",
			Subject:  &tok.Range,
		}}
	}

	switch next := p.Peek(); next.Type {
	case lex.TokenIdent, lex.TokenQualifiedIdent:
		p.Read()
		return &ast.Var{Dollar: tok, Name: next.Value, SrcRange: ksl.RangeOver(tok.Range, next.Range)}, nil
	default:
		return nil, ksl.Diagnostics{{
			Severity: ksl.DiagError,
			Summary:  "Invalid variable reference",
			Detail:   "A valid identifier is required here.",
			Subject:  &next.Range,
		}}
	}
}

func (p *parser) parseHeredoc() (*ast.Heredoc, ksl.Diagnostics) {
	beginTok := p.Read()
	if beginTok.Type != lex.TokenHeredocBegin {
		return &ast.Heredoc{SrcRange: ksl.Range{Start: beginTok.Range.Start, End: beginTok.Range.Start}}, ksl.Diagnostics{{
			Severity: ksl.DiagError,
			Summary:  DiagUnexpectedToken,
			Detail:   "Expected a heredoc here.",
			Subject:  &beginTok.Range,
		}}
	}

	p.PushIncludeNewlines(false)
	defer p.PopIncludeNewlines()

	var stripIndent bool
	var diags ksl.Diagnostics
	var endTok lex.Token

	marker := strings.TrimSpace(beginTok.Value)
	marker = marker[2:]
	if marker[0] == '-' {
		stripIndent = true
		marker = marker[1:]
	}

	var lines []*ast.Str

Token:
	for {
		switch next := p.Peek(); next.Type {
		case lex.TokenStringLit:
			p.Read()
			lines = append(lines, &ast.Str{Value: next.Value, SrcRange: next.Range})
		case lex.TokenHeredocEnd:
			endTok = p.Read()
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
			endTok = p.recover(lex.TokenHeredocEnd)
			diags = append(diags, &ksl.Diagnostic{
				Severity: ksl.DiagError,
				Summary:  "Invalid heredoc",
				Detail:   "Expected heredoc marker.",
				Subject:  &next.Range,
			})
			break Token
		}
	}

	return &ast.Heredoc{
		Begin:       beginTok,
		Marker:      marker,
		StripIndent: stripIndent,
		Values:      lines,
		End:         endTok,
		SrcRange:    ksl.RangeBetween(beginTok.Range, p.PrevRange()),
	}, diags
}

func (p *parser) parseFunctionCall() (*ast.FunctionCall, ksl.Diagnostics) {
	name := p.Read()
	if !name.IsAny(lex.TokenIdent, lex.TokenQualifiedIdent) {
		return &ast.FunctionCall{SrcRange: ksl.Range{Start: name.Range.Start, End: name.Range.Start}}, ksl.Diagnostics{{
			Severity: ksl.DiagError,
			Summary:  DiagInvalidFunctionName,
			Detail:   "A function name is required here.",
			Subject:  &name.Range,
		}}
	}

	var diags ksl.Diagnostics
	args := []ast.Expr{}
	kwargs := []*ast.Attribute{}

	tok := p.Peek()
	if tok.Type != lex.TokenLParen {
		diags = append(diags, &ksl.Diagnostic{
			Severity: ksl.DiagError,
			Summary:  DiagMissingParenthesis,
			Detail:   "A function call must contain parenthesis.",
			Subject:  &tok.Range,
		})
	} else {
		argList, argDiags := p.parseArgumentList()
		diags = append(diags, argDiags...)
		args, kwargs = argList.Args, argList.Kwargs
	}

	return &ast.FunctionCall{
		Name: &ast.Str{Value: name.Value, SrcRange: name.Range},
		ArgList: &ast.ArgList{
			Args:   args,
			Kwargs: kwargs,
		},
		SrcRange: ksl.RangeOver(name.Range, p.PrevRange()),
	}, diags
}

func (p *parser) parseArgumentList() (*ast.ArgList, ksl.Diagnostics) {
	openTok := p.Read()
	if openTok.Type != lex.TokenLParen {
		return &ast.ArgList{SrcRange: ksl.Range{Start: openTok.Range.Start, End: openTok.Range.Start}}, ksl.Diagnostics{{
			Severity: ksl.DiagError,
			Summary:  DiagUnexpectedToken,
			Detail:   "Expected an opening parenthesis here.",
			Subject:  &openTok.Range,
		}}
	}

	args := []ast.Expr{}
	kwargs := []*ast.Attribute{}
	var diags ksl.Diagnostics
	var closeTok lex.Token

	p.PushIncludeNewlines(false)
	defer p.PopIncludeNewlines()

Token:
	for {
		first, second := p.Peek2()
		if first.Type == lex.TokenRParen {
			closeTok = p.Read()
			break Token
		}

		if first.Type == lex.TokenIdent && second.Type == lex.TokenEqual {
			kv, kvDiags := p.parseKeyValue()
			diags = append(diags, kvDiags...)
			if p.recovery && kvDiags.HasErrors() {
				closeTok = p.recover(lex.TokenRParen)
				break Token
			}
			kwargs = append(kwargs, kv)
		} else {
			expr, exprDiags := p.parseExpression()
			diags = append(diags, exprDiags...)
			if p.recovery && exprDiags.HasErrors() {
				closeTok = p.recover(lex.TokenRParen)
				break Token
			}
			args = append(args, expr)
			if len(kwargs) > 0 {
				diags = append(diags, &ksl.Diagnostic{
					Severity: ksl.DiagError,
					Summary:  DiagInvalidArgument,
					Detail:   "Positional arguments must come before keyword arguments.",
					Subject:  expr.Range().Ptr(),
				})
			}
		}

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

	return &ast.ArgList{
		LParen:   openTok,
		Args:     args,
		Kwargs:   kwargs,
		RParen:   closeTok,
		SrcRange: ksl.RangeBetween(openTok.Range, closeTok.Range),
	}, diags
}

func (p *parser) parseArray() (*ast.List, ksl.Diagnostics) {
	openTok := p.Read()
	if openTok.Type != lex.TokenLBrack {
		return &ast.List{SrcRange: ksl.Range{Start: openTok.Range.Start, End: openTok.Range.Start}}, ksl.Diagnostics{{
			Severity: ksl.DiagError,
			Summary:  DiagUnexpectedToken,
			Detail:   "An left bracket was expected here.",
			Subject:  &openTok.Range,
		}}
	}

	p.PushIncludeNewlines(false)
	defer p.PopIncludeNewlines()

	var closeTok lex.Token
	var diags ksl.Diagnostics
	values := []ast.Expr{}

	for {
		next := p.Peek()
		if next.Type == lex.TokenRBrack {
			closeTok = p.Read()
			break
		}

		val, valueDiags := p.parseExpression()
		values = append(values, val)
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

	return &ast.List{
		LBrack:   openTok,
		Values:   values,
		RBrack:   closeTok,
		SrcRange: ksl.RangeBetween(openTok.Range, closeTok.Range),
	}, diags
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

// recover seeks forward in the token stream until it finds lex.TokenType "end",
// then returns with the peeker pointed at the following token.
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

// oppositeBracket finds the bracket that opposes the given bracketer, or
// NilToken if the given token isn't a bracketer.
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

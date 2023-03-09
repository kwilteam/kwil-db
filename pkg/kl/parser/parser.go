package parser

import (
	"errors"
	"fmt"
	"kwil/pkg/kl/ast"
	"kwil/pkg/kl/scanner"
	"kwil/pkg/kl/token"
)

type parser struct {
	scanner *scanner.Scanner

	errors []error

	trace  bool
	curTok token.Token
}

type Opt func(*parser)

func WithTraceOff() Opt {
	return func(p *parser) {
		p.trace = false
	}
}

func New(s *scanner.Scanner, opts ...Opt) *parser {
	p := &parser{
		scanner: s,
		errors:  []error{},
		trace:   true, // use a lot
	}

	for _, opt := range opts {
		opt(p)
	}

	// init curTok
	p.next()

	return p
}

func Parse(src []byte, opts ...Opt) (a *ast.Ast, err error) {
	var p *parser
	defer func() {
		if r := recover(); r != nil {
			p.errorExpected("panic")
		}

		if len(p.errors) != 0 {
			err = errors.New(p.Error())
		}
	}()

	h := func(msg string) { p.error(msg) }
	s := scanner.New(src, h)
	p = New(s, opts...)
	a = p.Parse()
	return a, err
}

func (p *parser) Error() string {
	switch len(p.errors) {
	case 0:
		return "no errors"
	case 1:
		return p.errors[0].Error()
	default:
		return fmt.Sprintf("%s (with %d+ errors)", p.errors[0], len(p.errors)-1)
	}
}

func (p *parser) Errors() []error {
	return p.errors
}

func (p *parser) error(msg string) {
	p.errors = append(p.errors, errors.New(msg))
}

func (p *parser) errorf(format string, args ...any) {
	p.errors = append(p.errors, fmt.Errorf(format, args...))
}

func (p *parser) curTokIs(t token.TokenType) bool {
	return p.curTok.Type == t
}

func (p *parser) errorExpected(msg string) {
	p.errorf(msg)
}

func (p *parser) expect(t token.TokenType) {
	p.expectWithoutAdvance(t)
	p.next()
}

func (p *parser) expectWithoutAdvance(t token.TokenType) bool {
	if !p.curTokIs(t) {
		p.errorExpected(fmt.Sprintf("expect current token to be %s got %s instead", t, p.curTok.Type))
		return false
	}
	return true
}

func (p *parser) parseIdent() *ast.Ident {
	name := ""
	if p.curTokIs(token.IDENT) {
		name = p.curTok.Literal
		p.next()
	} else {
		p.expect(token.IDENT)
	}
	return &ast.Ident{Value: name}
}

type CompositeLit struct {
}

type CallExpr struct {
}

func (p *parser) next() {
	p.curTok = *p.scanner.Next()
}

//func (p *parser) parseParameterList() []ast.Expr {
//	if p.trace {
//		defer un(trace("parseParams"))
//	}
//
//	fmt.Println("parseParameterList", p.curTok, p.peekTok)
//
//	p.expect(token.LPAREN)
//
//	fmt.Println("parseParameterList.....", p.curTok, p.peekTok)
//
//	var params = []ast.Expr{}
//	for !p.curTokIs(token.RPAREN) && !p.curTokIs(token.EOF) {
//		switch p.curTok.Type {
//		case token.IDENT:
//			param := p.parseIdent()
//			params = append(params, param)
//		case token.INTEGER:
//			param := &ast.BasicLit{Value: p.curTok.Literal, Type: token.INTEGER}
//			params = append(params, param)
//			p.next()
//		}
//
//		//if p.peekTokIs(token.COMMA) {
//		//	p.next()
//		//}
//	}
//
//	p.expect(token.RPAREN)
//	return params
//}

func (p *parser) parseColumnAttr() *ast.AttrDef {
	if p.trace {
		defer un(trace("parseColumnAttr"))
	}

	attr := &ast.AttrDef{}
	switch p.curTok.Type {
	default:
		// attribute without parameters
		attr.Type = &ast.AttrType{
			Token: p.curTok,
		}
	}

	p.next()
	return attr
}

func (p *parser) parseColumnAttrList() []*ast.AttrDef {
	if p.trace {
		defer un(trace("parseColumnAttrList"))
	}

	attrs := []*ast.AttrDef{}

	for !p.curTokIs(token.COMMA) && !p.curTokIs(token.RBRACE) && !p.curTokIs(token.EOF) {
		if !p.curTok.Type.IsAttrType() {
			p.errorExpected(fmt.Sprintf("expect current token to be attr type got %s instead", p.curTok.Type))
			p.next()
		} else {
			attr := p.parseColumnAttr()
			attrs = append(attrs, attr)
		}
	}

	return attrs
}

func (p *parser) parseColumnDef() *ast.ColumnDef {
	if p.trace {
		defer un(trace("parseColumnDef"))
	}

	p.expectWithoutAdvance(token.IDENT)

	col := &ast.ColumnDef{}
	col.Name = p.parseIdent()
	col.Type = p.parseIdent()
	col.Attrs = p.parseColumnAttrList()

	if p.curTokIs(token.COMMA) {
		p.next()
	}

	return col
}

// parseColumnDefList parses a list of column definitions(separated by commas, enclosed in braces).
func (p *parser) parseColumnDefList() []*ast.ColumnDef {
	if p.trace {
		defer un(trace("parseColumnDefList"))
	}

	p.expect(token.LBRACE)

	columns := []*ast.ColumnDef{}

	for !p.curTokIs(token.COMMA) && !p.curTokIs(token.RBRACE) && !p.curTokIs(token.EOF) {
		columns = append(columns, p.parseColumnDef())
	}

	p.expect(token.RBRACE)

	return columns
}

func (p *parser) parseBlockDeclaration() *ast.BlockStmt {
	if p.trace {
		defer un(trace("parseBlockDeclaration"))
	}

	p.expect(token.LBRACE)

	block := &ast.BlockStmt{Token: p.curTok}
	block.Statements = []ast.Stmt{}

	for !p.curTokIs(token.RBRACE) && !p.curTokIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}

		//p.next()
	}

	p.expect(token.RBRACE)
	return block
}

func (p *parser) parseDatabaseDeclaration() *ast.DatabaseDecl {
	if p.trace {
		defer un(trace("parseDatabaseDeclaration"))
	}

	p.expect(token.DATABASE)

	db := &ast.DatabaseDecl{}
	db.Name = p.parseIdent()
	db.Body = p.parseBlockDeclaration()
	return db
}

func (p *parser) parseTableDeclaration() *ast.TableDecl {
	if p.trace {
		defer un(trace("parseTableDeclaration"))
	}

	p.expect(token.TABLE)

	stmt := &ast.TableDecl{}
	stmt.Name = p.parseIdent()
	stmt.Body = p.parseColumnDefList()

	return stmt
}

func (p *parser) parseStatement() ast.Stmt {
	if p.trace {
		defer un(trace("parseStatement"))
	}

	switch p.curTok.Type {
	case token.DATABASE:
		return p.parseDatabaseDeclaration()
	case token.TABLE:
		return p.parseTableDeclaration()
	default:
		fmt.Printf("unknown statement, token: %s\n", p.curTok)
		return nil
	}
}

func (p *parser) Parse() *ast.Ast {
	_ast := &ast.Ast{}
	_ast.Statements = []ast.Stmt{}

	for !p.curTokIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			_ast.Statements = append(_ast.Statements, stmt)
		}
		p.next()
	}

	return _ast
}

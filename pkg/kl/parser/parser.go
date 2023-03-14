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

	trace   bool
	tok     token.Token // current token
	lit     string      // current literal
	peekTok token.Token // next token
	peekLit string      // next literal
}

type Opt func(*parser)

func WithTraceOff() Opt {
	return func(p *parser) {
		p.trace = false
	}
}

func new(s *scanner.Scanner, opts ...Opt) *parser {
	p := &parser{
		scanner: s,
		errors:  []error{},
		trace:   true, // use a lot
	}

	for _, opt := range opts {
		opt(p)
	}

	// init tok and peekTok
	p.next()
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
	p = new(s, opts...)
	a = p.parse()
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

func (p *parser) curTokFrom(ts ...token.Token) bool {
	for _, t := range ts {
		if p.tok == t {
			return true
		}
	}
	return false
}

func (p *parser) curTokIs(t token.Token) bool {
	return p.tok == t
}

func (p *parser) errorExpected(msg string) {
	p.errorf(msg)
}

func (p *parser) expect(t token.Token) {
	p.expectWithoutAdvance(t)
	p.next()
}

func (p *parser) expectWithoutAdvance(t token.Token) bool {
	if !p.curTokIs(t) {
		p.errorExpected(fmt.Sprintf("expect current token to be %s got %s instead", t, p.tok))
		return false
	}
	return true
}

func (p *parser) parseBasicLit() *ast.BasicLit {
	if !p.curTokFrom(token.INTEGER, token.STRING) {
		p.errorExpected(fmt.Sprintf("expect basic literal, got %s", p.tok))
	}

	x := &ast.BasicLit{Type: p.tok, Value: p.lit}
	p.next()
	return x
}

func (p *parser) parseIdent() *ast.Ident {
	name := ""
	if p.curTokIs(token.IDENT) {
		name = p.lit
		p.next()
	} else {
		p.expect(token.IDENT)
	}
	return &ast.Ident{Name: name}
}

func (p *parser) parseIdentList() (l []*ast.Ident) {
	if p.trace {
		defer un(trace("parseIdentList"))
	}

	l = append(l, p.parseIdent())
	for p.curTokIs(token.COMMA) {
		p.next()
		l = append(l, p.parseIdent())
	}

	return l
}

func (p *parser) parseTypeName() *ast.Ident {
	if p.trace {
		defer un(trace("parseTypeName"))
	}

	return p.parseIdent()
}

type CompositeLit struct {
}

type CallExpr struct {
}

func (p *parser) next() {
	p.tok, p.lit = p.peekTok, p.peekLit
	p.peekTok, p.peekLit = p.scanner.Next()
}

func (p *parser) parseParameterList() (l []ast.Expr) {
	if p.trace {
		defer un(trace("parseParameterList"))
	}

	p.expect(token.LPAREN)

	l = append(l, p.parseParameter())
	for p.curTokIs(token.COMMA) {
		p.next()
		l = append(l, p.parseParameter())
	}

	p.expect(token.RPAREN)
	return l
}

func (p *parser) parseParameter() (param ast.Expr) {
	if p.trace {
		defer un(trace("parseParameter"))
	}

	switch p.tok {
	case token.IDENT:
		name := p.parseIdent()
		if p.curTokIs(token.PERIOD) {
			p.next()
			selector := p.parseIdent()
			param = &ast.SelectorExpr{Name: name, Sel: selector}
		} else {
			param = name
		}
	case token.INTEGER, token.STRING:
		param = p.parseBasicLit()
	}

	return
}

func (p *parser) parseColumnAttr() *ast.AttrDef {
	if p.trace {
		defer un(trace("parseColumnAttr"))
	}

	attr := &ast.AttrDef{Name: &ast.Ident{Name: p.lit}, Type: p.tok}
	switch p.tok {
	case token.MIN, token.MAX, token.MINLEN, token.MAXLEN:
		// attribute with parameters
		p.next()
		p.expect(token.LPAREN)
		if !p.curTokIs(token.RPAREN) {
			attr.Param = p.parseParameter()
		}
		p.expect(token.RPAREN)
	default:
		// attribute without parameters
		p.next()
	}

	return attr
}

func (p *parser) parseColumnAttrList() []*ast.AttrDef {
	if p.trace {
		defer un(trace("parseColumnAttrList"))
	}

	attrs := []*ast.AttrDef{}

	for !p.curTokIs(token.COMMA) && !p.curTokIs(token.RBRACE) && !p.curTokIs(token.EOF) {
		if !p.tok.IsAttr() {
			p.errorExpected(fmt.Sprintf("expect current token to be attr got (%s:%s) instead", p.tok, p.lit))
			p.next() // should advance to next attr
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
	col.Type = p.parseTypeName()
	col.Attrs = p.parseColumnAttrList()

	if p.curTokIs(token.COMMA) {
		p.next()
	}

	return col
}

func (p *parser) parserIndexDef(unique bool) *ast.IndexDef {
	if p.trace {
		defer un(trace("parserIndexDef"))
	}

	index := &ast.IndexDef{}
	index.Name = p.parseIdent()

	if unique {
		index.Unique = true
		p.expect(token.UNIQUE)
	} else {
		p.expect(token.INDEX)
	}

	index.Columns = p.parseParameterList()

	if p.curTokIs(token.COMMA) {
		p.next()
	}

	return index
}

// parseColumnDefList parses a list of column definitions(separated by commas, enclosed in braces).
func (p *parser) parseColumnDefList() (cols []ast.Stmt) {
	if p.trace {
		defer un(trace("parseColumnDefList"))
	}

	p.expect(token.LBRACE)

	for !p.curTokIs(token.COMMA) && !p.curTokIs(token.RBRACE) && !p.curTokIs(token.EOF) {
		switch p.peekTok {
		case token.INDEX:
			cols = append(cols, p.parserIndexDef(false))
		case token.UNIQUE:
			cols = append(cols, p.parserIndexDef(true))
		default:
			cols = append(cols, p.parseColumnDef())
		}
	}

	p.expect(token.RBRACE)
	return cols
}

func (p *parser) parseBlockDeclaration() *ast.BlockStmt {
	if p.trace {
		defer un(trace("parseBlockDeclaration"))
	}

	p.expect(token.LBRACE)

	block := &ast.BlockStmt{Token: p.tok}
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

	decl := &ast.TableDecl{}
	decl.Name = p.parseIdent()
	decl.Body = []ast.Stmt{}
	decl.Idx = []ast.Stmt{}

	l := p.parseColumnDefList()
	for _, v := range l {
		switch v.(type) {
		case *ast.IndexDef:
			decl.Idx = append(decl.Idx, v)
		default:
			decl.Body = append(decl.Body, v)
		}
	}

	return decl
}

func (p *parser) parseActionDeclaration() *ast.ActionDecl {
	if p.trace {
		defer un(trace("parseActionDeclaration"))
	}

	p.expect(token.ACTION)

	act := &ast.ActionDecl{}
	act.Name = p.parseIdent()
	act.Params = p.parseParameterList()

	if p.curTokIs(token.PUBLIC) || p.curTokIs(token.PRIVATE) {
		act.Public = p.tok == token.PUBLIC
		p.next()
	}

	act.Body = p.parseBlockDeclaration()

	return act
}

func (p *parser) parseInsertStatement() *ast.InsertStmt {
	if p.trace {
		defer un(trace("parseInsertStatement"))
	}

	p.expect(token.INSERT)
	p.expect(token.INTO)

	stmt := &ast.InsertStmt{}
	stmt.Table = p.parseIdent()

	// optional column list
	if !p.curTokIs(token.VALUES) {
		stmt.Columns = p.parseParameterList()
	}

	p.expect(token.VALUES)

	stmt.Values = p.parseParameterList()
	if p.curTokIs(token.COMMA) {
		p.next()
	}
	return stmt

}

func (p *parser) parseStatement() ast.Stmt {
	if p.trace {
		defer un(trace("parseStatement"))
	}

	switch p.tok {
	case token.DATABASE:
		return p.parseDatabaseDeclaration()
	case token.TABLE:
		return p.parseTableDeclaration()
	case token.ACTION:
		return p.parseActionDeclaration()
	case token.INSERT:
		return p.parseInsertStatement()
	default:
		fmt.Printf("unknown statement, token: %s, literal: %s\n", p.tok, p.lit)
		p.next()
		return nil
	}
}

func (p *parser) parse() *ast.Ast {
	// since top level is only a database declaration, maybe just dbDecl in Ast?
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

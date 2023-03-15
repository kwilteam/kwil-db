package parser

import (
	"fmt"
	"kwil/pkg/kl/ast"
	"kwil/pkg/kl/scanner"
	"kwil/pkg/kl/token"
)

type parser struct {
	scanner *scanner.Scanner
	errors  scanner.ErrorList
	file    *token.File

	trace bool
	tok   token.Token // current token
	lit   string      // current literal
	pos   token.Pos   // current position

	peekTok token.Token // next token
	peekLit string      // next literal
	peekPos token.Pos   // next position
}

type Opt func(*parser)

func WithTraceOff() Opt {
	return func(p *parser) {
		p.trace = false
	}
}

func WithTraceOn() Opt {
	return func(p *parser) {
		p.trace = true
	}
}

func (p *parser) init(src []byte, opts ...Opt) {
	eh := func(pos token.Position, msg string) { p.errors.Add(pos, msg) }
	p.scanner = scanner.New(src, eh)

	for _, opt := range opts {
		opt(p)
	}

	// init tok and peekTok
	p.next()
	p.next()
}

func Parse(src []byte, opts ...Opt) (a *ast.Database, err error) {
	var p parser
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic: %v", e)
		}

		err = p.errors.Err()
	}()

	//h := func(pos token.Position, msg string) { p.errors.Add(pos, msg) }
	//s := scanner.New(src, h)
	//p = init(s, opts...)

	p.init(src, opts...)
	a = p.parse()
	return a, err
}

func (p *parser) error(pos token.Pos, msg string) {
	p.errors.Add(p.file.Position(pos), msg)
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

func (p *parser) errorExpected(pos token.Pos, msg string) {
	msg = fmt.Sprintf("%d: %s", int(pos), msg)
	p.error(pos, msg)
}

func (p *parser) expect(t token.Token) token.Pos {
	pos := p.pos
	p.expectWithoutAdvance(t)
	p.next()
	return pos
}

func (p *parser) expectWithoutAdvance(t token.Token) bool {
	if !p.curTokIs(t) {
		p.errorExpected(p.pos, t.String())
		return false
	}
	return true
}

func (p *parser) parseBasicLit() *ast.BasicLit {
	if !p.curTokFrom(token.INTEGER, token.STRING) {
		p.errorExpected(p.pos, "integer or string")
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
	p.tok, p.lit, p.pos = p.peekTok, p.peekLit, p.peekPos
	p.peekTok, p.peekLit, p.peekPos = p.scanner.Next()
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
			p.errorExpected(p.pos, fmt.Sprintf("expect current token to be attr got (%s:%s) instead", p.tok, p.lit))
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
	case token.INSERT:
		return p.parseInsertStatement()
	default:
		fmt.Printf("unknown statement, token: %s, literal: %s\n", p.tok, p.lit)
		p.next()
		return nil
	}
}

var declStart = map[token.Token]bool{
	token.TABLE:  true,
	token.ACTION: true,
}

var stmtStart = map[token.Token]bool{
	token.INSERT: true,
}

func (p *parser) jump(to map[token.Token]bool) {
	for !p.curTokIs(token.EOF) {
		if to[p.tok] {
			return
		}
		p.next()
	}
}

// expectSemicolon expects an optional(before closing brace or paren) semicolon
func (p *parser) expectSemicolon(next map[token.Token]bool) {
	switch p.tok {
	case token.SEMICOLON:
		p.next()
	case token.RPAREN:
	case token.RBRACE:
	case token.EOF:
	default:
		p.errorExpected(p.pos, fmt.Sprintf("expected semicolon, got %s", p.tok))
		p.jump(next)
	}
}

// expectSemicolon expects an optional(before closing brace or paren) comma
func (p *parser) expectComma(next map[token.Token]bool) {
	switch p.tok {
	case token.COMMA:
		p.next()
	case token.RPAREN:
	case token.RBRACE:
	case token.EOF:
	default:
		p.errorExpected(p.pos, fmt.Sprintf("expected comma, got %s", p.tok))
		p.jump(next)
	}
}

func (p *parser) parseDeclaration() ast.Decl {
	if p.trace {
		defer un(trace("parseDeclaration"))
	}

	switch p.tok {
	case token.TABLE:
		return p.parseTableDeclaration()
	case token.ACTION:
		return p.parseActionDeclaration()
	default:
		p.errorExpected(p.pos, fmt.Sprintf("expected table or action, got %s(%s)", p.tok, p.lit))
		p.jump(declStart)
		return &ast.BadDecl{}
	}
}

func (p *parser) parse() *ast.Database {
	if p.trace {
		defer un(trace("parse"))
	}

	if len(p.errors) != 0 {
		return nil
	}

	_ = p.expect(token.DATABASE)
	dbName := p.parseIdent()
	p.expectSemicolon(declStart)

	db := &ast.Database{
		Name:  dbName,
		Decls: []ast.Decl{},
	}

	for !p.curTokIs(token.EOF) {
		db.Decls = append(db.Decls, p.parseDeclaration())
	}

	return db
}

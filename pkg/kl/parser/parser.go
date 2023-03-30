package parser

import (
	"fmt"
	"kwil/pkg/kl/ast"
	"kwil/pkg/kl/scanner"
	"kwil/pkg/kl/token"
	"strings"
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
	p.file = &token.File{Size: len(src), Lines: []int{0}}
	//p.errors = scanner.ErrorList{}

	eh := func(pos token.Position, msg string) { p.errors.Add(pos, msg) }
	p.scanner = scanner.New(src, eh, p.file)

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
			p.errors.Add(p.file.Position(p.pos), err.Error())
		}

		err = p.errors.Err()
	}()

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
	msg = "expected " + msg
	if pos == p.pos {
		// the error happened at the current position;
		// make the error message more specific
		switch {
		case p.tok.IsLiteral():
			// print 123 rather than 'INT', etc.
			msg += ", found " + p.lit
		default:
			msg += ", found '" + p.tok.String() + "'"
		}
	}
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

func (p *parser) parseParameterList(paramPrefixToken token.Token) (l []ast.Expr) {
	if p.trace {
		defer un(trace("parseParameterList"))
	}

	p.expect(token.LPAREN)

	if !p.curTokIs(token.RPAREN) {
		l = append(l, p.parseParameter(paramPrefixToken))
		for p.curTokIs(token.COMMA) {
			p.next()
			l = append(l, p.parseParameter(paramPrefixToken))
		}
	}

	p.expect(token.RPAREN)
	return l
}

func (p *parser) parseParameter(prefixToken token.Token) (param ast.Expr) {
	if p.trace {
		defer un(trace("parseParameter"))
	}

	expectPrefix := prefixToken != token.ILLEGAL

	switch p.tok {
	case token.IDENT:
		pos := p.pos
		name := p.parseIdent()
		if expectPrefix {
			if !strings.Contains(name.Name, prefixToken.String()) {
				p.errorExpected(pos, fmt.Sprintf("%s prefix", prefixToken.String()))
			}
		}

		if p.curTokIs(token.PERIOD) {
			p.next()
			selector := p.parseIdent()
			param = &ast.SelectorExpr{Name: name, Sel: selector}
		} else {
			param = name
		}
	case token.INTEGER, token.STRING:
		param = p.parseBasicLit()
		//default:
		//	p.errorExpected(p.pos, "parameter")
	}

	return
}

func (p *parser) parseColumnAttr() *ast.AttrDef {
	if p.trace {
		defer un(trace("parseColumnAttr"))
	}

	attr := &ast.AttrDef{Name: &ast.Ident{Name: p.lit}, Type: p.tok}
	switch p.tok {
	case token.MIN, token.MAX, token.MINLEN, token.MAXLEN, token.DEFAULT:
		// attribute with parameters
		p.next()
		p.expect(token.LPAREN)
		if !p.curTokIs(token.RPAREN) {
			attr.Param = p.parseParameter(token.ILLEGAL)
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
			p.errorExpected(p.pos, "column attribute")
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

	colName := p.parseIdent()
	colType := p.parseTypeName()
	colAttrs := p.parseColumnAttrList()

	if p.curTokIs(token.COMMA) {
		p.next()
	}

	return &ast.ColumnDef{Name: colName, Type: colType, Attrs: colAttrs}
}

func (p *parser) parserIndexDef(unique bool) *ast.IndexDef {
	if p.trace {
		defer un(trace("parserIndexDef"))
	}

	indexName := p.parseIdent()
	indexUnique := false
	if unique {
		indexUnique = true
		p.expect(token.UNIQUE)
	} else {
		p.expect(token.INDEX)
	}

	indexColumns := p.parseParameterList(token.ILLEGAL)

	if p.curTokIs(token.COMMA) {
		p.next()
	}

	return &ast.IndexDef{Name: indexName, Unique: indexUnique, Columns: indexColumns}
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
	act.Params = p.parseParameterList(token.DOLLAR)

	if p.curTokIs(token.PUBLIC) || p.curTokIs(token.PRIVATE) {
		act.Public = p.tok == token.PUBLIC
		p.next()
	}

	act.Body = p.parseBlockDeclaration()

	return act
}

// parseSQLStatement parses a whole SQL statement as a string.
func (p *parser) parseSQLStatement() *ast.SQLStmt {
	if p.trace {
		defer un(trace("parseInsertStatement"))
	}

	var rawSQL []string
	for !p.curTokIs(token.SEMICOLON) && !p.curTokIs(token.RBRACE) && !p.curTokIs(token.EOF) {
		tok := p.tok
		lit := p.lit
		p.next()

		// parse table.column tokens
		if tok == token.IDENT && p.curTokIs(token.PERIOD) {
			p.next()
			selector := p.parseIdent()
			lit = fmt.Sprintf("%s.%s", lit, selector)
		}

		// parse function calls, left parenthesis needs to be appended to the function name
		if p.tok == token.LPAREN {
			switch strings.ToLower(lit) {
			case ",", ";", "from", "as", "join", "on", "where", "group", "having", "order", "limit", "offset", "into", "values":
			default:
				lit += "("
				p.next()
			}
		}

		rawSQL = append(rawSQL, lit)
	}

	if p.curTokIs(token.SEMICOLON) {
		p.next()
	}

	return &ast.SQLStmt{SQL: strings.Join(rawSQL, " ")}
}

func (p *parser) parseStatement() ast.Stmt {
	if p.trace {
		defer un(trace("parseStatement"))
	}

	pos := p.pos

	switch p.tok {
	case token.INSERT, token.WITH, token.REPLACE, token.SELECT, token.UPDATE, token.DROP, token.DELETE:
		return p.parseSQLStatement()
	default:
		p.errorExpected(pos, fmt.Sprintf("unknown statement, token: %s, literal: %s\n", p.tok, p.lit))
		p.next()
		return nil
	}
}

var declStart = map[token.Token]bool{
	token.TABLE:  true,
	token.ACTION: true,
}

var sqlStart = map[token.Token]bool{
	token.INSERT:  true,
	token.SELECT:  true,
	token.UPDATE:  true,
	token.DROP:    true,
	token.WITH:    true,
	token.REPLACE: true,
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
		p.errorExpected(p.pos, "semicolon")
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
		p.errorExpected(p.pos, "comma")
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
		p.errorExpected(p.pos, "table or action")
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

	if err := db.Validate(); err != nil {
		p.errorExpected(p.pos, err.Error())
	}

	return db
}

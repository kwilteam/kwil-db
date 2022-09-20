package dbml

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
)

type Parser struct {
	s *Scanner

	token Token
	lit   string

	Debug bool
}

type Option func(p *Parser)

func WithDebug() Option {
	return func(p *Parser) {
		p.Debug = true
	}
}

func NewParser(s *Scanner, opts ...Option) *Parser {
	p := &Parser{
		s:     s,
		token: ILLEGAL,
		lit:   "",
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

func (p *Parser) Parse() (*DBML, error) {
	dbml := &DBML{}
	p.advance()
	for {
		switch p.token {
		case PROJECT:
			project, err := p.parseProject()
			if err != nil {
				return nil, err
			}
			p.debug("project", project)
			dbml.Project = *project
		case TABLE:
			table, err := p.parseTable()
			if err != nil {
				return nil, err
			}
			p.debug("table", table)

			// TODO:
			// * register table to tables map, for check ref
			dbml.Tables = append(dbml.Tables, *table)

		case REF:
			ref, err := p.parseRefs()
			if err != nil {
				return nil, err
			}
			p.debug("Refs", ref)

			// TODO:
			// * Check refs is valid or not (by tables map)
			dbml.Refs = append(dbml.Refs, *ref)

		case ROLE:
			return nil, fmt.Errorf("role is not supported yet")

		case QUERY:
			query, err := p.parseQuery()
			if err != nil {
				return nil, err
			}
			p.debug("query", query)
			dbml.Queries = append(dbml.Queries, *query)

		case ENUM:
			enum, err := p.parseEnum()
			if err != nil {
				return nil, err
			}
			p.debug("Enum", enum)
			dbml.Enums = append(dbml.Enums, *enum)

		case TABLEGROUP:
			tableGroup, err := p.parseTableGroup()
			if err != nil {
				return nil, err
			}
			p.debug("TableGroup", tableGroup)
			dbml.TableGroups = append(dbml.TableGroups, *tableGroup)
		case EOF:
			return dbml, nil
		default:
			p.debug("token", p.token.String(), "lit", p.lit)
			return nil, p.expect("project, ref, table, enum, query, role, tablegroup")
		}
	}
}

func (p *Parser) parseRole() (*Role, error) {
	return nil, nil
}

func (p *Parser) parseQuery() (query *Query, err error) {
	query = &Query{}

	p.s.BlockMode(func() {
		p.advance()

		// Handle for query <optional_name>...
		if IsIdent(p.token) {
			query.Name = p.lit
			p.advance()
		}

		switch p.token {
		case COLON:
			p.advance()
			if p.token != EXPR && p.token != TSTRING {
				err = p.expect("expr | tstring")
				return
			}

			query.Expression = strings.TrimSpace(p.lit)
		case BLOCK:
			query.Expression = strings.TrimSpace(p.lit)
		}
	})

	p.advance()
	return query, nil
}

func (p *Parser) parseTableGroup() (*TableGroup, error) {
	tableGroup := &TableGroup{}
	p.advance()
	if p.token != IDENT && p.token != DSTRING {
		return nil, fmt.Errorf("TableGroup name is invalid: %s", p.lit)
	}
	tableGroup.Name = p.lit
	p.advance()
	if p.token != LBRACE {
		return nil, p.expect("{")
	}
	p.advance()

	for p.token == IDENT || p.token == DSTRING {
		tableGroup.Members = append(tableGroup.Members, p.lit)
		p.advance()
	}
	if p.token != RBRACE {
		return nil, p.expect("}")
	}

	p.advance()
	return tableGroup, nil
}

func (p *Parser) parseQualfiedName() (string, error) {
	var name string
	if !IsIdent(p.token) && p.token != DSTRING {
		return name, fmt.Errorf("name is invalid: %s", p.lit)
	}

	var parts []string
	r := regexp.MustCompile("^[a-zA-Z1-9]+$")
	for {
		switch p.token {
		case IDENT, DSTRING:
		default:
			if !r.MatchString(p.lit) {
				return name, fmt.Errorf("name is invalid: %s", p.lit)
			}
		}
		parts = append(parts, p.lit)
		p.advance()
		if p.token != PERIOD {
			break
		}
		p.advance()
	}

	return strings.Join(parts, "."), nil
}

func (p *Parser) parseEnum() (*Enum, error) {
	enum := &Enum{}
	p.advance()

	name, err := p.parseQualfiedName()
	if err != nil {
		return nil, err
	}

	enum.Name = name

	if p.token != LBRACE {
		return nil, p.expect("{")
	}
	p.advance()

	for IsIdent(p.token) {
		enumValue := EnumValue{
			Name: p.lit,
		}
		p.advance()
		if p.token == LBRACK {
			// handle [Note: ...]
			p.advance()
			if p.token == NOTE {
				note, err := p.parseDescription()
				if err != nil {
					return nil, p.expect("note: 'string'")
				}
				enumValue.Note = note
				p.advance()
			}
			if p.token != RBRACK {
				return nil, p.expect("]")
			}
			p.advance()
		}
		enum.Values = append(enum.Values, enumValue)
	}
	if p.token != RBRACE {
		return nil, p.expect("}")
	}

	p.advance()
	return enum, nil
}

func (p *Parser) parseRefs() (*Ref, error) {
	ref := &Ref{}
	p.advance()

	// Handle for Ref <optional_name>...
	if p.token == IDENT {
		ref.Name = p.lit
		p.advance()
	}

	// Ref: from > to
	if p.token == COLON {
		p.advance()
		rel, err := p.parseRelationship()
		if err != nil {
			return nil, err
		}
		ref.Relationships = append(ref.Relationships, *rel)
		return ref, nil
	}

	if p.token == LBRACE {
		p.advance()

		for {
			switch p.token {
			case RBRACE:
				p.advance()
				return ref, nil
			case IDENT, DSTRING:
				rel, err := p.parseRelationship()
				if err != nil {
					return nil, err
				}
				ref.Relationships = append(ref.Relationships, *rel)
			default:
				return nil, p.expect("Ref: { from > to }")
			}
		}
	}

	return nil, p.expect("Ref: | Refs {}")
}

func (p *Parser) parseFullyQualifiedRelationship() (Rel, error) {
	fqr := Rel{}
	if p.token != IDENT {
		return fqr, p.expect("schema?.table.column(s)")
	}

	var parts []string

	for {
		if p.token == IDENT || p.token == DSTRING {
			parts = append(parts, p.lit)
		} else if p.token == LPAREN {
			p.advance()
			for IsIdent(p.token) {
				parts = append(parts, p.lit)
				p.advance()
				if p.token == COMMA {
					p.advance()
				}
			}
			if p.token != RPAREN {
				return fqr, p.expect(")")
			}
			p.advance()
			break
		} else if p.token == WHITESPACE || p.token == NEWLINE {
			p.advance()
			break
		} else if p.token != PERIOD {
			break
		}
		p.advanceIgnore(COMMENT)
	}

	switch len(parts) {
	case 1:
		return fqr, fmt.Errorf("invalid fully qualified relationship: %s", strings.Join(parts, "."))
	case 2:
		fqr.Name = parts[0]
		fqr.Columns = []string{parts[1]}
	default:
		fqr.Name = strings.Join(parts[:2], ".")
		fqr.Columns = parts[2:]
	}

	return fqr, nil
}

func (p *Parser) parseRelationship() (*Relationship, error) {
	rel := &Relationship{}

	from, err := p.parseFullyQualifiedRelationship()
	if err != nil {
		return nil, err
	}
	rel.From = from

	if reltype, ok := RelationshipMap[p.token]; ok {
		rel.Type = reltype
	} else {
		return nil, p.expect("> | < | <> | -")
	}

	p.advance()

	to, err := p.parseFullyQualifiedRelationship()
	if err != nil {
		return nil, err
	}
	rel.To = to
	return rel, nil
}

func (p *Parser) parseTable() (*Table, error) {
	table := &Table{}
	p.advance()

	name, err := p.parseQualfiedName()
	if err != nil {
		return nil, err
	}
	table.Name = name

	switch p.token {
	case AS:
		// handle as
		p.advance()
		switch p.token {
		case STRING, IDENT:
			table.Alias = p.lit
		default:
			return nil, p.expect("as NAME")
		}
		p.advance()
		fallthrough
	case LBRACE:
		p.advance()
		for {
			switch p.token {
			case INDEXES:
				indexes, err := p.parseIndexes()
				if err != nil {
					return nil, err
				}
				table.Indexes = indexes
			case RBRACE:
				p.advance()
				return table, nil
			default:
				columnName := p.lit
				currentToken := p.token
				p.advance()
				if currentToken == NOTE && p.token == COLON {
					note, err := p.parseString()
					if err != nil {
						return nil, err
					}
					table.Note = note
					p.advance()
				} else {
					column, err := p.parseColumn(columnName)
					if err != nil {
						return nil, err
					}
					table.Columns = append(table.Columns, *column)
				}
			}
		}
	default:
		return nil, p.expect("{")
	}
}

func (p *Parser) parseIndexes() ([]Index, error) {
	indexes := []Index{}

	p.advance()
	if p.token != LBRACE {
		return nil, p.expect("{")
	}

	p.advance()
	for {
		if p.token == RBRACE {
			p.advance() // pop }
			return indexes, nil
		}
		// parse an Index
		index, err := p.parseIndex()
		if err != nil {
			return nil, err
		}
		p.debug("index", index)
		indexes = append(indexes, *index)
	}
}

func (p *Parser) parseIndex() (*Index, error) {
	index := &Index{}

	if p.token == LPAREN {
		p.advance()
		for IsIdent(p.token) {
			index.Fields = append(index.Fields, p.lit)
			p.advance()
			if p.token == COMMA {
				p.advance()
			}
		}
		if p.token != RPAREN {
			return nil, p.expect(")")
		}
	} else if IsIdent(p.token) {
		index.Fields = append(index.Fields, p.lit)
	} else {
		return nil, p.expect("field_name")
	}

	p.advance()

	if p.token == LBRACK {
		// Handle index setting [settings...]
		commaAllowed := false

		for {
			p.advance()
			switch {
			case p.token == IDENT && strings.ToLower(p.lit) == "name":
				name, err := p.parseDescription()
				if err != nil {
					return nil, p.expect("name: 'index_name'")
				}
				index.Settings.Name = name
			case p.token == NOTE:
				note, err := p.parseDescription()
				if err != nil {
					return nil, p.expect("note: 'index note'")
				}
				index.Settings.Note = note
			case p.token == PK:
				index.Settings.PK = true
			case p.token == UNIQUE:
				index.Settings.Unique = true
			case p.token == TYPE:
				p.advance()
				if p.token != COLON {
					return nil, p.expect(":")
				}
				p.advance()
				if p.token != IDENT || (p.lit != "hash" && p.lit != "btree") {
					return nil, p.expect("hash|btree")
				}
				index.Settings.Type = p.lit
			case p.token == COMMA:
				if !commaAllowed {
					return nil, p.expect("[index settings...]")
				}
			case p.token == RBRACK:
				p.advance()
				return index, nil
			default:
				return nil, p.expect("note|name|type|pk|unique")
			}
			commaAllowed = !commaAllowed
		}
	}

	return index, nil
}

func (p *Parser) parseColumn(name string) (*Column, error) {
	column := &Column{
		Name: name,
	}
	if p.token != IDENT {
		return nil, p.expect("int, varchar,...")
	}
	colType := p.lit
	p.advance()

	for p.token == PERIOD {
		p.advance()
		if p.token != IDENT {
			return nil, p.expect("int, varchar,...")
		}
		colType += "." + p.lit
		p.advance()
	}

	column.Type = colType

	// parse for type
	switch p.token {
	case LPAREN:
		p.advance()
		if p.token != INT {
			return nil, p.expect("int")
		}

		sz, err := strconv.ParseInt(p.lit, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid size: %s", p.lit)
		}
		column.Size = int(sz)

		p.advance()
		if p.token != RPAREN {
			return nil, p.expect(RPAREN.String())
		}
		p.advance()
		if p.token != LBRACK {
			break
		}
		fallthrough
	case LBRACK:
		columnSetting, err := p.parseColumnSettings()
		if err != nil {
			return nil, fmt.Errorf("parse column settings: %w", err)
		}
		column.Settings = *columnSetting
	}

	p.debug("column", column)
	return column, nil
}

func (p *Parser) parseColumnSettings() (*ColumnSetting, error) {
	columnSetting := &ColumnSetting{}
	commaAllowed := false

	p.advance()
	for {
		switch p.token {
		case PK:
			columnSetting.PK = true
			p.advance()
		case PRIMARY:
			p.advance()
			if p.token != KEY {
				return nil, p.expect("KEY")
			}
			columnSetting.PK = true
			p.advance()
		case REF:
			p.advance()
			if p.token != COLON {
				return nil, p.expect(":")
			}
			p.advance()
			if p.token != LT && p.token != GT && p.token != LTGT && p.token != SUB {
				return nil, p.expect("< | > | <> | -")
			}
			columnSetting.Ref = &OneWayRef{}
			columnSetting.Ref.Type = RelationshipMap[p.token]
			p.advance()
			fqr, err := p.parseFullyQualifiedRelationship()
			if err != nil {
				return nil, err
			}

			columnSetting.Ref.To = fqr
		case NOT:
			p.advance()
			if p.token != NULL {
				return nil, p.expect("null")
			}
			columnSetting.NotNull = true
			p.advance()
		case UNIQUE:
			columnSetting.Unique = true
			p.advance()
		case INCREMENT:
			columnSetting.AutoIncrement = true
			p.advance()
		case DEFAULT:
			p.advance()
			if p.token != COLON {
				return nil, p.expect(":")
			}
			p.advance()
			switch p.token {
			case STRING, DSTRING, TSTRING, INT, FLOAT, EXPR:
				//TODO:
				//	* handle default value by expr
				//	* validate default value by type
				columnSetting.Default = p.lit
			default:
				return nil, p.expect("default value")
			}
			p.advance()
		case NOTE:
			str, err := p.parseDescription()
			if err != nil {
				return nil, err
			}
			columnSetting.Note = str
			p.advance()
		case COMMA:
			if !commaAllowed {
				return nil, p.expect("pk | primary key | unique")
			}
			p.advance()
		case RBRACK:
			p.advance()
			return columnSetting, nil
		default:
			return nil, p.expect("pk, primary key, unique")
		}
		commaAllowed = !commaAllowed
	}
}

func (p *Parser) parseProject() (*Project, error) {
	project := &Project{}
	p.advance()
	if p.token != IDENT && p.token != DSTRING {
		return nil, p.expect("project_name")
	}

	project.Name = p.lit
	p.advance()

	if p.token != LBRACE {
		return nil, p.expect("{")
	}
	for {
		p.advance()
		switch p.token {
		case IDENT:
			switch p.lit {
			case "database_type":
				str, err := p.parseDescription()
				if err != nil {
					return nil, err
				}
				project.DatabaseType = str
			default:
				return nil, p.expect("database_type")
			}
		case NOTE:
			note, err := p.parseDescription()
			if err != nil {
				return nil, err
			}
			project.Note = note
		case RBRACE:
			p.advance()
			return project, nil
		default:
			return nil, fmt.Errorf("invalid token %s", p.lit)
		}
	}
}

func (p *Parser) parseString() (string, error) {
	p.advance()
	switch p.token {
	case STRING, DSTRING, TSTRING:
		return p.lit, nil
	default:
		return "", p.expect("string, double quote string, triple string")
	}
}

func (p *Parser) parseDescription() (string, error) {
	p.advance()
	if p.token != COLON {
		return "", p.expect(":")
	}
	return p.parseString()
}

func (p *Parser) advance() {
	p.advanceIgnore(COMMENT, WHITESPACE, NEWLINE)
}

func (p *Parser) advanceIgnore(tokens ...Token) {
	for {
		p.token, p.lit = p.s.Read()
		p.debug("token:", p.token.String(), "lit:", p.lit)
		if !slices.Contains(tokens, p.token) {
			break
		}
	}
}

func (p *Parser) expect(expected string) error {
	l, c := p.s.LineInfo()
	return fmt.Errorf("[%d:%d] invalid token '%s', expected: '%s'", l, c, p.lit, expected)
}

func (p *Parser) debug(args ...interface{}) {
	if p.Debug {
		for _, arg := range args {
			fmt.Printf("%#v\t", arg)
		}
		fmt.Println()
	}
}

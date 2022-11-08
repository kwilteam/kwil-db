package sqlspec

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"ksl/sqlutil"
)

type planner struct{}

func NewPlanner() Planner {
	return &planner{}
}

func (p *planner) PlanChanges(changes []SchemaChange, opts ...PlanOption) (*Plan, error) {
	var planOpts PlanOptions
	for _, o := range opts {
		o(&planOpts)
	}
	s := &state{
		Plan: Plan{
			Name:          planOpts.Name,
			Reversible:    true,
			Transactional: true,
		},
		PlanOptions: planOpts,
	}
	if err := s.plan(changes); err != nil {
		return nil, err
	}
	for _, c := range s.Changes {
		if c.Reverse == "" {
			s.Reversible = false
		}
	}
	return &s.Plan, nil
}

type state struct {
	Plan
	PlanOptions
}

func (s *state) plan(changes []SchemaChange) error {
	if s.SchemaQualifier != nil {
		if err := CheckChangesScope(changes); err != nil {
			return err
		}
	}
	planned := s.topLevel(changes)
	planned, err := DetachCycles(planned)
	if err != nil {
		return err
	}
	for _, c := range planned {
		switch c := c.(type) {
		case *AddTable:
			err = s.addTable(c)
		case *DropTable:
			s.dropTable(c)
		case *ModifyTable:
			err = s.modifyTable(c)
		case *RenameTable:
			s.renameTable(c)
		case *AddEnum:
			err = s.addEnum(c)
		case *DropEnum:
			err = s.dropEnum(c)
		case *ModifyEnum:
			err = s.modifyEnum(c)
		case *AddQuery:
			err = s.addQuery(c)
		case *DropQuery:
			err = s.dropQuery(c)
		case *ModifyQuery:
			err = s.modifyQuery(c)
		case *AddRole:
			err = s.addRole(c)
		case *DropRole:
			err = s.dropRole(c)
		case *ModifyRole:
			err = s.modifyRole(c)
		case *AddQueryToRole:
			err = s.addRoleQuery(c)
		case *DropQueryFromRole:
			err = s.dropRoleQuery(c)
		default:
			err = fmt.Errorf("unsupported change %T", c)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *state) addQuery(in *AddQuery) error       { return nil }
func (s *state) dropQuery(in *DropQuery) error     { return nil }
func (s *state) modifyQuery(in *ModifyQuery) error { return nil }
func (s *state) addRole(in *AddRole) error         { return nil }
func (s *state) dropRole(in *DropRole) error       { return nil }
func (s *state) modifyRole(in *ModifyRole) error   { return nil }

func (s *state) addRoleQuery(in *AddQueryToRole) error {
	s.append(&Change{
		Cmd:     s.addRoleQueryStatement(in.R, in.Q),
		Source:  in,
		Comment: fmt.Sprintf("Drop query %q from role %q", in.Q.Name, in.R.Name),
		Reverse: s.dropRoleQueryStatement(in.R, in.Q),
	})
	return nil
}

func (s *state) dropRoleQuery(in *DropQueryFromRole) error {
	s.append(&Change{
		Cmd:     s.dropRoleQueryStatement(in.R, in.Q),
		Source:  in,
		Comment: fmt.Sprintf("Drop query %q from role %q", in.Q.Name, in.R.Name),
		Reverse: s.addRoleQueryStatement(in.R, in.Q),
	})

	return nil
}

func (s *state) dropRoleQueryStatement(r *Role, q *Query) string {
	b := s.Build("DELETE FROM").Table(KwilManagementSchema, KwilRoleQueriesTable).P("rq")
	b.P("USING").Table(KwilManagementSchema, KwilRoleTable).P("r").Comma().Table(KwilManagementSchema, KwilQueryTable).P("q")
	b.P("WHERE", "r.name", "=").Ident(r.Name).P("AND", "q.name", "=").Ident(q.Name)
	return b.String()
}

func (s *state) addRoleQueryStatement(r *Role, q *Query) string {
	b := s.Build("INSERT INTO").Table(KwilManagementSchema, KwilRoleQueriesTable)
	b.Wrap(func(b *sqlutil.Builder) { b.P("role_id", ",", "query_id") })
	b.P("VALUES").Wrap(func(b *sqlutil.Builder) {
		b.Wrap(func(b *sqlutil.Builder) {
			b.P("SELECT", "id", "FROM").Table(KwilManagementSchema, KwilRoleTable).P("WHERE name", "=", quote(r.Name))
		})
		b.P(",")
		b.Wrap(func(b *sqlutil.Builder) {
			b.P("SELECT", "id", "FROM").Table(KwilManagementSchema, KwilQueryTable).P("WHERE", "name", "=", quote(q.Name))
		})
	})
	return b.String()
}

func (s *state) topLevel(changes []SchemaChange) []SchemaChange {
	planned := make([]SchemaChange, 0, len(changes))
	for _, c := range changes {
		switch c := c.(type) {
		case *AddSchema:
			b := s.Build("CREATE SCHEMA")
			if has(c.Extra, &IfNotExists{}) {
				b.P("IF NOT EXISTS")
			}
			b.Ident(c.S.Name)
			s.append(&Change{
				Cmd:     b.String(),
				Source:  c,
				Reverse: s.Build("DROP SCHEMA").Ident(c.S.Name).P("CASCADE").String(),
				Comment: fmt.Sprintf("Add new schema named %q", c.S.Name),
			})
		case *DropSchema:
			b := s.Build("DROP SCHEMA")
			if has(c.Extra, &IfExists{}) {
				b.P("IF EXISTS")
			}
			b.Ident(c.S.Name).P("CASCADE")
			s.append(&Change{
				Cmd:     b.String(),
				Source:  c,
				Comment: fmt.Sprintf("Drop schema named %q", c.S.Name),
			})
		default:
			planned = append(planned, c)
		}
	}
	return planned
}

func (s *state) addEnum(in *AddEnum) error {
	e := in.E
	et := &EnumType{T: e.Name, Values: e.Values[:], Schema: e.Schema}
	create, drop := s.createDropEnum(e.Schema, et)
	s.append(&Change{
		Cmd:     create,
		Reverse: drop,
		Comment: fmt.Sprintf("create enum type %q", et.T),
	})
	return nil
}

func (s *state) dropEnum(in *DropEnum) error {
	e := in.E
	et := &EnumType{T: e.Name, Values: e.Values[:], Schema: e.Schema}
	create, drop := s.createDropEnum(e.Schema, et)
	s.append(&Change{
		Cmd:     drop,
		Reverse: create,
		Comment: fmt.Sprintf("drop enum type %q", et.T),
	})
	return nil
}

func (s *state) modifyEnum(mod *ModifyEnum) error {
	for _, c := range mod.Changes {
		switch c := c.(type) {
		case *AddEnumValue:
			s.append(&Change{
				Cmd:     s.Build("ALTER TYPE").P(c.E.Name, "ADD VALUE", quote(c.V)).String(),
				Comment: fmt.Sprintf("add value to enum type: %q", c.E.Name),
			})
		default:
			return fmt.Errorf("unsupported enum modification %T", c)
		}
	}
	return nil
}

// addTable builds and executes the query for creating a table in a
func (s *state) addTable(add *AddTable) error {
	// Create enum types before using them in the `CREATE TABLE` statement.
	// if err := s.mayAddEnums(add.T, add.T.Columns...); err != nil {
	// 	return err
	// }
	var (
		errs []string
		b    = s.Build("CREATE TABLE")
	)
	if has(add.Extra, &IfNotExists{}) {
		b.P("IF NOT EXISTS")
	}
	b.Table(tname(add.T))
	b.Wrap(func(b *sqlutil.Builder) {
		b.MapComma(add.T.Columns, func(i int, b *sqlutil.Builder) {
			if err := s.column(b, add.T, add.T.Columns[i]); err != nil {
				errs = append(errs, err.Error())
			}
		})
		if pk := add.T.PrimaryKey; pk != nil {
			b.Comma().P("PRIMARY KEY")
			s.indexParts(b, pk.Parts)
		}
		if len(add.T.ForeignKeys) > 0 {
			b.Comma()
			s.fks(b, add.T.ForeignKeys...)
		}
		for _, attr := range add.T.Attrs {
			if c, ok := attr.(*Check); ok {
				b.Comma()
				check(b, c)
			}
		}
	})
	if p := (Partition{}); has(add.T.Attrs, &p) {
		s, err := formatPartition(p)
		if err != nil {
			errs = append(errs, err.Error())
		}
		b.P(s)
	}
	if len(errs) > 0 {
		return fmt.Errorf("create table %q: %s", add.T.Name, strings.Join(errs, ", "))
	}
	s.append(&Change{
		Cmd:     b.String(),
		Source:  add,
		Comment: fmt.Sprintf("create %q table", add.T.Name),
		Reverse: s.Build("DROP TABLE").Table(tname(add.T)).String(),
	})
	s.addIndexes(add.T, add.T.Indexes...)
	s.addComments(add.T)
	return nil
}

// dropTable builds and executes the query for dropping a table from a
func (s *state) dropTable(drop *DropTable) {
	b := s.Build("DROP TABLE")
	if has(drop.Extra, &IfExists{}) {
		b.P("IF EXISTS")
	}
	b.Table(tname(drop.T))
	s.append(&Change{
		Cmd:     b.String(),
		Source:  drop,
		Comment: fmt.Sprintf("drop %q table", drop.T.Name),
	})
}

// modifyTable builds the statements that bring the table into its modified state.
func (s *state) modifyTable(modify *ModifyTable) error {
	var (
		alter       []SchemaChange
		addI, dropI []*Index
		changes     []*Change
	)
	for _, change := range skipAutoChanges(modify.Changes) {
		switch change := change.(type) {
		case *AddAttr, *ModifyAttr:
			from, to, err := commentChange(change)
			if err != nil {
				return err
			}
			changes = append(changes, s.tableComment(modify.T, to, from))
		case *DropAttr:
			return fmt.Errorf("unsupported change type: %T", change)
		case *AddIndex:
			if c := (Comment{}); has(change.I.Attrs, &c) {
				changes = append(changes, s.indexComment(modify.T, change.I, c.Text, ""))
			}
			addI = append(addI, change.I)
		case *DropIndex:
			// Unlike DROP INDEX statements that are executed separately,
			// DROP CONSTRAINT are added to the ALTER TABLE statement below.
			if isUniqueConstraint(change.I) {
				alter = append(alter, change)
			} else {
				dropI = append(dropI, change.I)
			}
		case *ModifyIndex:
			k := change.Change
			if change.Change.Is(ChangeComment) {
				from, to, err := commentChange(CommentDiff(change.From.Attrs, change.To.Attrs))
				if err != nil {
					return err
				}
				changes = append(changes, s.indexComment(modify.T, change.To, to, from))
				// If only the comment of the index was changed.
				if k &= ^ChangeComment; k.Is(NoChange) {
					continue
				}
			}
			// Index modification requires rebuilding the index.
			addI = append(addI, change.To)
			dropI = append(dropI, change.From)
		case *RenameIndex:
			changes = append(changes, &Change{
				Source:  change,
				Comment: fmt.Sprintf("rename an index from %q to %q", change.From.Name, change.To.Name),
				Cmd:     s.Build("ALTER INDEX").Ident(change.From.Name).P("RENAME TO").Ident(change.To.Name).String(),
				Reverse: s.Build("ALTER INDEX").Ident(change.To.Name).P("RENAME TO").Ident(change.From.Name).String(),
			})
		case *ModifyForeignKey:
			// Foreign-key modification is translated into 2 steps.
			// Dropping the current foreign key and creating a new one.
			alter = append(alter, &DropForeignKey{
				F: change.From,
			}, &AddForeignKey{
				F: change.To,
			})
		case *AddColumn:
			// if err := s.mayAddEnums(modify.T, change.C); err != nil {
			// 	return err
			// }
			if c := (Comment{}); has(change.C.Attrs, &c) {
				changes = append(changes, s.columnComment(modify.T, change.C, c.Text, ""))
			}
			alter = append(alter, change)
		case *ModifyColumn:
			k := change.Change
			if change.Change.Is(ChangeComment) {
				from, to, err := commentChange(CommentDiff(change.From.Attrs, change.To.Attrs))
				if err != nil {
					return err
				}
				changes = append(changes, s.columnComment(modify.T, change.To, to, from))
				// If only the comment of the column was changed.
				if k &= ^ChangeComment; k.Is(NoChange) {
					continue
				}
			}
			// from, ok1 := hasEnumType(change.From)
			// to, ok2 := hasEnumType(change.To)
			// switch {
			// Enum was changed (underlying values).
			// case ok1 && ok2 && s.enumIdent(modify.T.Schema, from) == s.enumIdent(modify.T.Schema, to):
			// if err := s.alterEnum(modify.T, from, to); err != nil {
			// 	return err
			// }
			// If only the enum values were changed,
			// there is no need to ALTER the table.
			// if k == ChangeType {
			// 	continue
			// }
			// Enum was added or changed.
			// case !ok1 && ok2 ||
			// 	ok1 && ok2 && s.enumIdent(modify.T.Schema, from) != s.enumIdent(modify.T.Schema, to):
			// if err := s.mayAddEnums(modify.T, change.To); err != nil {
			// 	return err
			// }
			// }
			alter = append(alter, &ModifyColumn{To: change.To, From: change.From, Change: k})
		case *RenameColumn:
			// "RENAME COLUMN" cannot be combined with other alterations.
			b := s.Build("ALTER TABLE").Table(tname(modify.T)).P("RENAME COLUMN")
			r := b.Clone()
			changes = append(changes, &Change{
				Source:  change,
				Comment: fmt.Sprintf("rename a column from %q to %q", change.From.Name, change.To.Name),
				Cmd:     b.Ident(change.From.Name).P("TO").Ident(change.To.Name).String(),
				Reverse: r.Ident(change.To.Name).P("TO").Ident(change.From.Name).String(),
			})
		default:
			alter = append(alter, change)
		}
	}
	s.dropIndexes(modify.T, dropI...)
	if len(alter) > 0 {
		if err := s.alterTable(modify.T, alter); err != nil {
			return err
		}
	}
	s.addIndexes(modify.T, addI...)
	s.append(changes...)
	return nil
}

// alterTable modifies the given table by executing on it a list of changes in one SQL statement.
func (s *state) alterTable(t *Table, changes []SchemaChange) error {
	var (
		reverse    []SchemaChange
		reversible = true
	)

	build := func(alter *alterChange, changes []SchemaChange) (string, error) {
		b := s.Build("ALTER TABLE").Table(tname(t))
		err := b.MapCommaErr(changes, func(i int, b *sqlutil.Builder) error {
			switch change := changes[i].(type) {
			case *AddColumn:
				b.P("ADD COLUMN")
				if err := s.column(b, t, change.C); err != nil {
					return err
				}
				reverse = append(reverse, &DropColumn{C: change.C})
			case *ModifyColumn:
				if err := s.alterColumn(b, alter, t, change); err != nil {
					return err
				}
				if change.Change.Is(ChangeGenerated) {
					reversible = false
				}
				reverse = append(reverse, &ModifyColumn{
					From:   change.To,
					To:     change.From,
					Change: change.Change & ^ChangeGenerated,
				})
				// toE, toHas := hasEnumType(change.To)
				// fromE, fromHas := hasEnumType(change.From)
				// In case the enum was dropped or replaced with a different one.
				// if fromHas && !toHas || fromHas && toHas && s.enumIdent(t.Schema, fromE) != s.enumIdent(t.Schema, toE) {
				// 	if err := s.mayDropEnum(alter, t.Schema, fromE); err != nil {
				// 		return err
				// 	}
				// }
			case *DropColumn:
				b.P("DROP COLUMN").Ident(change.C.Name)
				reverse = append(reverse, &AddColumn{C: change.C})
				// if e, ok := hasEnumType(change.C); ok {
				// 	if err := s.mayDropEnum(alter, t.Schema, e); err != nil {
				// 		return err
				// 	}
				// }
			case *AddIndex:
				b.P("ADD CONSTRAINT").Ident(change.I.Name).P("UNIQUE")
				s.indexParts(b, change.I.Parts)
				// Skip reversing this operation as it is the inverse of
				// the operation below and should not be used besides this.
			case *DropIndex:
				b.P("DROP CONSTRAINT").Ident(change.I.Name)
				reverse = append(reverse, &AddIndex{I: change.I})
			case *AddForeignKey:
				b.P("ADD")
				s.fks(b, change.F)
				reverse = append(reverse, &DropForeignKey{F: change.F})
			case *DropForeignKey:
				b.P("DROP CONSTRAINT").Ident(change.F.Name)
				reverse = append(reverse, &AddForeignKey{F: change.F})
			case *AddCheck:
				check(b.P("ADD"), change.C)
				// Reverse operation is supported if
				// the constraint name is not generated.
				if reversible = reversible && change.C.Name != ""; reversible {
					reverse = append(reverse, &DropCheck{C: change.C})
				}
			case *DropCheck:
				b.P("DROP CONSTRAINT").Ident(change.C.Name)
				reverse = append(reverse, &AddCheck{C: change.C})
			case *ModifyCheck:
				switch {
				case change.From.Name == "":
					return errors.New("cannot modify unnamed check constraint")
				case change.From.Name != change.To.Name:
					return fmt.Errorf("mismatch check constraint names: %q != %q", change.From.Name, change.To.Name)
				case change.From.Expr != change.To.Expr,
					has(change.From.Attrs, &NoInherit{}) && !has(change.To.Attrs, &NoInherit{}),
					!has(change.From.Attrs, &NoInherit{}) && has(change.To.Attrs, &NoInherit{}):
					b.P("DROP CONSTRAINT").Ident(change.From.Name).Comma().P("ADD")
					check(b, change.To)
				default:
					return errors.New("unknown check constraint change")
				}
				reverse = append(reverse, &ModifyCheck{
					From: change.To,
					To:   change.From,
				})
			}
			return nil
		})
		if err != nil {
			return "", err
		}
		return b.String(), nil
	}
	cmd := &alterChange{}
	stmt, err := build(cmd, changes)
	if err != nil {
		return fmt.Errorf("alter table %q: %v", t.Name, err)
	}
	cmd.main = &Change{
		Cmd: stmt,
		Source: &ModifyTable{
			T:       t,
			Changes: changes,
		},
		Comment: fmt.Sprintf("modify %q table", t.Name),
	}
	if reversible {
		// Changes should be reverted in
		// a reversed order they were created.
		ReverseChanges(reverse)
		if cmd.main.Reverse, err = build(&alterChange{}, reverse); err != nil {
			return fmt.Errorf("reverse alter table %q: %v", t.Name, err)
		}
	}
	cmd.append(s)
	return nil
}

// alterChange describes an alter table Change where its main command
// can be supported by additional statements before and after it is executed.
type alterChange struct {
	main          *Change
	before, after []*Change
}

func (a *alterChange) append(s *state) {
	s.append(a.before...)
	s.append(a.main)
	s.append(a.after...)
}

func (s *state) alterColumn(b *sqlutil.Builder, alter *alterChange, t *Table, c *ModifyColumn) error {
	for k := c.Change; !k.Is(NoChange); {
		b.P("ALTER COLUMN").Ident(c.To.Name)
		switch {
		case k.Is(ChangeType):
			if err := s.alterType(b, alter, t, c); err != nil {
				return err
			}
			k &= ^ChangeType
		case k.Is(ChangeNullability) && c.To.Type.Nullable:
			if t, ok := c.To.Type.Type.(*SerialType); ok {
				return fmt.Errorf("NOT NULL constraint is required for %s column %q", t.T, c.To.Name)
			}
			b.P("DROP NOT NULL")
			k &= ^ChangeNullability
		case k.Is(ChangeNullability) && !c.To.Type.Nullable:
			b.P("SET NOT NULL")
			k &= ^ChangeNullability
		case k.Is(ChangeDefault) && c.To.Default == nil:
			b.P("DROP DEFAULT")
			k &= ^ChangeDefault
		case k.Is(ChangeDefault) && c.To.Default != nil:
			s.columnDefault(b.P("SET"), c.To)
			k &= ^ChangeDefault
		case k.Is(ChangeAttr):
			toI, ok := identity(c.To.Attrs)
			if !ok {
				return fmt.Errorf("unexpected attribute change (expect IDENTITY): %v", c.To.Attrs)
			}
			// The syntax for altering identity columns is identical to sequence_options.
			// https://www.postgresql.org/docs/current/sql-altersequence.html
			b.P("SET GENERATED", toI.Generation, "SET START WITH", strconv.FormatInt(toI.Sequence.Start, 10), "SET INCREMENT BY", strconv.FormatInt(toI.Sequence.Increment, 10))
			// Skip SEQUENCE RESTART in case the "start value" is less than the "current value" in one
			// of the states (inspected and desired), because this function is used for both UP and DOWN.
			if fromI, ok := identity(c.From.Attrs); (!ok || fromI.Sequence.Last < toI.Sequence.Start) && toI.Sequence.Last < toI.Sequence.Start {
				b.P("RESTART")
			}
			k &= ^ChangeAttr
		case k.Is(ChangeGenerated):
			if has(c.To.Attrs, &GeneratedExpr{}) {
				return fmt.Errorf("unexpected generation expression change (expect DROP EXPRESSION): %v", c.To.Attrs)
			}
			b.P("DROP EXPRESSION")
			k &= ^ChangeGenerated
		default: // e.g. ChangeComment.
			return fmt.Errorf("unexpected column change: %d", k)
		}
		if !k.Is(NoChange) {
			b.Comma()
		}
	}
	return nil
}

// alterType appends the clause(s) to alter the column type and assuming the
// "ALTER COLUMN <Name>" was called before by the alterColumn function.
func (s *state) alterType(b *sqlutil.Builder, alter *alterChange, t *Table, c *ModifyColumn) error {
	// Commands for creating and dropping serial sequences.
	createDropSeq := func(st *SerialType) (string, string, string) {
		seq := fmt.Sprintf(`%s%q`, s.schemaPrefix(t.Schema), st.sequence(t, c.To))
		drop := s.Build("DROP SEQUENCE IF EXISTS").P(seq).String()
		create := s.Build("CREATE SEQUENCE IF NOT EXISTS").P(seq, "OWNED BY").
			P(fmt.Sprintf(`%s%q.%q`, s.schemaPrefix(t.Schema), t.Name, c.To.Name)).
			String()
		return create, drop, seq
	}
	toS, toHas := c.To.Type.Type.(*SerialType)
	fromS, fromHas := c.From.Type.Type.(*SerialType)
	switch {
	// Sequence was dropped.
	case fromHas && !toHas:
		b.P("DROP DEFAULT")
		create, drop, _ := createDropSeq(fromS)
		// Sequence should be deleted after it was dropped
		// from the DEFAULT value.
		alter.after = append(alter.after, &Change{
			Source:  c,
			Comment: fmt.Sprintf("drop sequence used by serial column %q", c.From.Name),
			Cmd:     drop,
			Reverse: create,
		})
		toT, err := FormatType(c.To.Type.Type)
		if err != nil {
			return err
		}
		fromT, err := FormatType(fromS.IntegerType())
		if err != nil {
			return err
		}
		// Underlying type was changed. e.g. serial to bigint.
		if toT != fromT {
			b.Comma().P("ALTER COLUMN").Ident(c.To.Name).P("TYPE", toT)
		}
	// Sequence was added.
	case !fromHas && toHas:
		create, drop, seq := createDropSeq(toS)
		// Sequence should be created before it is used by the
		// column DEFAULT value.
		alter.before = append(alter.before, &Change{
			Source:  c,
			Comment: fmt.Sprintf("create sequence for serial column %q", c.To.Name),
			Cmd:     create,
			Reverse: drop,
		})
		b.P("SET DEFAULT", fmt.Sprintf("nextval('%s')", seq))
		toT, err := FormatType(toS.IntegerType())
		if err != nil {
			return err
		}
		fromT, err := FormatType(c.From.Type.Type)
		if err != nil {
			return err
		}
		// Underlying type was changed. e.g. integer to bigserial (bigint).
		if toT != fromT {
			b.Comma().P("ALTER COLUMN").Ident(c.To.Name).P("TYPE", toT)
		}
	// Serial type was changed. e.g. serial to bigserial.
	case fromHas && toHas:
		f, err := FormatType(toS.IntegerType())
		if err != nil {
			return err
		}
		b.P("TYPE", f)
	default:
		var (
			f   string
			err error
		)
		if e, ok := c.To.Type.Type.(*EnumType); ok {
			f = s.enumIdent(t.Schema, e)
		} else if f, err = FormatType(c.To.Type.Type); err != nil {
			return err
		}
		b.P("TYPE", f)
	}
	if collate := (Collation{}); has(c.To.Attrs, &collate) {
		b.P("COLLATE", collate.Value)
	}
	return nil
}

func (s *state) renameTable(c *RenameTable) {
	s.append(&Change{
		Source:  c,
		Comment: fmt.Sprintf("rename a table from %q to %q", c.From.Name, c.To.Name),
		Cmd:     s.Build("ALTER TABLE").Table(tname(c.From)).P("RENAME TO").Table(tname(c.To)).String(),
		Reverse: s.Build("ALTER TABLE").Table(tname(c.To)).P("RENAME TO").Table(tname(c.From)).String(),
	})
}

func (s *state) addComments(t *Table) {
	var c Comment
	if has(t.Attrs, &c) && c.Text != "" {
		s.append(s.tableComment(t, c.Text, ""))
	}
	for i := range t.Columns {
		if has(t.Columns[i].Attrs, &c) && c.Text != "" {
			s.append(s.columnComment(t, t.Columns[i], c.Text, ""))
		}
	}
	for i := range t.Indexes {
		if has(t.Indexes[i].Attrs, &c) && c.Text != "" {
			s.append(s.indexComment(t, t.Indexes[i], c.Text, ""))
		}
	}
}

func (s *state) tableComment(t *Table, to, from string) *Change {
	b := s.Build("COMMENT ON TABLE").Table(tname(t)).P("IS")
	return &Change{
		Cmd:     b.Clone().P(quote(to)).String(),
		Comment: fmt.Sprintf("set comment to table: %q", t.Name),
		Reverse: b.Clone().P(quote(from)).String(),
	}
}

func (s *state) columnComment(t *Table, c *Column, to, from string) *Change {
	b := s.Build("COMMENT ON COLUMN").Table(tname(t))
	b.WriteByte('.')
	b.Ident(c.Name).P("IS")
	return &Change{
		Cmd:     b.Clone().P(quote(to)).String(),
		Comment: fmt.Sprintf("set comment to column: %q on table: %q", c.Name, t.Name),
		Reverse: b.Clone().P(quote(from)).String(),
	}
}

func (s *state) indexComment(t *Table, idx *Index, to, from string) *Change {
	b := s.Build("COMMENT ON INDEX").Ident(idx.Name).P("IS")
	return &Change{
		Cmd:     b.Clone().P(quote(to)).String(),
		Comment: fmt.Sprintf("set comment to index: %q on table: %q", idx.Name, t.Name),
		Reverse: b.Clone().P(quote(from)).String(),
	}
}

func (s *state) dropIndexes(t *Table, indexes ...*Index) {
	rs := &state{}
	rs.addIndexes(t, indexes...)
	for i, idx := range indexes {
		s.append(&Change{
			Cmd:     rs.Changes[i].Reverse,
			Comment: fmt.Sprintf("drop index %q from table: %q", idx.Name, t.Name),
			Reverse: rs.Changes[i].Cmd,
		})
	}
}

// func (s *state) mayAddEnums(t *Table, columns ...*Column) error {
// 	for _, c := range columns {
// 		e, ok := hasEnumType(c)
// 		if !ok {
// 			continue
// 		}
// 		if e.T == "" {
// 			return fmt.Errorf("missing enum name for column %q", c.Name)
// 		}
// 		if exists, err := s.enumExists(t.Schema, e); err != nil {
// 			return err
// 		} else if exists {
// 			// Enum exists and was not created
// 			// on this migration phase.
// 			continue
// 		}
// 		name := s.enumIdent(t.Schema, e)
// 		if prev, ok := s.created[name]; ok {
// 			if !sqlutil.ValuesEqual(prev.Values, e.Values) {
// 				return fmt.Errorf("enum type %s has inconsistent desired state: %q != %q", name, prev.Values, e.Values)
// 			}
// 			continue
// 		}
// 		s.created[name] = e
// 		create, drop := s.createDropEnum(t.Schema, e)
// 		s.append(&Change{
// 			Cmd:     create,
// 			Reverse: drop,
// 			Comment: fmt.Sprintf("create enum type %q", e.T),
// 		})
// 	}
// 	return nil
// }

// func (s *state) alterEnum(from, to *EnumType) error {
// 	if len(from.Values) > len(to.Values) {
// 		return fmt.Errorf("dropping enum (%q) value is not supported", from.T)
// 	}
// 	for i := range from.Values {
// 		if from.Values[i] != to.Values[i] {
// 			return fmt.Errorf("replacing or reordering enum (%q) value is not supported: %q != %q", to.T, to.Values, from.Values)
// 		}
// 	}
// 	name := s.enumIdent(from.Schema, from)
// 	if prev, ok := s.altered[name]; ok {
// 		if !sqlutil.ValuesEqual(prev.Values, to.Values) {
// 			return fmt.Errorf("enum type %s has inconsistent desired state: %q != %q", name, prev.Values, to.Values)
// 		}
// 		return nil
// 	}
// 	s.altered[name] = to
// 	for _, v := range to.Values[len(from.Values):] {
// 		s.append(&Change{
// 			Cmd:     s.Build("ALTER TYPE").P(name, "ADD VALUE", quote(v)).String(),
// 			Comment: fmt.Sprintf("add value to enum type: %q", from.T),
// 		})
// 	}
// 	return nil
// }

// func (s *state) enumExists(ns *Schema, e *EnumType) (bool, error) {
// 	_, ok := ns.Enum(e.T)
// 	return ok, nil
// 	// query, args := `SELECT * FROM pg_type t JOIN pg_namespace n on t.typnamespace = n.oid WHERE t.typname = $1 AND t.typtype = 'e'`, []any{e.T}
// 	// if es := s.enumSchema(ns, e); es != "" {
// 	// 	query += " AND n.nspname = $2"
// 	// 	args = append(args, es)
// 	// }
// 	// rows, err := s.QueryContext(ctx, query, args...)
// 	// if err != nil {
// 	// 	return false, fmt.Errorf("check enum existence: %w", err)
// 	// }
// 	// defer rows.Close()
// 	// return rows.Next(), rows.Err()
// }

// // mayDropEnum drops dangling enum types form the
// func (s *state) mayDropEnum(alter *alterChange, ns *Schema, e *EnumType) error {
// 	name := s.enumIdent(ns, e)
// 	if _, ok := s.dropped[name]; ok {
// 		return nil
// 	}
// 	schemas := []*Schema{ns}
// 	// In case there is a realm attached, traverse the entire tree.
// 	if ns.Realm != nil && len(ns.Realm.Schemas) > 0 {
// 		schemas = ns.Realm.Schemas
// 	}
// 	for i := range schemas {
// 		for _, t := range schemas[i].Tables {
// 			for _, c := range t.Columns {
// 				e1, ok := hasEnumType(c)
// 				// Although we search in siblings schemas, use the
// 				// table's one for building the enum identifier.
// 				if ok && s.enumIdent(ns, e1) == name {
// 					return nil
// 				}
// 			}
// 		}
// 	}
// 	s.dropped[name] = e
// 	create, drop := s.createDropEnum(ns, e)
// 	alter.after = append(alter.after, &Change{
// 		Cmd:     drop,
// 		Reverse: create,
// 		Comment: fmt.Sprintf("drop enum type %q", e.T),
// 	})
// 	return nil
// }

func (s *state) addIndexes(t *Table, indexes ...*Index) {
	for _, idx := range indexes {
		b := s.Build("CREATE")
		if idx.Unique {
			b.P("UNIQUE")
		}
		b.P("INDEX")
		if idx.Name != "" {
			b.Ident(idx.Name)
		}
		b.P("ON").Table(tname(t))
		s.index(b, idx)
		s.append(&Change{
			Cmd:     b.String(),
			Comment: fmt.Sprintf("create index %q to table: %q", idx.Name, t.Name),
			Reverse: func() string {
				b := s.Build("DROP INDEX")
				// Unlike MySQL, the DROP command is not attached to ALTER TABLE.
				// Therefore, we print indexes with their qualified name, because
				// the connection that executes the statements may not be attached
				// to this
				if t.Schema != nil {
					b.WriteString(s.schemaPrefix(t.Schema))
				}
				b.Ident(idx.Name)
				return b.String()
			}(),
		})
	}
}

func (s *state) column(b *sqlutil.Builder, t *Table, c *Column) error {
	f, err := s.formatType(t, c)
	if err != nil {
		return err
	}
	b.Ident(c.Name).P(f)
	if !c.Type.Nullable {
		b.P("NOT")
	} else if t, ok := c.Type.Type.(*SerialType); ok {
		return fmt.Errorf("NOT NULL constraint is required for %s column %q", t.T, c.Name)
	}
	b.P("NULL")
	s.columnDefault(b, c)
	for _, attr := range c.Attrs {
		switch a := attr.(type) {
		case *Comment:
		case *Collation:
			b.P("COLLATE").Ident(a.Value)
		case *Identity, *GeneratedExpr:
			// Handled below.
		default:
			return fmt.Errorf("unexpected column attribute: %T", attr)
		}
	}
	switch hasI, hasX := has(c.Attrs, &Identity{}), has(c.Attrs, &GeneratedExpr{}); {
	case hasI && hasX:
		return fmt.Errorf("both identity and generation expression specified for column %q", c.Name)
	case hasI:
		id, _ := identity(c.Attrs)
		b.P("GENERATED", id.Generation, "AS IDENTITY")
		if id.Sequence.Start != defaultSeqStart || id.Sequence.Increment != defaultSeqIncrement {
			b.Wrap(func(b *sqlutil.Builder) {
				if id.Sequence.Start != defaultSeqStart {
					b.P("START WITH", strconv.FormatInt(id.Sequence.Start, 10))
				}
				if id.Sequence.Increment != defaultSeqIncrement {
					b.P("INCREMENT BY", strconv.FormatInt(id.Sequence.Increment, 10))
				}
			})
		}
	case hasX:
		x := &GeneratedExpr{}
		has(c.Attrs, x)
		b.P("GENERATED ALWAYS AS", sqlutil.MayWrap(x.Expr), "STORED")
	}
	return nil
}

// columnDefault writes the default value of column to the builder.
func (s *state) columnDefault(b *sqlutil.Builder, c *Column) {
	switch x := c.Default.(type) {
	case *LiteralExpr:
		v := x.Value
		switch c.Type.Type.(type) {
		case *BoolType, *DecimalType, *IntegerType, *FloatType:
		default:
			v = quote(v)
		}
		b.P("DEFAULT", v)
	case *RawExpr:
		// Ignore identity functions added by the differ.
		if _, ok := c.Type.Type.(*SerialType); !ok {
			b.P("DEFAULT", x.Expr)
		}
	}
}

func (s *state) indexParts(b *sqlutil.Builder, parts []*IndexPart) {
	b.Wrap(func(b *sqlutil.Builder) {
		b.MapComma(parts, func(i int, b *sqlutil.Builder) {
			switch part := parts[i]; {
			case part.Column != nil:
				b.Ident(part.Column.Name)
			case part.Expr != nil:
				b.WriteString(sqlutil.MayWrap(part.Expr.(*RawExpr).Expr))
			}
			s.partAttrs(b, parts[i])
		})
	})
}

func (s *state) partAttrs(b *sqlutil.Builder, p *IndexPart) {
	if p.Descending {
		b.P("DESC")
	}
	for _, attr := range p.Attrs {
		switch attr := attr.(type) {
		case *IndexColumnProperty:
			switch {
			// Defaults when DESC is specified.
			case p.Descending && attr.NullsFirst:
			case p.Descending && attr.NullsLast:
				b.P("NULL LAST")
			// Defaults when DESC is not specified.
			case !p.Descending && attr.NullsLast:
			case !p.Descending && attr.NullsFirst:
				b.P("NULL FIRST")
			}
		case *Collation:
			b.P("COLLATE").Ident(attr.Value)
		default:
			panic(fmt.Sprintf("unexpected index part attribute: %T", attr))
		}
	}
}

func (s *state) index(b *sqlutil.Builder, idx *Index) {
	// Avoid appending the default method.
	if t := (IndexType{}); has(idx.Attrs, &t) && strings.ToUpper(t.T) != IndexTypeBTree {
		b.P("USING", t.T)
	}
	s.indexParts(b, idx.Parts)
	if c := (IndexInclude{}); has(idx.Attrs, &c) {
		b.P("INCLUDE")
		b.Wrap(func(b *sqlutil.Builder) {
			b.MapComma(c.Columns, func(i int, b *sqlutil.Builder) {
				b.Ident(c.Columns[i])
			})
		})
	}
	if p, ok := indexStorageParams(idx.Attrs); ok {
		b.P("WITH")
		b.Wrap(func(b *sqlutil.Builder) {
			var parts []string
			if p.AutoSummarize {
				parts = append(parts, "autosummarize = true")
			}
			if p.PagesPerRange != 0 && p.PagesPerRange != DefaultPagePerRange {
				parts = append(parts, fmt.Sprintf("pages_per_range = %d", p.PagesPerRange))
			}
			b.WriteString(strings.Join(parts, ", "))
		})
	}
	if p := (IndexPredicate{}); has(idx.Attrs, &p) {
		b.P("WHERE").P(p.Predicate)
	}
	for _, attr := range idx.Attrs {
		switch attr.(type) {
		case *Comment, *ConstraintType, *IndexType, *IndexInclude, *IndexPredicate, *IndexStorageParams:
		default:
			panic(fmt.Sprintf("unexpected index attribute: %T", attr))
		}
	}
}

func (s *state) fks(b *sqlutil.Builder, fks ...*ForeignKey) {
	b.MapComma(fks, func(i int, b *sqlutil.Builder) {
		fk := fks[i]
		if fk.Name != "" {
			b.P("CONSTRAINT").Ident(fk.Name)
		}
		b.P("FOREIGN KEY")
		b.Wrap(func(b *sqlutil.Builder) {
			b.MapComma(fk.Columns, func(i int, b *sqlutil.Builder) {
				b.Ident(fk.Columns[i].Name)
			})
		})
		b.P("REFERENCES").Table(tname(fk.RefTable))
		b.Wrap(func(b *sqlutil.Builder) {
			b.MapComma(fk.RefColumns, func(i int, b *sqlutil.Builder) {
				b.Ident(fk.RefColumns[i].Name)
			})
		})
		if fk.OnUpdate != "" {
			b.P("ON UPDATE", string(fk.OnUpdate))
		}
		if fk.OnDelete != "" {
			b.P("ON DELETE", string(fk.OnDelete))
		}
	})
}

func (s *state) append(c ...*Change) {
	s.Changes = append(s.Changes, c...)
}

// Build instantiates a new builder and writes the given phrase to it.
func (s *state) Build(phrases ...string) *sqlutil.Builder {
	b := &sqlutil.Builder{QuoteChar: '"', Schema: s.SchemaQualifier}
	return b.P(phrases...)
}

// skipAutoChanges filters unnecessary changes that are automatically
// happened by the database when ALTER TABLE is executed.
func skipAutoChanges(changes []SchemaChange) []SchemaChange {
	var (
		dropC   = make(map[string]bool)
		planned = make([]SchemaChange, 0, len(changes))
	)
	for _, c := range changes {
		if c, ok := c.(*DropColumn); ok {
			dropC[c.C.Name] = true
		}
	}
search:
	for _, c := range changes {
		switch c := c.(type) {
		// Indexes involving the column are automatically dropped
		// with it. This is true for multi-columns indexes as well.
		// See https://www.postgresql.org/docs/current/sql-altertable.html
		case *DropIndex:
			for _, p := range c.I.Parts {
				if p.Column != nil && dropC[p.Column.Name] {
					continue search
				}
			}
		// Simple case for skipping constraint dropping,
		// if the child table columns were dropped.
		case *DropForeignKey:
			for _, c := range c.F.Columns {
				if dropC[c.Name] {
					continue search
				}
			}
		}
		planned = append(planned, c)
	}
	return planned
}

// commentChange extracts the information for modifying a comment from the given change.
func commentChange(c SchemaChange) (from, to string, err error) {
	switch c := c.(type) {
	case *AddAttr:
		toC, ok := c.A.(*Comment)
		if ok {
			to = toC.Text
			return
		}
		err = fmt.Errorf("unexpected AddAttr.(%T) for comment change", c.A)
	case *ModifyAttr:
		fromC, ok1 := c.From.(*Comment)
		toC, ok2 := c.To.(*Comment)
		if ok1 && ok2 {
			from, to = fromC.Text, toC.Text
			return
		}
		err = fmt.Errorf("unsupported ModifyAttr(%T, %T) change", c.From, c.To)
	default:
		err = fmt.Errorf("unexpected change %T", c)
	}
	return
}

// checks writes the CHECK constraint to the builder.
func check(b *sqlutil.Builder, c *Check) {
	if c.Name != "" {
		b.P("CONSTRAINT").Ident(c.Name)
	}
	b.P("CHECK", sqlutil.MayWrap(c.Expr))
	if has(c.Attrs, &NoInherit{}) {
		b.P("NO INHERIT")
	}
}

// isUniqueConstraint reports if the index is a valid UNIQUE constraint.
func isUniqueConstraint(i *Index) bool {
	if c := (ConstraintType{}); !has(i.Attrs, &c) || !c.IsUnique() || !i.Unique {
		return false
	}
	// UNIQUE constraint cannot use functional indexes,
	// and all its parts must have the default sort ordering.
	for _, p := range i.Parts {
		if p.Expr != nil || p.Descending {
			return false
		}
	}
	for _, a := range i.Attrs {
		switch a := a.(type) {
		// UNIQUE constraints must have BTREE type indexes.
		case *IndexType:
			if strings.ToUpper(a.T) != IndexTypeBTree {
				return false
			}
		// Partial indexes are not allowed.
		case *IndexPredicate:
			return false
		}
	}
	return true
}

func quote(s string) string {
	if sqlutil.IsQuoted(s, '\'') {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

func (s *state) createDropEnum(sc *Schema, e *EnumType) (string, string) {
	name := s.enumIdent(sc, e)
	return s.Build("CREATE TYPE").
			P(name, "AS ENUM").
			Wrap(func(b *sqlutil.Builder) {
				b.MapComma(e.Values, func(i int, b *sqlutil.Builder) {
					b.WriteString(quote(e.Values[i]))
				})
			}).
			String(),
		s.Build("DROP TYPE").P(name).String()
}

func (s *state) enumIdent(ns *Schema, e *EnumType) string {
	es := s.enumSchema(ns, e)
	if es != "" {
		return fmt.Sprintf("%q.%q", es, e.T)
	}
	return strconv.Quote(e.T)
}

func (s *state) enumSchema(ns *Schema, e *EnumType) (es string) {
	switch {
	// In case the plan uses a specific schema qualifier.
	case s.SchemaQualifier != nil:
		es = *s.SchemaQualifier
	// Enum schema has higher precedence.
	case e.Schema != nil:
		es = e.Schema.Name
	// Fallback to table schema if exists.
	case ns != nil:
		es = ns.Name
	}
	return
}

// schemaPrefix returns the schema prefix based on the planner config.
func (s *state) schemaPrefix(ns *Schema) string {
	switch {
	case s.SchemaQualifier != nil:
		// In case the qualifier is empty, ignore.
		if *s.SchemaQualifier != "" {
			return fmt.Sprintf("%q.", *s.SchemaQualifier)
		}
	case ns != nil && ns.Name != "":
		return fmt.Sprintf("%q.", ns.Name)
	}
	return ""
}

// formatType formats the type but takes into account the qualifier.
func (s *state) formatType(t *Table, c *Column) (string, error) {
	switch tt := c.Type.Type.(type) {
	case *EnumType:
		return s.enumIdent(t.Schema, tt), nil
	case *ArrayType:
		if e, ok := tt.Type.(*EnumType); ok {
			return s.enumIdent(t.Schema, e) + "[]", nil
		}
	}
	return FormatType(c.Type.Type)
}

// func hasEnumType(c *Column) (*EnumType, bool) {
// 	switch t := c.Type.Type.(type) {
// 	case *EnumType:
// 		return t, true
// 	case *ArrayType:
// 		if e, ok := t.Type.(*EnumType); ok {
// 			return e, true
// 		}
// 	}
// 	return nil, false
// }

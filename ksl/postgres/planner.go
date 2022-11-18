package postgres

import (
	"context"
	"fmt"
	"ksl"
	"ksl/sqlschema"
	"strings"
)

type Planner struct{}

func (p Planner) Plan(migration sqlschema.Migration) (sqlschema.MigrationPlan, error) {
	return p.PlanContext(context.Background(), migration)
}

func (p Planner) PlanContext(ctx context.Context, migration sqlschema.Migration) (sqlschema.MigrationPlan, error) {
	s := &planctx{
		plan:      sqlschema.MigrationPlan{},
		migration: migration,
	}
	if err := s.planMigration(); err != nil {
		return sqlschema.MigrationPlan{}, err
	}
	return s.plan, nil
}

type planctx struct {
	plan      sqlschema.MigrationPlan
	migration sqlschema.Migration
}

func (s *planctx) planMigration() error {
	pair := sqlschema.MakePair(s.migration.Before, s.migration.After)

	for _, change := range s.migration.Changes {
		stmt, err := RenderStep(pair, change)
		if err != nil {
			return err
		}
		s.plan.Statements = append(s.plan.Statements, stmt)
	}
	return nil
}

func RenderStep(dbs sqlschema.Pair[sqlschema.Database], step sqlschema.MigrationStep) (sqlschema.Statement, error) {
	switch step := step.(type) {
	case sqlschema.AlterEnum:
		return renderAlterEnum(dbs, step)
	case sqlschema.CreateEnum:
		return renderCreateEnum(dbs, step)
	case sqlschema.DropEnum:
		return renderDropEnum(dbs, step)
	case sqlschema.CreateTable:
		return renderCreateTable(dbs, step)
	case sqlschema.DropTable:
		return renderDropTable(dbs, step)
	case sqlschema.AddForeignKey:
		return renderAddForeignKey(dbs, step)
	case sqlschema.DropForeignKey:
		return renderDropForeignKey(dbs, step)
	case sqlschema.AlterTable:
		return renderAlterTable(dbs, step)
	case sqlschema.CreateIndex:
		return renderCreateIndex(dbs, step)
	case sqlschema.DropIndex:
		return renderDropIndex(dbs, step)
	case sqlschema.RenameIndex:
		return renderRenameIndex(dbs, step)
	case sqlschema.RenameForeignKey:
		return renderRenameForeignKey(dbs, step)
	case sqlschema.CreateExtension:
		// return renderCreateExtension(dbs, step)
	case sqlschema.AlterExtension:
		// return renderAlterExtension(dbs, step)
	case sqlschema.DropExtension:
		// return renderDropExtension(dbs, step)
	default:
		return sqlschema.Statement{}, fmt.Errorf("sqlschema: unknown migration step %T", step)
	}
	return sqlschema.Statement{Comment: "empty", Steps: []sqlschema.Step{{Cmd: "empty"}}}, nil
}

func renderAlterEnum(pair sqlschema.Pair[sqlschema.Database], step sqlschema.AlterEnum) (sqlschema.Statement, error) {
	if len(step.DroppedVariants) == 0 {
		var steps []sqlschema.Step
		for _, v := range step.CreatedVariants {
			name := pair.Prev.WalkEnum(step.Enums.Prev).Name()
			steps = append(steps, sqlschema.Step{
				Cmd:     fmt.Sprintf("ALTER TYPE %s ADD VALUE %s", quoteIdent(name), quoteString(v)),
				Comment: fmt.Sprintf("Add variant %q to enum %q.", quoteString(v), quoteIdent(name)),
			})
		}
		return sqlschema.Statement{Steps: steps}, nil
	}

	enums := sqlschema.MakePair(pair.Prev.WalkEnum(step.Enums.Prev), pair.Next.WalkEnum(step.Enums.Next))
	comment := fmt.Sprintf("Alter enum %s", quoteIdent(enums.Prev.Name()))
	if len(step.CreatedVariants) > 0 {
		comment += fmt.Sprintf(", adding variants %s", quoteString(step.CreatedVariants...))
	}
	if len(step.DroppedVariants) > 0 {
		comment += fmt.Sprintf(", dropping variants %s", quoteString(step.DroppedVariants...))
	}
	comment += "."

	stmt := sqlschema.Statement{Comment: comment}

	tmpName := enums.Next.Name() + "_tmp"
	tmpOldName := enums.Prev.Name() + "_old"

	// begin transaction
	stmt.Steps.Add(sqlschema.Step{
		Cmd:     "BEGIN",
		Comment: "Begin transaction.",
	})

	// create a new enum with the new name
	stmt.Steps.Add(sqlschema.Step{
		Cmd:     fmt.Sprintf("CREATE TYPE %s AS ENUM (%s)", quoteIdent(tmpName), quoteString(enums.Next.Values()...)),
		Comment: fmt.Sprintf("Create new enum %s with variants %s.", quoteIdent(tmpName), quoteString(enums.Next.Values()...)),
	})

	// TODO: find defaults using the old enum

	// alter columns using the old enum to use the new enum
	for _, col := range pair.Next.WalkColumns() {
		if e, ok := col.Type().Type.(sqlschema.EnumType); ok && e.ID == enums.Next.ID {
			var array string
			if col.Arity() == sqlschema.List {
				array = "[]"
			}

			stmt.Steps.Add(sqlschema.Step{
				Cmd: fmt.Sprintf(
					"ALTER TABLE %s ALTER COLUMN %s TYPE %s%s USING %s::text::%s%s",
					quoteIdent(col.Table().Name()),
					quoteIdent(col.Name()),
					quoteIdent(tmpName), array,
					quoteIdent(col.Name()),
					quoteIdent(tmpName), array,
				),
				Comment: fmt.Sprintf("Alter column %s to use new enum %s.", quoteIdent(col.Name()), quoteIdent(tmpName)),
			})
		}
	}

	// rename old enum
	stmt.Steps.Add(sqlschema.Step{
		Cmd:     fmt.Sprintf("ALTER TYPE %s RENAME TO %s", quoteIdent(enums.Prev.Name()), quoteIdent(tmpOldName)),
		Comment: fmt.Sprintf("Rename old enum %s to %s.", quoteIdent(enums.Prev.Name()), quoteIdent(tmpOldName)),
	})

	// rename new enum
	stmt.Steps.Add(sqlschema.Step{
		Cmd:     fmt.Sprintf("ALTER TYPE %s RENAME TO %s", quoteIdent(tmpName), quoteIdent(enums.Next.Name())),
		Comment: fmt.Sprintf("Rename new enum %s to %s.", quoteIdent(tmpName), quoteIdent(enums.Next.Name())),
	})

	// drop old enum
	stmt.Steps.Add(sqlschema.Step{
		Cmd:     fmt.Sprintf("DROP TYPE %s", quoteIdent(tmpOldName)),
		Comment: fmt.Sprintf("Drop old enum %s.", quoteIdent(tmpOldName)),
	})

	// TODO: reinstall dropped defaults

	// finish transaction
	stmt.Steps.Add(sqlschema.Step{Cmd: "COMMIT", Comment: "Commit transaction."})

	return stmt, nil
}

func renderCreateEnum(pair sqlschema.Pair[sqlschema.Database], step sqlschema.CreateEnum) (sqlschema.Statement, error) {
	enum := pair.Next.WalkEnum(step.Enum)
	stmt := sqlschema.Statement{Comment: fmt.Sprintf("Create enum %s with variants %s.", quoteIdent(enum.Name()), quoteString(enum.Values()...))}
	stmt.Steps.Add(sqlschema.Step{Cmd: fmt.Sprintf("CREATE TYPE %s AS ENUM (%s)", quoteIdent(enum.Name()), quoteString(enum.Values()...))})
	return stmt, nil
}

func renderDropEnum(pair sqlschema.Pair[sqlschema.Database], step sqlschema.DropEnum) (sqlschema.Statement, error) {
	enum := pair.Prev.WalkEnum(step.Enum)
	stmt := sqlschema.Statement{Comment: fmt.Sprintf("Drop enum %s.", quoteIdent(enum.Name()))}
	stmt.Steps.Add(sqlschema.Step{Cmd: fmt.Sprintf("DROP TYPE %s", quoteIdent(enum.Name()))})
	return stmt, nil
}

func renderCreateTable(pair sqlschema.Pair[sqlschema.Database], step sqlschema.CreateTable) (sqlschema.Statement, error) {
	table := pair.Next.WalkTable(step.Table)
	stmt := sqlschema.Statement{Comment: fmt.Sprintf("Create table %s.", quoteIdent(table.Name()))}
	var columnstr string
	for i, col := range table.Columns() {
		columnstr += sqlschema.SQLIndentation + renderColumn(col)
		if i < len(table.Columns())-1 {
			columnstr += ",\n"
		}
	}

	var primaryKey string
	if pk, ok := table.PrimaryKey().Get(); ok {
		named := fmt.Sprintf("CONSTRAINT %s ", quoteIdent(pk.Name()))
		primaryKey = fmt.Sprintf(",\n\n%s%sPRIMARY KEY (%s)", sqlschema.SQLIndentation, named, quoteIdent(pk.ColumnNames()...))
	}

	stmt.Steps.Add(sqlschema.Step{Cmd: fmt.Sprintf("CREATE TABLE %s (\n%s%s\n)", quoteIdent(table.Name()), columnstr, primaryKey)})
	return stmt, nil
}

func renderColumn(col sqlschema.ColumnWalker) string {
	var builder strings.Builder
	builder.WriteString(quoteIdent(col.Name()))
	builder.WriteString(" ")
	builder.WriteString(renderColumnType(col))
	if col.IsRequired() {
		builder.WriteString(" NOT NULL")
	}
	return builder.String()
}

func renderColumnType(col sqlschema.ColumnWalker) string {
	var array string
	if col.Arity() == sqlschema.List {
		array = "[]"
	}

	if enum, ok := col.EnumType().Get(); ok {
		return fmt.Sprintf("%s%s", quoteIdent(enum.Name()), array)
	}

	switch t := col.Type().Type.(type) {
	case PostgresType:
		return fmt.Sprintf("%s%s", t.String(), array)
	case ksl.BuiltInScalar:
		return fmt.Sprintf("%s%s", DefaultNativeTypeForScalar(t).String(), array)
	default:
		panic(fmt.Sprintf("unknown type %T", t))
	}
}

func renderDropTable(pair sqlschema.Pair[sqlschema.Database], step sqlschema.DropTable) (sqlschema.Statement, error) {
	table := pair.Prev.WalkTable(step.Table)
	stmt := sqlschema.Statement{Comment: fmt.Sprintf("Drop table %s.", quoteIdent(table.Name()))}
	stmt.Steps.Add(sqlschema.Step{Cmd: fmt.Sprintf("DROP TABLE %s", quoteIdent(table.Name()))})
	return stmt, nil
}

func renderAddForeignKey(pair sqlschema.Pair[sqlschema.Database], step sqlschema.AddForeignKey) (sqlschema.Statement, error) {
	fk := pair.Next.WalkForeignKey(step.ForeignKey)
	stmt := sqlschema.Statement{Comment: fmt.Sprintf("Add foreign key %s.", quoteIdent(fk.ConstraintName()))}
	stmt.Steps.Add(sqlschema.Step{
		Cmd: fmt.Sprintf(
			"ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s) ON DELETE %s ON UPDATE %s",
			quoteIdent(fk.Table().Name()),
			quoteIdent(fk.ConstraintName()),
			quoteIdent(fk.ConstrainedColumnNames()...),
			quoteIdent(fk.ReferencedTable().Name()),
			quoteIdent(fk.ReferencedColumnNames()...),
			fk.OnDeleteAction().DDL(),
			fk.OnUpdateAction().DDL(),
		),
	})
	return stmt, nil
}

func renderDropForeignKey(pair sqlschema.Pair[sqlschema.Database], step sqlschema.DropForeignKey) (sqlschema.Statement, error) {
	fk := pair.Prev.WalkForeignKey(step.ForeignKey)
	stmt := sqlschema.Statement{Comment: fmt.Sprintf("Drop foreign key %s.", quoteIdent(fk.ConstraintName()))}
	stmt.Steps.Add(sqlschema.Step{
		Cmd: fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", quoteIdent(fk.Table().Name()), quoteIdent(fk.ConstraintName())),
	})
	return stmt, nil
}

func renderAlterTable(pair sqlschema.Pair[sqlschema.Database], step sqlschema.AlterTable) (sqlschema.Statement, error) {
	prev, next := pair.Prev.WalkTable(step.Tables.Prev), pair.Next.WalkTable(step.Tables.Next)
	var lines []string
	var before, after []sqlschema.Step
	var stmt sqlschema.Statement

	for _, change := range step.Changes {
		switch change := change.(type) {
		case sqlschema.DropPrimaryKey:
			lines = append(lines, fmt.Sprintf("DROP CONSTRAINT %s", quoteIdent(prev.PrimaryKey().MustGet().Name())))
		case sqlschema.RenamePrimaryKey:
			lines = append(lines, fmt.Sprintf("RENAME CONSTRAINT %s TO %s", quoteIdent(prev.PrimaryKey().MustGet().Name()), quoteIdent(next.PrimaryKey().MustGet().Name())))
		case sqlschema.AddPrimaryKey:
			var named string
			if pk, ok := next.PrimaryKey().Get(); ok {
				named = fmt.Sprintf(" CONSTRAINT %s", quoteIdent(pk.Name()))
			}
			lines = append(lines, fmt.Sprintf("ADD%s PRIMARY KEY (%s)", named, quoteIdent(next.PrimaryKey().MustGet().ColumnNames()...)))
		case sqlschema.AddColumn:
			column := pair.Next.WalkColumn(change.Column)
			lines = append(lines, fmt.Sprintf("ADD COLUMN %s", renderColumn(column)))
		case sqlschema.AlterColumn:
			pc, nc := pair.Prev.WalkColumn(change.Columns.Prev), pair.Next.WalkColumn(change.Columns.Next)
			b, c, a := renderAlterColumn(sqlschema.MakePair(pc, nc), change.Changes)
			before = append(before, b...)
			lines = append(lines, c...)
			after = append(after, a...)
		case sqlschema.DropColumn:
			column := pair.Prev.WalkColumn(change.Column)
			lines = append(lines, fmt.Sprintf("DROP COLUMN %s", quoteIdent(column.Name())))
		case sqlschema.DropAndRecreateColumn:
			column := pair.Prev.WalkColumn(change.Columns.Prev)
			lines = append(lines, fmt.Sprintf("DROP COLUMN %s", quoteIdent(column.Name())))
			column = pair.Next.WalkColumn(change.Columns.Next)
			lines = append(lines, fmt.Sprintf("ADD COLUMN %s", renderColumn(column)))
		case sqlschema.RenameColumn:
			column := pair.Prev.WalkColumn(change.Columns.Prev)
			newColumn := pair.Next.WalkColumn(change.Columns.Next)
			lines = append(lines, fmt.Sprintf("RENAME COLUMN %s TO %s", quoteIdent(column.Name()), quoteIdent(newColumn.Name())))
		}
	}

	stmt.Steps.Add(before...)
	stmt.Steps.Add(sqlschema.Step{
		Cmd:     fmt.Sprintf("ALTER TABLE %s %s", quoteIdent(prev.Name()), strings.Join(lines, ",\n")),
		Comment: fmt.Sprintf("Alter table %s.", quoteIdent(prev.Name())),
	})
	stmt.Steps.Add(after...)

	return stmt, nil
}

func renderCreateIndex(pair sqlschema.Pair[sqlschema.Database], step sqlschema.CreateIndex) (sqlschema.Statement, error) {
	index := pair.Next.WalkIndex(step.Index)

	var unique string
	if index.IsUnique() {
		unique = "UNIQUE "
	}
	var using string
	if algo := index.Algorithm().String(); algo != "" {
		using = fmt.Sprintf("USING %s ", algo)
	}

	var columnData []string
	for _, column := range index.Columns() {
		name := quoteIdent(column.Name())
		switch column.SortOrder() {
		case sqlschema.Ascending:
			name += " ASC"
		case sqlschema.Descending:
			name += " DESC"
		}
		columnData = append(columnData, name)
	}

	stmt := sqlschema.Statement{Comment: fmt.Sprintf("Create index %s.", quoteIdent(index.Name()))}
	stmt.Steps.Add(sqlschema.Step{
		Cmd: fmt.Sprintf(
			"CREATE %sINDEX %s ON %s %s(%s)",
			unique,
			quoteIdent(index.Name()),
			quoteIdent(index.Table().Name()),
			using,
			strings.Join(columnData, ", "),
		),
	})
	return stmt, nil
}

func renderDropIndex(pair sqlschema.Pair[sqlschema.Database], step sqlschema.DropIndex) (sqlschema.Statement, error) {
	index := pair.Prev.WalkIndex(step.Index)
	stmt := sqlschema.Statement{Comment: fmt.Sprintf("Drop index %s.", quoteIdent(index.Name()))}
	stmt.Steps.Add(sqlschema.Step{Cmd: fmt.Sprintf("DROP INDEX %s", quoteIdent(index.Name()))})
	return stmt, nil
}

func renderRenameIndex(pair sqlschema.Pair[sqlschema.Database], step sqlschema.RenameIndex) (sqlschema.Statement, error) {
	index := pair.Next.WalkIndex(step.Index.Next)
	stmt := sqlschema.Statement{Comment: fmt.Sprintf("Rename index %s.", quoteIdent(index.Name()))}
	stmt.Steps.Add(sqlschema.Step{Cmd: fmt.Sprintf("ALTER INDEX %s RENAME TO %s", quoteIdent(index.Name()), quoteIdent(index.Name()))})
	return stmt, nil
}

func renderRenameForeignKey(pair sqlschema.Pair[sqlschema.Database], step sqlschema.RenameForeignKey) (sqlschema.Statement, error) {
	fk := pair.Next.WalkForeignKey(step.ForeignKeys.Next)
	stmt := sqlschema.Statement{Comment: fmt.Sprintf("Rename foreign key %s.", quoteIdent(fk.ConstraintName()))}
	stmt.Steps.Add(sqlschema.Step{
		Cmd: fmt.Sprintf("ALTER TABLE %s RENAME CONSTRAINT %s TO %s", quoteIdent(fk.Table().Name()), quoteIdent(fk.ConstraintName()), quoteIdent(fk.ConstraintName())),
	})
	return stmt, nil
}

func renderAlterColumn(pair sqlschema.Pair[sqlschema.ColumnWalker], changes sqlschema.ColumnChanges) ([]sqlschema.Step, []string, []sqlschema.Step) {
	acc := expandChanges(pair, changes)

	tableName := quoteIdent(pair.Prev.Table().Name())
	columnName := quoteIdent(pair.Prev.Name())

	alterColumnPrefix := fmt.Sprintf("ALTER COLUMN %s", columnName)

	var clauses []string
	var before, after []sqlschema.Step

	if acc.DropDefault {
		clauses = append(clauses, fmt.Sprintf("%s DROP DEFAULT", alterColumnPrefix))
		// TODO: might need to drop sequences
	}

	if acc.SetDefault != nil {
		clauses = append(clauses, fmt.Sprintf(
			"%s SET DEFAULT %s",
			alterColumnPrefix,
			renderDefault(acc.SetDefault, renderColumnType(pair.Next)),
		))
	}

	if acc.DropNotNull {
		clauses = append(clauses, fmt.Sprintf("%s DROP NOT NULL", alterColumnPrefix))
	}

	if acc.SetNotNull {
		clauses = append(clauses, fmt.Sprintf("%s SET NOT NULL", alterColumnPrefix))
	}

	if acc.SetType {
		clauses = append(clauses, fmt.Sprintf("%s SET DATA TYPE %s", alterColumnPrefix, renderColumnType(pair.Next)))
	}

	if acc.AddSequence {
		seqName := fmt.Sprintf("%s_%s_seq", tableName, columnName)
		before = append(before, sqlschema.Step{
			Cmd:     fmt.Sprintf("CREATE SEQUENCE %s", seqName),
			Comment: fmt.Sprintf("Create sequence %s for column %s.%s.", seqName, tableName, columnName),
		})
		clauses = append(clauses, fmt.Sprintf("%s SET DEFAULT nextval(%s)", alterColumnPrefix, quoteString(seqName)))
		after = append(after, sqlschema.Step{
			Cmd:     fmt.Sprintf("ALTER SEQUENCE %s OWNED BY %s.%s", seqName, tableName, columnName),
			Comment: fmt.Sprintf("Set sequence %s owner to %s.%s.", seqName, tableName, columnName),
		})
	}

	return before, clauses, after
}

func renderDefault(value sqlschema.Value, dataType string) string {
	return "?????"
}

func expandChanges(pair sqlschema.Pair[sqlschema.ColumnWalker], changes sqlschema.ColumnChanges) sqlschema.AlterColumnChanges {
	var acc sqlschema.AlterColumnChanges

	if changes.DefaultChanged() {
		if pair.Next.Default() != nil {
			acc.SetDefault = pair.Next.Default()
		} else {
			acc.DropDefault = true
		}
	}

	if changes.ArityChanged() {
		p, n := pair.Prev.Arity(), pair.Next.Arity()
		switch {
		case p == sqlschema.Required && n == sqlschema.Nullable:
			acc.DropNotNull = true
		case p == sqlschema.Nullable && n == sqlschema.Required:
			acc.SetNotNull = true
		case p == sqlschema.List && n == sqlschema.Nullable:
			acc.SetType = true
			acc.DropNotNull = true
		case p == sqlschema.List && n == sqlschema.Required:
			acc.SetType = true
			acc.SetNotNull = true
		case (p == sqlschema.Nullable || p == sqlschema.Required) && n == sqlschema.List:
			acc.SetType = true
		}
	}

	if changes.TypeChanged() {
		acc.SetType = true
	}

	if changes.AutoIncrementChanged() {
		if pair.Prev.Get().AutoIncrement {
			acc.DropDefault = true
		} else {
			acc.AddSequence = true
		}
	}
	return acc
}

// func renderCreateExtension(pair sqlschema.Pair[sqlschema.Database], step CreateExtension) (sqlschema.Statement, error) {
// 	extension := pair.Next.WalkExtension(step.Extension)
// 	stmt := sqlschema.Statement{Comment: fmt.Sprintf("Create extension %s.", quoteIdent(extension.Name()))}
// 	stmt.Steps.Add(sqlschema.Step{Cmd: fmt.Sprintf("CREATE EXTENSION %s", quoteIdent(extension.Name()))})
// 	return stmt, nil
// }

// func renderAlterExtension(pair sqlschema.Pair[sqlschema.Database], step AlterExtension) (sqlschema.Statement, error) {
// 	extension := pair.Next.WalkExtension(step.Extension.Next)
// 	stmt := sqlschema.Statement{Comment: fmt.Sprintf("Alter extension %s.", quoteIdent(extension.Name()))}
// 	stmt.Steps.Add(sqlschema.Step{Cmd: fmt.Sprintf("ALTER EXTENSION %s", quoteIdent(extension.Name()))})
// 	return stmt, nil
// }

// func renderDropExtension(pair sqlschema.Pair[sqlschema.Database], step DropExtension) (sqlschema.Statement, error) {
// 	extension := pair.Prev.WalkExtension(step.Extension)
// 	stmt := sqlschema.Statement{Comment: fmt.Sprintf("Drop extension %s.", quoteIdent(extension.Name()))}
// 	stmt.Steps.Add(sqlschema.Step{Cmd: fmt.Sprintf("DROP EXTENSION %s", quoteIdent(extension.Name()))})
// 	return stmt, nil
// }

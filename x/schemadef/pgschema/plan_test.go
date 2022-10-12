package pgschema

import (
	"strconv"
	"testing"

	"kwil/x/schemadef/sqlschema"
	"kwil/x/sql/sqlutil"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestPlanChanges(t *testing.T) {
	tests := []struct {
		changes  []sqlschema.SchemaChange
		options  []sqlschema.PlanOption
		mock     func(mock)
		wantPlan *sqlschema.Plan
		wantErr  bool
	}{
		{
			changes: []sqlschema.SchemaChange{
				&sqlschema.AddSchema{S: sqlschema.New("test"), Extra: []sqlschema.SchemaClause{&sqlschema.IfNotExists{}}},
				&sqlschema.DropSchema{S: sqlschema.New("test"), Extra: []sqlschema.SchemaClause{&sqlschema.IfExists{}}},
				&sqlschema.DropSchema{S: sqlschema.New("test"), Extra: []sqlschema.SchemaClause{}},
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    false,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{
						Cmd:     `CREATE SCHEMA IF NOT EXISTS "test"`,
						Reverse: `DROP SCHEMA "test" CASCADE`,
					},
					{
						Cmd: `DROP SCHEMA IF EXISTS "test" CASCADE`,
					},
					{
						Cmd: `DROP SCHEMA "test" CASCADE`,
					},
				},
			},
		},
		{
			changes: []sqlschema.SchemaChange{
				func() sqlschema.SchemaChange {
					users := &sqlschema.Table{
						Name: "users",
						Columns: []*sqlschema.Column{
							{Name: "id", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "bigint"}}},
						},
					}
					pets := &sqlschema.Table{
						Name: "pets",
						Columns: []*sqlschema.Column{
							{Name: "id", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "bigint"}}},
							{Name: "user_id",
								Type: &sqlschema.ColumnType{
									Type: &sqlschema.IntegerType{T: "bigint"},
								},
							},
						},
					}
					fk := &sqlschema.ForeignKey{
						Name:       "pets_user_id_fkey",
						Table:      pets,
						OnUpdate:   sqlschema.NoAction,
						OnDelete:   sqlschema.Cascade,
						RefTable:   users,
						Columns:    []*sqlschema.Column{pets.Columns[1]},
						RefColumns: []*sqlschema.Column{users.Columns[0]},
					}
					pets.ForeignKeys = []*sqlschema.ForeignKey{fk}
					return &sqlschema.ModifyTable{
						T: pets,
						Changes: []sqlschema.SchemaChange{
							&sqlschema.DropForeignKey{
								F: fk,
							},
						},
					}
				}(),
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    true,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{
						Cmd:     `ALTER TABLE "pets" DROP CONSTRAINT "pets_user_id_fkey"`,
						Reverse: `ALTER TABLE "pets" ADD CONSTRAINT "pets_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE`,
					},
				},
			},
		},
		{
			changes: []sqlschema.SchemaChange{
				func() sqlschema.SchemaChange {
					users := &sqlschema.Table{
						Name: "users",
						Columns: []*sqlschema.Column{
							{Name: "id", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "bigint"}}},
							{Name: "name", Type: &sqlschema.ColumnType{Type: &sqlschema.StringType{T: "varchar(255)"}}}},
					}
					return &sqlschema.ModifyTable{
						T: users,
						Changes: []sqlschema.SchemaChange{
							&sqlschema.DropIndex{
								I: sqlschema.NewIndex("name_index").
									AddParts(sqlschema.NewColumnPart(sqlschema.NewColumn("name"))),
							},
						},
					}
				}(),
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    true,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{
						Cmd:     `DROP INDEX "name_index"`,
						Reverse: `CREATE INDEX "name_index" ON "users" ("name")`,
					},
				},
			},
		},
		{
			changes: []sqlschema.SchemaChange{
				func() sqlschema.SchemaChange {
					users := &sqlschema.Table{
						Name: "users",
						Columns: []*sqlschema.Column{
							{Name: "id", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "bigint"}}},
							{Name: "nickname", Type: &sqlschema.ColumnType{Type: &sqlschema.StringType{T: "varchar(255)"}}}},
					}
					return &sqlschema.ModifyTable{
						T: users,
						Changes: []sqlschema.SchemaChange{
							&sqlschema.DropIndex{
								I: sqlschema.NewUniqueIndex("unique_nickname").
									AddColumns(sqlschema.NewColumn("nickname")).
									AddAttrs(&ConstraintType{T: "u"}),
							},
						},
					}
				}(),
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    true,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{
						Cmd:     `ALTER TABLE "users" DROP CONSTRAINT "unique_nickname"`,
						Reverse: `ALTER TABLE "users" ADD CONSTRAINT "unique_nickname" UNIQUE ("nickname")`,
					},
				},
			},
		},
		{
			changes: []sqlschema.SchemaChange{
				&sqlschema.AddSchema{S: &sqlschema.Schema{Name: "test"}},
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    true,
				Transactional: true,
				Changes:       []*sqlschema.Change{{Cmd: `CREATE SCHEMA "test"`, Reverse: `DROP SCHEMA "test" CASCADE`}}},
		},
		{
			changes: []sqlschema.SchemaChange{
				&sqlschema.DropSchema{S: &sqlschema.Schema{Name: "atlas"}},
			},
			wantPlan: &sqlschema.Plan{
				Transactional: true,
				Changes:       []*sqlschema.Change{{Cmd: `DROP SCHEMA "atlas" CASCADE`}},
			},
		},
		{
			changes: []sqlschema.SchemaChange{
				&sqlschema.AddTable{
					T: &sqlschema.Table{
						Name: "posts",
						Columns: []*sqlschema.Column{
							{Name: "id", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "integer"}}, Attrs: []sqlschema.Attr{&Identity{}, &sqlschema.Comment{}}},
							{Name: "text", Type: &sqlschema.ColumnType{Type: &sqlschema.StringType{T: "text"}, Nullable: true}},
							{Name: "directions", Type: &sqlschema.ColumnType{Type: &ArrayType{T: "direction[]", Type: &sqlschema.EnumType{T: "direction", Values: []string{"NORTH", "SOUTH"}, Schema: sqlschema.New("public")}}}},
							{Name: "states", Type: &sqlschema.ColumnType{Type: &ArrayType{T: "state[]", Type: &sqlschema.EnumType{T: "state", Values: []string{"ON", "OFF"}}}}},
						},
						Attrs: []sqlschema.Attr{
							&sqlschema.Comment{},
							&sqlschema.Check{Name: "id_nonzero", Expr: `("id" > 0)`},
							&sqlschema.Check{Name: "text_len", Expr: `(length("text") > 0)`, Attrs: []sqlschema.Attr{&NoInherit{}}},
							&sqlschema.Check{Name: "a_in_b", Expr: `(a) in (b)`},
							&Partition{T: "HASH", Parts: []*PartitionPart{{Column: "text"}}},
						},
					},
				},
			},
			mock: func(m mock) {
				m.ExpectQuery(sqlutil.Escape("SELECT * FROM pg_type t JOIN pg_namespace n on t.typnamespace = n.oid WHERE t.typname = $1 AND t.typtype = 'e' AND n.nspname = $2")).
					WithArgs("direction", "public").
					WillReturnRows(sqlmock.NewRows([]string{"name"}))
				m.ExpectQuery(sqlutil.Escape("SELECT * FROM pg_type t JOIN pg_namespace n on t.typnamespace = n.oid WHERE t.typname = $1 AND t.typtype = 'e'")).
					WithArgs("state").
					WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("state"))
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    true,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{Cmd: `CREATE TYPE "public"."direction" AS ENUM ('NORTH', 'SOUTH')`, Reverse: `DROP TYPE "public"."direction"`},
					{Cmd: `CREATE TYPE "state" AS ENUM ('ON', 'OFF')`, Reverse: `DROP TYPE "state"`},
					{Cmd: `CREATE TABLE "posts" ("id" integer NOT NULL GENERATED BY DEFAULT AS IDENTITY, "text" text NULL, "directions" "public"."direction"[] NOT NULL, "states" "state"[] NOT NULL, CONSTRAINT "id_nonzero" CHECK ("id" > 0), CONSTRAINT "text_len" CHECK (length("text") > 0) NO INHERIT, CONSTRAINT "a_in_b" CHECK ((a) in (b))) PARTITION BY HASH ("text")`, Reverse: `DROP TABLE "posts"`},
				},
			},
		},
		{
			changes: []sqlschema.SchemaChange{
				&sqlschema.AddTable{
					T: &sqlschema.Table{
						Name: "posts",
						Columns: []*sqlschema.Column{
							{Name: "id", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "integer"}}, Attrs: []sqlschema.Attr{&Identity{Sequence: &Sequence{Start: 1024}}}},
						},
					},
				},
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    true,
				Transactional: true,
				Changes:       []*sqlschema.Change{{Cmd: `CREATE TABLE "posts" ("id" integer NOT NULL GENERATED BY DEFAULT AS IDENTITY (START WITH 1024))`, Reverse: `DROP TABLE "posts"`}},
			},
		},
		{
			changes: []sqlschema.SchemaChange{
				&sqlschema.AddTable{
					T: &sqlschema.Table{
						Name: "posts",
						Columns: []*sqlschema.Column{
							{Name: "id", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "integer"}}},
							{Name: "nid", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "integer"}}, Attrs: []sqlschema.Attr{&sqlschema.GeneratedExpr{Expr: "id+1"}}},
						},
					},
				},
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    true,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{
						Cmd:     `CREATE TABLE "posts" ("id" integer NOT NULL, "nid" integer NOT NULL GENERATED ALWAYS AS (id+1) STORED)`,
						Reverse: `DROP TABLE "posts"`,
					},
				},
			},
		},
		{
			changes: []sqlschema.SchemaChange{
				&sqlschema.ModifyTable{
					T: sqlschema.NewTable("posts").
						AddColumns(
							sqlschema.NewIntColumn("c1", "int").
								SetGeneratedExpr(&sqlschema.GeneratedExpr{Expr: "id+1"}),
						),
					Changes: []sqlschema.SchemaChange{
						&sqlschema.ModifyColumn{
							Change: sqlschema.ChangeGenerated,
							From: sqlschema.NewIntColumn("c1", "int").
								SetGeneratedExpr(&sqlschema.GeneratedExpr{Expr: "id+1"}),
							To: sqlschema.NewIntColumn("c1", "int"),
						},
					},
				},
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    false,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{
						Cmd: `ALTER TABLE "posts" ALTER COLUMN "c1" DROP EXPRESSION`,
					},
				},
			},
		},
		{
			changes: []sqlschema.SchemaChange{
				&sqlschema.AddTable{
					T: &sqlschema.Table{
						Name: "posts",
						Columns: []*sqlschema.Column{
							{Name: "id", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "integer"}}, Attrs: []sqlschema.Attr{&Identity{Sequence: &Sequence{Increment: 2}}}},
						},
					},
				},
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    true,
				Transactional: true,
				Changes:       []*sqlschema.Change{{Cmd: `CREATE TABLE "posts" ("id" integer NOT NULL GENERATED BY DEFAULT AS IDENTITY (INCREMENT BY 2))`, Reverse: `DROP TABLE "posts"`}},
			},
		},
		{
			changes: []sqlschema.SchemaChange{
				&sqlschema.AddTable{
					T: &sqlschema.Table{
						Name: "posts",
						Columns: []*sqlschema.Column{
							{Name: "id", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "integer"}}, Attrs: []sqlschema.Attr{&Identity{Sequence: &Sequence{Start: 100, Increment: 2}}}},
						},
					},
				},
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    true,
				Transactional: true,
				Changes:       []*sqlschema.Change{{Cmd: `CREATE TABLE "posts" ("id" integer NOT NULL GENERATED BY DEFAULT AS IDENTITY (START WITH 100 INCREMENT BY 2))`, Reverse: `DROP TABLE "posts"`}},
			},
		},
		{
			changes: []sqlschema.SchemaChange{
				&sqlschema.DropTable{T: &sqlschema.Table{Name: "posts"}},
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    false,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{Cmd: `DROP TABLE "posts"`},
				},
			},
		},
		{
			changes: []sqlschema.SchemaChange{
				func() sqlschema.SchemaChange {
					users := &sqlschema.Table{
						Name: "users",
						Columns: []*sqlschema.Column{
							{Name: "id", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "bigint"}}},
						},
					}
					return &sqlschema.ModifyTable{
						T: users,
						Changes: []sqlschema.SchemaChange{
							&sqlschema.AddColumn{
								C: &sqlschema.Column{Name: "name", Type: &sqlschema.ColumnType{Type: &sqlschema.StringType{T: "varchar", Size: 255}}, Attrs: []sqlschema.Attr{&sqlschema.Comment{Text: "foo"}}, Default: &sqlschema.Literal{V: "'logged_in'"}},
							},
							&sqlschema.AddColumn{
								C: &sqlschema.Column{Name: "last", Type: &sqlschema.ColumnType{Type: &sqlschema.StringType{T: "varchar", Size: 255}}, Attrs: []sqlschema.Attr{&sqlschema.Comment{Text: "bar"}}, Default: &sqlschema.RawExpr{X: "'logged_in'"}},
							},
							&sqlschema.AddIndex{
								I: &sqlschema.Index{
									Name: "id_key",
									Parts: []*sqlschema.IndexPart{
										{Column: users.Columns[0], Descending: true},
									},
									Attrs: []sqlschema.Attr{
										&sqlschema.Comment{Text: "comment"},
										&IndexPredicate{Predicate: "success"},
									},
								},
							},
							&sqlschema.AddIndex{
								I: &sqlschema.Index{
									Name: "id_brin",
									Parts: []*sqlschema.IndexPart{
										{Column: users.Columns[0], Descending: true},
									},
									Attrs: []sqlschema.Attr{
										&IndexType{T: IndexTypeBRIN},
										&IndexStorageParams{PagesPerRange: 2},
									},
								},
							},
							&sqlschema.AddCheck{
								C: &sqlschema.Check{Name: "name_not_empty", Expr: `("name" <> '')`},
							},
							&sqlschema.DropCheck{
								C: &sqlschema.Check{Name: "id_nonzero", Expr: `("id" <> 0)`},
							},
							&sqlschema.ModifyCheck{
								From: &sqlschema.Check{Name: "id_iseven", Expr: `("id" % 2 = 0)`},
								To:   &sqlschema.Check{Name: "id_iseven", Expr: `(("id") % 2 = 0)`},
							},
							&sqlschema.AddIndex{
								I: &sqlschema.Index{
									Name: "include_key",
									Parts: []*sqlschema.IndexPart{
										{Column: users.Columns[0]},
									},
									Attrs: []sqlschema.Attr{
										&IndexInclude{Columns: []string{"a", "b"}},
									},
								},
							},
						},
					}
				}(),
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    true,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{
						Cmd:     `ALTER TABLE "users" ADD COLUMN "name" character varying(255) NOT NULL DEFAULT 'logged_in', ADD COLUMN "last" character varying(255) NOT NULL DEFAULT 'logged_in', ADD CONSTRAINT "name_not_empty" CHECK ("name" <> ''), DROP CONSTRAINT "id_nonzero", DROP CONSTRAINT "id_iseven", ADD CONSTRAINT "id_iseven" CHECK (("id") % 2 = 0)`,
						Reverse: `ALTER TABLE "users" DROP CONSTRAINT "id_iseven", ADD CONSTRAINT "id_iseven" CHECK ("id" % 2 = 0), ADD CONSTRAINT "id_nonzero" CHECK ("id" <> 0), DROP CONSTRAINT "name_not_empty", DROP COLUMN "last", DROP COLUMN "name"`,
					},
					{
						Cmd:     `CREATE INDEX "id_key" ON "users" ("id" DESC) WHERE success`,
						Reverse: `DROP INDEX "id_key"`,
					},
					{
						Cmd:     `CREATE INDEX "id_brin" ON "users" USING BRIN ("id" DESC) WITH (pages_per_range = 2)`,
						Reverse: `DROP INDEX "id_brin"`,
					},
					{
						Cmd:     `CREATE INDEX "include_key" ON "users" ("id") INCLUDE ("a", "b")`,
						Reverse: `DROP INDEX "include_key"`,
					},
					{
						Cmd:     `COMMENT ON COLUMN "users" ."name" IS 'foo'`,
						Reverse: `COMMENT ON COLUMN "users" ."name" IS ''`,
					},
					{
						Cmd:     `COMMENT ON COLUMN "users" ."last" IS 'bar'`,
						Reverse: `COMMENT ON COLUMN "users" ."last" IS ''`,
					},
					{
						Cmd:     `COMMENT ON INDEX "id_key" IS 'comment'`,
						Reverse: `COMMENT ON INDEX "id_key" IS ''`,
					},
				},
			},
		},
		{
			changes: []sqlschema.SchemaChange{
				func() sqlschema.SchemaChange {
					users := &sqlschema.Table{
						Name: "users",
						Columns: []*sqlschema.Column{
							{Name: "id", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "bigint"}}},
						},
					}
					return &sqlschema.ModifyTable{
						T: users,
						Changes: []sqlschema.SchemaChange{
							&sqlschema.DropColumn{
								C: &sqlschema.Column{Name: "name", Type: &sqlschema.ColumnType{Type: &sqlschema.StringType{T: "varchar"}}},
							},
							&sqlschema.ModifyColumn{
								From:   &sqlschema.Column{Name: "id", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "bigint"}}, Attrs: []sqlschema.Attr{&Identity{}, &sqlschema.Comment{Text: "comment"}}},
								To:     &sqlschema.Column{Name: "id", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "bigint"}}, Attrs: []sqlschema.Attr{&Identity{Sequence: &Sequence{Start: 1024}}}},
								Change: sqlschema.ChangeAttr | sqlschema.ChangeComment,
							},
						},
					}
				}(),
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    true,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{
						Cmd:     `ALTER TABLE "users" DROP COLUMN "name", ALTER COLUMN "id" SET GENERATED BY DEFAULT SET START WITH 1024 SET INCREMENT BY 1 RESTART`,
						Reverse: `ALTER TABLE "users" ALTER COLUMN "id" SET GENERATED BY DEFAULT SET START WITH 1 SET INCREMENT BY 1 RESTART, ADD COLUMN "name" character varying NOT NULL`,
					},
					{
						Cmd:     `COMMENT ON COLUMN "users" ."id" IS ''`,
						Reverse: `COMMENT ON COLUMN "users" ."id" IS 'comment'`,
					},
				},
			},
		},
		{
			changes: []sqlschema.SchemaChange{
				func() sqlschema.SchemaChange {
					users := &sqlschema.Table{
						Name: "users",
						Columns: []*sqlschema.Column{
							{Name: "id", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "bigint"}}},
						},
					}
					return &sqlschema.ModifyTable{
						T: users,
						Changes: []sqlschema.SchemaChange{
							&sqlschema.DropColumn{
								C: &sqlschema.Column{Name: "name", Type: &sqlschema.ColumnType{Type: &sqlschema.StringType{T: "varchar"}}},
							},
							&sqlschema.ModifyColumn{
								From:   &sqlschema.Column{Name: "id", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "bigint"}}, Attrs: []sqlschema.Attr{&Identity{Sequence: &Sequence{Start: 0, Last: 1025}}}},
								To:     &sqlschema.Column{Name: "id", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "bigint"}}, Attrs: []sqlschema.Attr{&Identity{Sequence: &Sequence{Start: 1024}}}},
								Change: sqlschema.ChangeAttr,
							},
						},
					}
				}(),
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    true,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{
						Cmd:     `ALTER TABLE "users" DROP COLUMN "name", ALTER COLUMN "id" SET GENERATED BY DEFAULT SET START WITH 1024 SET INCREMENT BY 1`,
						Reverse: `ALTER TABLE "users" ALTER COLUMN "id" SET GENERATED BY DEFAULT SET START WITH 1 SET INCREMENT BY 1, ADD COLUMN "name" character varying NOT NULL`,
					},
				},
			},
		},
		{
			changes: []sqlschema.SchemaChange{
				&sqlschema.ModifyTable{
					T: &sqlschema.Table{Name: "users", Schema: &sqlschema.Schema{Name: "public"}},
					Changes: []sqlschema.SchemaChange{
						&sqlschema.AddAttr{
							A: &sqlschema.Comment{Text: "foo"},
						},
					},
				},
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    true,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{Cmd: `COMMENT ON TABLE "public"."users" IS 'foo'`, Reverse: `COMMENT ON TABLE "public"."users" IS ''`},
				},
			},
		},
		{
			changes: []sqlschema.SchemaChange{
				&sqlschema.ModifyTable{
					T: &sqlschema.Table{Name: "users", Schema: &sqlschema.Schema{Name: "public"}},
					Changes: []sqlschema.SchemaChange{
						&sqlschema.ModifyAttr{
							To:   &sqlschema.Comment{Text: "foo"},
							From: &sqlschema.Comment{Text: "bar"},
						},
					},
				},
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    true,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{Cmd: `COMMENT ON TABLE "public"."users" IS 'foo'`, Reverse: `COMMENT ON TABLE "public"."users" IS 'bar'`},
				},
			},
		},
		{
			changes: []sqlschema.SchemaChange{
				func() sqlschema.SchemaChange {
					users := sqlschema.NewTable("users").
						SetSchema(sqlschema.New("public")).
						AddColumns(
							sqlschema.NewEnumColumn("state", sqlschema.EnumName("state"), sqlschema.EnumValues("on", "off")),
							sqlschema.NewEnumColumn("status", sqlschema.EnumName("status"), sqlschema.EnumValues("a", "b"), sqlschema.EnumSchema(sqlschema.New("test"))),
						)
					return &sqlschema.ModifyTable{
						T: users,
						Changes: []sqlschema.SchemaChange{
							&sqlschema.ModifyColumn{
								From:   &sqlschema.Column{Name: "state", Type: &sqlschema.ColumnType{Type: &sqlschema.StringType{T: "text"}}},
								To:     users.Columns[0],
								Change: sqlschema.ChangeType,
							},
							&sqlschema.ModifyColumn{
								From:   &sqlschema.Column{Name: "status", Type: &sqlschema.ColumnType{Type: &sqlschema.StringType{T: "text"}}},
								To:     users.Columns[1],
								Change: sqlschema.ChangeType,
							},
							&sqlschema.DropColumn{
								C: sqlschema.NewEnumColumn("dc1", sqlschema.EnumName("de"), sqlschema.EnumValues("on")),
							},
							&sqlschema.DropColumn{
								C: sqlschema.NewEnumColumn("dc2", sqlschema.EnumName("de"), sqlschema.EnumValues("on")),
							},
						},
					}
				}(),
			},
			mock: func(m mock) {
				m.ExpectQuery(sqlutil.Escape("SELECT * FROM pg_type t JOIN pg_namespace n on t.typnamespace = n.oid WHERE t.typname = $1 AND t.typtype = 'e' AND n.nspname = $2 ")).
					WithArgs("state", "public").
					WillReturnRows(sqlmock.NewRows([]string{"name"}))
				m.ExpectQuery(sqlutil.Escape("SELECT * FROM pg_type t JOIN pg_namespace n on t.typnamespace = n.oid WHERE t.typname = $1 AND t.typtype = 'e' AND n.nspname = $2 ")).
					WithArgs("status", "test").
					WillReturnRows(sqlmock.NewRows([]string{"name"}))
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    true,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{Cmd: `CREATE TYPE "public"."state" AS ENUM ('on', 'off')`, Reverse: `DROP TYPE "public"."state"`},
					{Cmd: `CREATE TYPE "test"."status" AS ENUM ('a', 'b')`, Reverse: `DROP TYPE "test"."status"`},
					{Cmd: `ALTER TABLE "public"."users" ALTER COLUMN "state" TYPE "public"."state", ALTER COLUMN "status" TYPE "test"."status", DROP COLUMN "dc1", DROP COLUMN "dc2"`, Reverse: `ALTER TABLE "public"."users" ADD COLUMN "dc2" "public"."de" NOT NULL, ADD COLUMN "dc1" "public"."de" NOT NULL, ALTER COLUMN "status" TYPE text, ALTER COLUMN "state" TYPE text`},
					{Cmd: `DROP TYPE "public"."de"`, Reverse: `CREATE TYPE "public"."de" AS ENUM ('on')`},
				},
			},
		},
		{
			changes: []sqlschema.SchemaChange{
				func() sqlschema.SchemaChange {
					users := sqlschema.NewTable("users").
						SetSchema(sqlschema.New("public")).
						AddColumns(
							sqlschema.NewEnumColumn("state", sqlschema.EnumName("state"), sqlschema.EnumValues("on", "off")),
						)
					return &sqlschema.ModifyTable{
						T: users,
						Changes: []sqlschema.SchemaChange{
							&sqlschema.ModifyColumn{
								From:   users.Columns[0],
								To:     &sqlschema.Column{Name: "state", Type: &sqlschema.ColumnType{Type: &sqlschema.EnumType{T: "state", Values: []string{"on", "off", "unknown"}}}},
								Change: sqlschema.ChangeType,
							},
						},
					}
				}(),
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    false,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{Cmd: `ALTER TYPE "public"."state" ADD VALUE 'unknown'`},
				},
			},
		},
		// Modify column type and drop comment.
		{
			changes: []sqlschema.SchemaChange{
				func() sqlschema.SchemaChange {
					users := sqlschema.NewTable("users").
						SetSchema(sqlschema.New("public")).
						AddColumns(
							sqlschema.NewEnumColumn("state", sqlschema.EnumName("state"), sqlschema.EnumValues("on", "off")),
						)
					return &sqlschema.ModifyTable{
						T: users,
						Changes: []sqlschema.SchemaChange{
							&sqlschema.ModifyColumn{
								From:   &sqlschema.Column{Name: "state", Type: &sqlschema.ColumnType{Type: &sqlschema.EnumType{T: "state", Values: []string{"on", "off"}}}, Attrs: []sqlschema.Attr{&sqlschema.Comment{Text: "foo"}}},
								To:     &sqlschema.Column{Name: "state", Type: &sqlschema.ColumnType{Type: &sqlschema.EnumType{T: "state", Values: []string{"on", "off", "unknown"}}}},
								Change: sqlschema.ChangeType | sqlschema.ChangeComment,
							},
						},
					}
				}(),
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    false,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{Cmd: `ALTER TYPE "public"."state" ADD VALUE 'unknown'`},
					{Cmd: `COMMENT ON COLUMN "public"."users" ."state" IS ''`, Reverse: `COMMENT ON COLUMN "public"."users" ."state" IS 'foo'`},
				},
			},
		},
		// Modify column type and add comment.
		{
			changes: []sqlschema.SchemaChange{
				func() sqlschema.SchemaChange {
					users := &sqlschema.Table{
						Name: "users",
						Columns: []*sqlschema.Column{
							{Name: "state", Type: &sqlschema.ColumnType{Type: &sqlschema.EnumType{T: "state", Values: []string{"on", "off"}}}},
						},
					}
					return &sqlschema.ModifyTable{
						T: users,
						Changes: []sqlschema.SchemaChange{
							&sqlschema.ModifyColumn{
								From:   &sqlschema.Column{Name: "state", Type: &sqlschema.ColumnType{Type: &sqlschema.EnumType{T: "state", Values: []string{"on", "off"}}}},
								To:     &sqlschema.Column{Name: "state", Type: &sqlschema.ColumnType{Type: &sqlschema.EnumType{T: "state", Values: []string{"on", "off", "unknown"}}}, Attrs: []sqlschema.Attr{&sqlschema.Comment{Text: "foo"}}},
								Change: sqlschema.ChangeType | sqlschema.ChangeComment,
							},
						},
					}
				}(),
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    false,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{Cmd: `ALTER TYPE "state" ADD VALUE 'unknown'`},
					{Cmd: `COMMENT ON COLUMN "users" ."state" IS 'foo'`, Reverse: `COMMENT ON COLUMN "users" ."state" IS ''`},
				},
			},
		},
		// Modify column comment.
		{
			changes: []sqlschema.SchemaChange{
				func() sqlschema.SchemaChange {
					users := &sqlschema.Table{
						Name: "users",
						Columns: []*sqlschema.Column{
							{Name: "state", Type: &sqlschema.ColumnType{Type: &sqlschema.StringType{T: "text"}}},
						},
					}
					return &sqlschema.ModifyTable{
						T: users,
						Changes: []sqlschema.SchemaChange{
							&sqlschema.ModifyColumn{
								From:   &sqlschema.Column{Name: "state", Type: &sqlschema.ColumnType{Type: &sqlschema.StringType{T: "text"}}, Attrs: []sqlschema.Attr{&sqlschema.Comment{Text: "bar"}}},
								To:     &sqlschema.Column{Name: "state", Type: &sqlschema.ColumnType{Type: &sqlschema.StringType{T: "text"}}, Attrs: []sqlschema.Attr{&sqlschema.Comment{Text: "foo"}}},
								Change: sqlschema.ChangeComment,
							},
						},
					}
				}(),
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    true,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{Cmd: `COMMENT ON COLUMN "users" ."state" IS 'foo'`, Reverse: `COMMENT ON COLUMN "users" ."state" IS 'bar'`},
				},
			},
		},
		// Modify index comment.
		{
			changes: []sqlschema.SchemaChange{
				func() sqlschema.SchemaChange {
					users := &sqlschema.Table{
						Name: "users",
						Columns: []*sqlschema.Column{
							{Name: "id", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "bigint"}}},
						},
					}
					return &sqlschema.ModifyTable{
						T: users,
						Changes: []sqlschema.SchemaChange{
							&sqlschema.ModifyIndex{
								From: sqlschema.NewIndex("id_key").
									AddColumns(users.Columns[0]).
									SetComment("foo"),
								To: sqlschema.NewIndex("id_key").
									AddColumns(users.Columns[0]).
									SetComment("bar"),
								Change: sqlschema.ChangeComment,
							},
						},
					}
				}(),
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    true,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{Cmd: `COMMENT ON INDEX "id_key" IS 'bar'`, Reverse: `COMMENT ON INDEX "id_key" IS 'foo'`},
				},
			},
		},
		// Modify default values.
		{
			changes: []sqlschema.SchemaChange{
				func() sqlschema.SchemaChange {
					users := &sqlschema.Table{
						Name: "users",
						Columns: []*sqlschema.Column{
							{Name: "id", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "bigint"}}},
							{Name: "one", Default: &sqlschema.Literal{V: "'one'"}, Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "bigint"}}},
							{Name: "two", Default: &sqlschema.Literal{V: "'two'"}, Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "bigint"}}},
						},
					}
					return &sqlschema.ModifyTable{
						T: users,
						Changes: []sqlschema.SchemaChange{
							&sqlschema.ModifyColumn{
								From:   &sqlschema.Column{Name: "one", Default: &sqlschema.Literal{V: "'one'"}, Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "bigint"}}},
								To:     &sqlschema.Column{Name: "one", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "bigint"}}},
								Change: sqlschema.ChangeDefault,
							},
							&sqlschema.ModifyColumn{
								From:   &sqlschema.Column{Name: "two", Default: &sqlschema.Literal{V: "'two'"}, Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "bigint"}}},
								To:     &sqlschema.Column{Name: "two", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "bigint"}}},
								Change: sqlschema.ChangeDefault,
							},
							&sqlschema.ModifyColumn{
								From:   &sqlschema.Column{Name: "id", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "bigint"}}, Attrs: []sqlschema.Attr{&Identity{}}},
								To:     &sqlschema.Column{Name: "id", Type: &sqlschema.ColumnType{Type: &sqlschema.IntegerType{T: "bigint"}}, Attrs: []sqlschema.Attr{&Identity{Sequence: &Sequence{Start: 1024}}}},
								Change: sqlschema.ChangeAttr,
							},
						},
					}
				}(),
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    true,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{
						Cmd:     `ALTER TABLE "users" ALTER COLUMN "one" DROP DEFAULT, ALTER COLUMN "two" DROP DEFAULT, ALTER COLUMN "id" SET GENERATED BY DEFAULT SET START WITH 1024 SET INCREMENT BY 1 RESTART`,
						Reverse: `ALTER TABLE "users" ALTER COLUMN "id" SET GENERATED BY DEFAULT SET START WITH 1 SET INCREMENT BY 1 RESTART, ALTER COLUMN "two" SET DEFAULT 'two', ALTER COLUMN "one" SET DEFAULT 'one'`,
					},
				},
			},
		},
		{
			changes: []sqlschema.SchemaChange{
				&sqlschema.RenameTable{
					From: sqlschema.NewTable("t1").SetSchema(sqlschema.New("s1")),
					To:   sqlschema.NewTable("t2").SetSchema(sqlschema.New("s2")),
				},
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    true,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{
						Cmd:     `ALTER TABLE "s1"."t1" RENAME TO "s2"."t2"`,
						Reverse: `ALTER TABLE "s2"."t2" RENAME TO "s1"."t1"`,
					},
				},
			},
		},
		{
			changes: []sqlschema.SchemaChange{
				&sqlschema.ModifyTable{
					T: sqlschema.NewTable("t1").SetSchema(sqlschema.New("s1")),
					Changes: []sqlschema.SchemaChange{
						&sqlschema.RenameColumn{
							From: sqlschema.NewColumn("a"),
							To:   sqlschema.NewColumn("b"),
						},
					},
				},
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    true,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{
						Cmd:     `ALTER TABLE "s1"."t1" RENAME COLUMN "a" TO "b"`,
						Reverse: `ALTER TABLE "s1"."t1" RENAME COLUMN "b" TO "a"`,
					},
				},
			},
		},
		{
			changes: func() []sqlschema.SchemaChange {
				s := sqlschema.New("s1")
				t := sqlschema.NewTable("t1").SetSchema(s)

				change := &sqlschema.ModifyTable{
					T: t,
					Changes: []sqlschema.SchemaChange{
						&sqlschema.RenameColumn{
							From: sqlschema.NewColumn("a"),
							To:   sqlschema.NewColumn("b"),
						},
						&sqlschema.AddColumn{
							C: sqlschema.NewIntColumn("c", "int"),
						},
					},
				}
				return []sqlschema.SchemaChange{change}
			}(),
			wantPlan: &sqlschema.Plan{
				Reversible:    true,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{
						Cmd:     `ALTER TABLE "s1"."t1" ADD COLUMN "c" integer NOT NULL`,
						Reverse: `ALTER TABLE "s1"."t1" DROP COLUMN "c"`,
					},
					{
						Cmd:     `ALTER TABLE "s1"."t1" RENAME COLUMN "a" TO "b"`,
						Reverse: `ALTER TABLE "s1"."t1" RENAME COLUMN "b" TO "a"`,
					},
				},
			},
		},
		{
			changes: []sqlschema.SchemaChange{
				&sqlschema.ModifyTable{
					T: sqlschema.NewTable("t1").SetSchema(sqlschema.New("s1")),
					Changes: []sqlschema.SchemaChange{
						&sqlschema.RenameIndex{
							From: sqlschema.NewIndex("a"),
							To:   sqlschema.NewIndex("b"),
						},
					},
				},
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    true,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{
						Cmd:     `ALTER INDEX "a" RENAME TO "b"`,
						Reverse: `ALTER INDEX "b" RENAME TO "a"`,
					},
				},
			},
		},
		// Invalid serial type.
		{
			changes: []sqlschema.SchemaChange{
				&sqlschema.AddTable{
					T: &sqlschema.Table{
						Name: "posts",
						Columns: []*sqlschema.Column{
							{Name: "id", Type: &sqlschema.ColumnType{Type: &SerialType{T: "serial"}, Nullable: true}},
						},
					},
				},
			},
			wantErr: true,
		},
		// Drop serial sequence.
		{
			changes: []sqlschema.SchemaChange{
				&sqlschema.ModifyTable{
					T: sqlschema.NewTable("posts").
						SetSchema(sqlschema.New("public")).
						AddColumns(
							sqlschema.NewIntColumn("c1", "integer"),
							sqlschema.NewIntColumn("c2", "integer"),
						),
					Changes: sqlschema.Changes{
						&sqlschema.ModifyColumn{
							From:   sqlschema.NewColumn("c1").SetType(&SerialType{T: "smallserial"}),
							To:     sqlschema.NewIntColumn("c1", "integer"),
							Change: sqlschema.ChangeType,
						},
						&sqlschema.ModifyColumn{
							From:   sqlschema.NewColumn("c2").SetType(&SerialType{T: "serial", SequenceName: "previous_name"}),
							To:     sqlschema.NewIntColumn("c2", "integer"),
							Change: sqlschema.ChangeType,
						},
					},
				},
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    true,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{
						Cmd:     `ALTER TABLE "public"."posts" ALTER COLUMN "c1" DROP DEFAULT, ALTER COLUMN "c1" TYPE integer, ALTER COLUMN "c2" DROP DEFAULT`,
						Reverse: `ALTER TABLE "public"."posts" ALTER COLUMN "c2" SET DEFAULT nextval('"public"."previous_name"'), ALTER COLUMN "c1" SET DEFAULT nextval('"public"."posts_c1_seq"'), ALTER COLUMN "c1" TYPE smallint`,
					},
					{
						Cmd:     `DROP SEQUENCE IF EXISTS "public"."posts_c1_seq"`,
						Reverse: `CREATE SEQUENCE IF NOT EXISTS "public"."posts_c1_seq" OWNED BY "public"."posts"."c1"`,
					},
					{
						Cmd:     `DROP SEQUENCE IF EXISTS "public"."previous_name"`,
						Reverse: `CREATE SEQUENCE IF NOT EXISTS "public"."previous_name" OWNED BY "public"."posts"."c2"`,
					},
				},
			},
		},
		// Add serial sequence.
		{
			changes: []sqlschema.SchemaChange{
				&sqlschema.ModifyTable{
					T: sqlschema.NewTable("posts").
						SetSchema(sqlschema.New("public")).
						AddColumns(
							sqlschema.NewColumn("c1").SetType(&SerialType{T: "serial"}),
							sqlschema.NewColumn("c2").SetType(&SerialType{T: "bigserial"}),
						),
					Changes: sqlschema.Changes{
						&sqlschema.ModifyColumn{
							From:   sqlschema.NewIntColumn("c1", "integer"),
							To:     sqlschema.NewColumn("c1").SetType(&SerialType{T: "serial"}),
							Change: sqlschema.ChangeType,
						},
						&sqlschema.ModifyColumn{
							From:   sqlschema.NewIntColumn("c2", "integer"),
							To:     sqlschema.NewColumn("c2").SetType(&SerialType{T: "bigserial"}),
							Change: sqlschema.ChangeType,
						},
					},
				},
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    true,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{
						Cmd:     `CREATE SEQUENCE IF NOT EXISTS "public"."posts_c1_seq" OWNED BY "public"."posts"."c1"`,
						Reverse: `DROP SEQUENCE IF EXISTS "public"."posts_c1_seq"`,
					},
					{
						Cmd:     `CREATE SEQUENCE IF NOT EXISTS "public"."posts_c2_seq" OWNED BY "public"."posts"."c2"`,
						Reverse: `DROP SEQUENCE IF EXISTS "public"."posts_c2_seq"`,
					},
					{
						Cmd:     `ALTER TABLE "public"."posts" ALTER COLUMN "c1" SET DEFAULT nextval('"public"."posts_c1_seq"'), ALTER COLUMN "c2" SET DEFAULT nextval('"public"."posts_c2_seq"'), ALTER COLUMN "c2" TYPE bigint`,
						Reverse: `ALTER TABLE "public"."posts" ALTER COLUMN "c2" DROP DEFAULT, ALTER COLUMN "c2" TYPE integer, ALTER COLUMN "c1" DROP DEFAULT`,
					},
				},
			},
		},
		// Change underlying sequence type.
		{
			changes: []sqlschema.SchemaChange{
				&sqlschema.ModifyTable{
					T: sqlschema.NewTable("posts").
						SetSchema(sqlschema.New("public")).
						AddColumns(
							sqlschema.NewColumn("c1").SetType(&SerialType{T: "serial"}),
							sqlschema.NewColumn("c2").SetType(&SerialType{T: "bigserial"}),
						),
					Changes: sqlschema.Changes{
						&sqlschema.ModifyColumn{
							From:   sqlschema.NewColumn("c1").SetType(&SerialType{T: "smallserial"}),
							To:     sqlschema.NewColumn("c1").SetType(&SerialType{T: "serial"}),
							Change: sqlschema.ChangeType,
						},
						&sqlschema.ModifyColumn{
							From:   sqlschema.NewColumn("c2").SetType(&SerialType{T: "serial"}),
							To:     sqlschema.NewColumn("c2").SetType(&SerialType{T: "bigserial"}),
							Change: sqlschema.ChangeType,
						},
					},
				},
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    true,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{
						Cmd:     `ALTER TABLE "public"."posts" ALTER COLUMN "c1" TYPE integer, ALTER COLUMN "c2" TYPE bigint`,
						Reverse: `ALTER TABLE "public"."posts" ALTER COLUMN "c2" TYPE integer, ALTER COLUMN "c1" TYPE smallint`,
					},
				},
			},
		},
		// Empty qualifier.
		{
			changes: []sqlschema.SchemaChange{
				&sqlschema.AddTable{
					T: sqlschema.NewTable("posts").
						SetSchema(sqlschema.New("test1")).
						AddColumns(
							sqlschema.NewEnumColumn("c1", sqlschema.EnumName("enum"), sqlschema.EnumValues("a"), sqlschema.EnumSchema(sqlschema.New("test2"))),
						),
				},
			},
			options: []sqlschema.PlanOption{
				func(o *sqlschema.PlanOptions) { o.SchemaQualifier = new(string) },
			},
			mock: func(m mock) {
				m.ExpectQuery(sqlutil.Escape("SELECT * FROM pg_type t JOIN pg_namespace n on t.typnamespace = n.oid WHERE t.typname = $1 AND t.typtype = 'e'")).
					WithArgs("enum").
					WillReturnRows(sqlmock.NewRows([]string{"name"}))
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    true,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{
						Cmd:     `CREATE TYPE "enum" AS ENUM ('a')`,
						Reverse: `DROP TYPE "enum"`,
					},
					{
						Cmd:     `CREATE TABLE "posts" ("c1" "enum" NOT NULL)`,
						Reverse: `DROP TABLE "posts"`,
					},
				},
			},
		},
		// Empty sequence qualifier.
		{
			changes: []sqlschema.SchemaChange{
				&sqlschema.ModifyTable{
					T: sqlschema.NewTable("posts").
						SetSchema(sqlschema.New("public")).
						AddColumns(
							sqlschema.NewColumn("c1").SetType(&SerialType{T: "serial"}),
						),
					Changes: sqlschema.Changes{
						&sqlschema.ModifyColumn{
							From:   sqlschema.NewIntColumn("c1", "integer"),
							To:     sqlschema.NewColumn("c1").SetType(&SerialType{T: "serial"}),
							Change: sqlschema.ChangeType,
						},
					},
				},
			},
			options: []sqlschema.PlanOption{
				func(o *sqlschema.PlanOptions) { o.SchemaQualifier = new(string) },
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    true,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{
						Cmd:     `CREATE SEQUENCE IF NOT EXISTS "posts_c1_seq" OWNED BY "posts"."c1"`,
						Reverse: `DROP SEQUENCE IF EXISTS "posts_c1_seq"`,
					},
					{
						Cmd:     `ALTER TABLE "posts" ALTER COLUMN "c1" SET DEFAULT nextval('"posts_c1_seq"')`,
						Reverse: `ALTER TABLE "posts" ALTER COLUMN "c1" DROP DEFAULT`,
					},
				},
			},
		},
		// Empty index qualifier.
		{
			changes: []sqlschema.SchemaChange{
				&sqlschema.ModifyTable{
					T: sqlschema.NewTable("posts").
						SetSchema(sqlschema.New("public")).
						AddColumns(
							sqlschema.NewIntColumn("c", "int"),
						),
					Changes: sqlschema.Changes{
						&sqlschema.AddIndex{
							I: sqlschema.NewIndex("i").AddColumns(sqlschema.NewIntColumn("c", "int")),
						},
					},
				},
			},
			options: []sqlschema.PlanOption{
				func(o *sqlschema.PlanOptions) { o.SchemaQualifier = new(string) },
			},
			wantPlan: &sqlschema.Plan{
				Reversible:    true,
				Transactional: true,
				Changes: []*sqlschema.Change{
					{
						Cmd:     `CREATE INDEX "i" ON "posts" ("c")`,
						Reverse: `DROP INDEX "i"`,
					},
				},
			},
		},
		// Empty qualifier in multi-schema mode should fail.
		{
			changes: []sqlschema.SchemaChange{
				&sqlschema.AddTable{T: sqlschema.NewTable("t1").SetSchema(sqlschema.New("s1")).AddColumns(sqlschema.NewIntColumn("a", "int"))},
				&sqlschema.AddTable{T: sqlschema.NewTable("t2").SetSchema(sqlschema.New("s2")).AddColumns(sqlschema.NewIntColumn("a", "int"))},
			},
			options: []sqlschema.PlanOption{
				func(o *sqlschema.PlanOptions) { o.SchemaQualifier = new(string) },
			},
			wantErr: true,
		},
	}
	for i, tt := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			db, mk, err := sqlmock.New()
			require.NoError(t, err)
			m := mock{mk}
			m.version("130000")
			if tt.mock != nil {
				tt.mock(m)
			}
			drv, err := Open(db)
			require.NoError(t, err)
			plan, err := drv.PlanChanges(tt.changes, tt.options...)
			if tt.wantErr {
				require.Error(t, err, "expect plan to fail")
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.wantPlan.Reversible, plan.Reversible)
			require.Equal(t, tt.wantPlan.Transactional, plan.Transactional)
			for i, c := range plan.Changes {
				require.Equal(t, tt.wantPlan.Changes[i].Cmd, c.Cmd)
				require.Equal(t, tt.wantPlan.Changes[i].Reverse, c.Reverse)
			}
		})
	}
}

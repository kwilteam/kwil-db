package mocks

const (
	CreateTable1DDL             = `CREATE TABLE "xf9b03342f27548fb1d86b5f2094be2f5db7dc06f389ed6daf8cdbbe3"."table1" (col1 text, col2 int4);`
	AlterTable1AddPrimaryKeyDDL = `ALTER TABLE "xf9b03342f27548fb1d86b5f2094be2f5db7dc06f389ed6daf8cdbbe3"."table1" ADD PRIMARY KEY (col1);`
	AlterTable1AddConstraintDDL = `ALTER TABLE "xf9b03342f27548fb1d86b5f2094be2f5db7dc06f389ed6daf8cdbbe3"."table1" ADD CONSTRAINT "c610486b430754ec33ccb819af1fc7a1fc68e87cd25a08133fdf5a15be4a181" CHECK (col2 >= 1);`
	CreateTable2DDL             = `CREATE TABLE "xf9b03342f27548fb1d86b5f2094be2f5db7dc06f389ed6daf8cdbbe3"."table2" (col1 text, col3 boolean);`
	AlterTable2AddPrimaryKeyDDL = `ALTER TABLE "xf9b03342f27548fb1d86b5f2094be2f5db7dc06f389ed6daf8cdbbe3"."table2" ADD PRIMARY KEY (col1);`
	CreateIndexDDL              = `CREATE INDEX my_index ON "xf9b03342f27548fb1d86b5f2094be2f5db7dc06f389ed6daf8cdbbe3"."table1" USING btree (col1, col2);`
)

// to make it easier to test
var (
	ALL_MOCK_DDL = []string{
		CreateTable1DDL,
		AlterTable1AddPrimaryKeyDDL,
		AlterTable1AddConstraintDDL,
		CreateTable2DDL,
		AlterTable2AddPrimaryKeyDDL,
		CreateIndexDDL,
	}
)

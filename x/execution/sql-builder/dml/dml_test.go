package dml_test

import (
	"kwil/x/execution/dto"
	"kwil/x/execution/mocks"
	"kwil/x/execution/sql-builder/dml"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_DML(t *testing.T) {
	// test create insert
	stmt, err := dml.BuildInsert("kwil", "test", []any{"id", "name"})
	if err != nil {
		t.Errorf("failed to build insert: %v", err)
	}

	if !assert.Contains(t, stmt, `INSERT INTO "kwil"."test" ("id", "name") VALUES ($1, $2)`) {
		t.Errorf("missing insert statement: %v", stmt)
	}

	// test create update
	var params []*dto.Parameter
	params = append(params, &mocks.Parameter1)
	params = append(params, &mocks.Parameter2)

	var wheres []*dto.WhereClause
	wheres = append(wheres, &mocks.WhereClause1)
	wheres = append(wheres, &mocks.WhereClause2)
	stmt, err = dml.BuildUpdate("kwil", "test", params, wheres)
	if err != nil {
		t.Errorf("failed to build update: %v", err)
	}

	if !assert.Contains(t, stmt, `UPDATE "kwil"."test" SET "col1"=$1,"col2"=$2 WHERE (("col3" IS FALSE) AND ("col1" IS FALSE))`) {
		t.Errorf("error generating update statement.  statement: %v", stmt)
	}

	// test create delete
	stmt, err = dml.BuildDelete("kwil", "test", wheres)
	if err != nil {
		t.Errorf("failed to build delete: %v", err)
	}

	if !assert.Contains(t, stmt, `DELETE FROM "kwil"."test" WHERE (("col3" IS FALSE) AND ("col1" IS FALSE))`) {
		t.Errorf("error generating delete statement.  statement: %v", stmt)
	}

}

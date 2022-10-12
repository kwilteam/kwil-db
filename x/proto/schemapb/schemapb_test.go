package schemapb

import (
	"kwil/x/schemadef/pgschema"
	"kwil/x/schemadef/sqlschema"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConversion(t *testing.T) {
	model := `table "table" {
  schema  = schema.schema
  comment = "table comment"
  column "col" {
    type    = integer
    comment = "column comment"
  }
  column "age" {
    type = integer
  }
  column "price" {
    type = int
  }
  column "account_name" {
    type = varchar(32)
  }
  column "varchar_length_is_not_required" {
    type = varchar
  }
  column "character_varying_length_is_not_required" {
    type = character_varying
  }
  column "tags" {
    type = hstore
  }
  column "created_at" {
    type    = timestamp(4)
    default = sql("current_timestamp(4)")
  }
  column "updated_at" {
    type    = time
    default = sql("current_time")
  }
  primary_key {
    columns = [column.col]
  }
  foreign_key "accounts" {
    columns     = [column.account_name]
    ref_columns = [table.accounts.column.name]
    on_update   = NO_ACTION
    on_delete   = SET_NULL
  }
  index "index" {
    columns = [column.col, column.age]
    comment = "index comment"
    type    = HASH
    where   = "active"
  }
  check "positive price" {
    expr = "price > 0"
  }
}
table "accounts" {
  schema = schema.schema
  column "name" {
    type = varchar(32)
  }
  column "type" {
    type = enum.account_type
  }
  primary_key {
    columns = [column.name]
  }
}
enum "account_type" {
  schema = schema.schema
  values = ["private", "business"]
}
schema "schema" {
}
`
	r := &sqlschema.Realm{}
	err := pgschema.EvalHCLBytes([]byte(model), r, nil)
	require.Nil(t, err)
	pb := FromRealm(r)
	conv := ToRealm(pb)
	data, err := pgschema.MarshalHCL(conv)
	require.Nil(t, err)
	require.Equal(t, model, string(data))
}

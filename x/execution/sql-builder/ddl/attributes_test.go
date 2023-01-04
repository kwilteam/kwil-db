package ddlbuilder_test

import (
	ddlb "kwil/x/execution/sql-builder/ddl"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Attributes(t *testing.T) {
	ab := ddlb.NewAttributeBuilder()

	// primary key
	pk := ab.Schema("my_schema").Table("my_table").PrimaryKey("id").Build()
	if !assert.Contains(t, pk, `ALTER TABLE "my_schema"."my_table" ADD PRIMARY KEY (id);`) {
		t.Error("invalid ddl built for primary_key:", pk)
	}

	// default int
	ab = ddlb.NewAttributeBuilder()
	def := ab.Schema("my_schema").Table("my_table").Default("id", 1).Build()

	if !assert.Contains(t, def, `ALTER TABLE "my_schema"."my_table" ALTER COLUMN id SET DEFAULT 1;`) {
		t.Error("invalid ddl built for default int", def)
	}

	// default string
	ab = ddlb.NewAttributeBuilder()
	def = ab.Schema("my_schema").Table("my_table").Default("name", "bennan").Build()

	if !assert.Contains(t, def, `ALTER TABLE "my_schema"."my_table" ALTER COLUMN name SET DEFAULT 'bennan';`) {
		t.Error("invalid ddl built for default string", def)
	}

	// not null
	ab = ddlb.NewAttributeBuilder()
	nn := ab.Schema("my_schema").Table("my_table").NotNull("id").Build()

	if !assert.Contains(t, nn, `ALTER TABLE "my_schema"."my_table" ALTER COLUMN id SET NOT NULL;`) {
		t.Error("invalid ddl built for not null", nn)
	}

	// unique
	ab = ddlb.NewAttributeBuilder()
	unq := ab.Schema("my_schema").Table("my_table").Unique("id").Build()

	if !assert.Contains(t, unq, `ALTER TABLE "my_schema"."my_table" ADD CONSTRAINT "c77ec7a8795765f70f42f54dc5d4adca040d8920b3622d2d21278743628f5ff" UNIQUE (id);`) {
		t.Error("invalid ddl built for unique", unq)
	}

	// min
	ab = ddlb.NewAttributeBuilder()
	min := ab.Schema("my_schema").Table("my_table").Min("id", 1).Build()

	if !assert.Contains(t, min, `ALTER TABLE "my_schema"."my_table" ADD CONSTRAINT "cd761e491dfd5b0e2b3d5b689a1ef94d35d03655f7223ab983809c2796a4a04" CHECK (id >= 1);`) {
		t.Error("invalid ddl built for min", min)
	}

	// max
	ab = ddlb.NewAttributeBuilder()
	max := ab.Schema("my_schema").Table("my_table").Max("id", 1).Build()

	if !assert.Contains(t, max, `ALTER TABLE "my_schema"."my_table" ADD CONSTRAINT "c6106b253f4f2758a02ece6153da1ca0b6fcff9d1a28298caca2f3caf9a1f2d" CHECK (id <= 1);`) {
		t.Error("invalid ddl built for max", max)
	}

	// min length
	ab = ddlb.NewAttributeBuilder()
	minlen := ab.Schema("my_schema").Table("my_table").MinLength("name", 1).Build()

	if !assert.Contains(t, minlen, `ALTER TABLE "my_schema"."my_table" ADD CONSTRAINT "c67f03dba56506a3fd1829b70683fe2ed68e6fc5e8cdf1096481f46abf962e2" CHECK (LENGTH(name) >= 1);`) {
		t.Error("invalid ddl built for min length", minlen)
	}

	// max length
	ab = ddlb.NewAttributeBuilder()
	maxlen := ab.Schema("my_schema").Table("my_table").MaxLength("name", 1).Build()

	if !assert.Contains(t, maxlen, `ALTER TABLE "my_schema"."my_table" ADD CONSTRAINT "c5c2001c7d84b90e0d5281fc6c674557361f04908f0dc8e3f6278f810cb9769" CHECK (LENGTH(name) <= 1);`) {
		t.Error("invalid ddl built for max length", maxlen)
	}
}

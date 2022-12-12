package ddlbuilder_test

import (
	ddlb "kwil/x/sqlx/ddl_builder"
	"testing"
)

func Test_Attributes(t *testing.T) {
	ab := ddlb.NewAttributeBuilder()

	// primary key
	pk := ab.Schema("my_schema").Table("my_table").PrimaryKey("id").Build()
	if pk != "ALTER TABLE my_schema.my_table ADD PRIMARY KEY (id);" {
		t.Fatal("invalid ddl built:", pk)
	}

	// default int
	ab = ddlb.NewAttributeBuilder()
	def := ab.Schema("my_schema").Table("my_table").Default("id", 1).Build()
	if def != "ALTER TABLE my_schema.my_table ALTER COLUMN id SET DEFAULT 1;" {
		t.Fatal("invalid ddl built:", def)
	}

	// default string
	ab = ddlb.NewAttributeBuilder()
	def = ab.Schema("my_schema").Table("my_table").Default("name", "bennan").Build()
	if def != "ALTER TABLE my_schema.my_table ALTER COLUMN name SET DEFAULT 'bennan';" {
		t.Fatal("invalid ddl built:", def)
	}

	// not null
	ab = ddlb.NewAttributeBuilder()
	nn := ab.Schema("my_schema").Table("my_table").NotNull("id").Build()
	if nn != "ALTER TABLE my_schema.my_table ALTER COLUMN id SET NOT NULL;" {
		t.Fatal("invalid ddl built:", nn)
	}

	// unique
	ab = ddlb.NewAttributeBuilder()
	unq := ab.Schema("my_schema").Table("my_table").Unique("id").Build()
	if unq != "ALTER TABLE my_schema.my_table ADD CONSTRAINT 77ec7a8795765f70f42f54dc5d4adca040d8920b3622d2d21278743628f5ff7 UNIQUE (id);" {
		t.Fatal("invalid ddl built:", unq)
	}

	// min
	ab = ddlb.NewAttributeBuilder()
	min := ab.Schema("my_schema").Table("my_table").Min("id", 1).Build()
	if min != "ALTER TABLE my_schema.my_table ADD CONSTRAINT d761e491dfd5b0e2b3d5b689a1ef94d35d03655f7223ab983809c2796a4a04f CHECK (id >= 1);" {
		t.Fatal("invalid ddl built:", min)
	}

	// max
	ab = ddlb.NewAttributeBuilder()
	max := ab.Schema("my_schema").Table("my_table").Max("id", 1).Build()
	if max != "ALTER TABLE my_schema.my_table ADD CONSTRAINT 6106b253f4f2758a02ece6153da1ca0b6fcff9d1a28298caca2f3caf9a1f2d2 CHECK (id <= 1);" {
		t.Fatal("invalid ddl built:", max)
	}

	// min length
	ab = ddlb.NewAttributeBuilder()
	minlen := ab.Schema("my_schema").Table("my_table").MinLength("name", 1).Build()
	if minlen != "ALTER TABLE my_schema.my_table ADD CONSTRAINT 67f03dba56506a3fd1829b70683fe2ed68e6fc5e8cdf1096481f46abf962e2a CHECK (LENGTH(name) >= 1);" {
		t.Fatal("invalid ddl built:", minlen)
	}

	// max length
	ab = ddlb.NewAttributeBuilder()
	maxlen := ab.Schema("my_schema").Table("my_table").MaxLength("name", 1).Build()
	if maxlen != "ALTER TABLE my_schema.my_table ADD CONSTRAINT 5c2001c7d84b90e0d5281fc6c674557361f04908f0dc8e3f6278f810cb97690 CHECK (LENGTH(name) <= 1);" {
		t.Fatal("invalid ddl built:", maxlen)
	}
}

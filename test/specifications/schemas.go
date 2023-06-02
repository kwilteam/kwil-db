package specifications

const prependedFilePath = "./test-data/"

func getSchemaFilePath(schemaName string) string {
	return prependedFilePath + schemaName + ".kf"
}

type testSchema struct {
	FileName string
}

func (s *testSchema) GetFilePath() string {
	return getSchemaFilePath(s.FileName)
}

var (
	schema_testdb = &testSchema{
		FileName: "test_db",
	}
	schema_invalidSQLSyntax = &testSchema{
		FileName: "invalid_sql_syntax",
	}

	schema_invalidSQLSyntaxFixed = &testSchema{
		FileName: "invalid_sql_syntax_fixed",
	}
)

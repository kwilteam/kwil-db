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
	SchemaTestDB = &testSchema{
		FileName: "test_db",
	}
	schemaInvalidSqlSyntax = &testSchema{
		FileName: "invalid_sql_syntax",
	}

	schemaInvalidSqlSyntaxFixed = &testSchema{
		FileName: "invalid_sql_syntax_fixed",
	}
	schemaInvalidExtensionInit = &testSchema{
		FileName: "invalid_extension_init",
	}
	userDB = &testSchema{
		FileName: "users",
	}
	TypesDB = &testSchema{
		FileName: "types",
	}
	ContextualVarsDB = &testSchema{
		FileName: "contextual_vars",
	}
	LogDB = &testSchema{
		FileName: "logs",
	}
)

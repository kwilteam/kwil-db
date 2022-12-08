package query

import (
	"gopkg.in/yaml.v2"
	"os"
	"testing"
)

const _YAML_QUERY_TEST_FILE = "test_queries.yaml"

func Test_Query_From_Map(t *testing.T) {
	bts, err := os.ReadFile(_YAML_QUERY_TEST_FILE)
	if err != nil {
		t.Fatal(err)
	}

	m := make(map[any]any)
	err = yaml.Unmarshal(bts, &m)
	if err != nil {
		t.Fatal(err)
	}

	//defs, err := LoadFromMap(m)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//
	//validate_query(t, defs)
}

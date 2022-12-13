package hasura

import "testing"

func Test_snakeCase(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "with space",
			args: args{"With Space"},
			want: "with_space",
		},
		{
			name: "without space",
			args: args{"WithoutSpace"},
			want: "withoutspace",
		},
		{
			name: "simple",
			args: args{"simple"},
			want: "simple",
		},
		{
			name: "with_underscore",
			args: args{"with_underscore"},
			want: "with_underscore",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := snakeCase(tt.args.name); got != tt.want {
				t.Errorf("convertHasuraTableName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_customHasuraTableName(t *testing.T) {
	type args struct {
		schema string
		table  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "normal",
			args: args{
				schema: "schema",
				table:  "table",
			},
			want: "schema_table",
		},
		{
			name: "table with space",
			args: args{
				schema: "schema",
				table:  "Author Details",
			},
			want: "schema_author_details",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := customTableName(tt.args.schema, tt.args.table); got != tt.want {
				t.Errorf("customTableName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_queryToExplain(t *testing.T) {
	type args struct {
		query string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no query head",
			args: args{"{wallets {balance}}"},
			want: `{"query": {"query": "{wallets {balance}}"}}`,
		},
		{
			name: "no query name",
			args: args{"query {wallets {balance}}"},
			want: `{"query": {"query": "{wallets {balance}}"}}`,
		},
		{
			name: "normal",
			args: args{"query test {wallets {balance}}"},
			want: `{"query": {"query": "query test {wallets {balance}}", "operationName": "test"}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := queryToExplain(tt.args.query); got != tt.want {
				t.Errorf("queryToExplain() = %v, want %v", got, tt.want)
			}
		})
	}
}

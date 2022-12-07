package server

import "testing"

func Test_isMutation(t *testing.T) {
	type args struct {
		query string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{
			name: "normal",
			args: args{"mutation insert {insert_profiles(objects: {age: 21}) { affected_rows }}"},
			want: true,
		},
		{
			name: "multi mutation",
			args: args{"mutation insert {insert_profiles(objects: {age: 21}) { affected_rows }}\nmutation insert {insert_profiles(objects: {age: 21}) { affected_rows }}"},
			want: true,
		},

		{
			name: "hybrid query",
			args: args{"query {profiles {name}}\nmutation insert {insert_profiles(objects: {age: 21}) { affected_rows }}"},
			want: true,
		},
		{
			name: "mutation is arguemet",
			// will fail
			args: args{"mutation insert {insert_profiles(objects: {address: \"st mutation sf\"}) { affected_rows }}"},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isMutation(tt.args.query); got != tt.want {
				t.Errorf("isMutation() = %v, want %v", got, tt.want)
			}
		})
	}
}

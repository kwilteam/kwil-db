package misc

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
		{
			name: "normal",
			args: args{"mutation normal {insert_profiles(objects: {age: 21}) { affected_rows }}"},
			want: true,
		},
		{
			name: "multi mutation",
			args: args{"mutation multi1 {insert_profiles(objects: {age: 21}) { affected_rows }}\nmutation multi2 {insert_profiles(objects: {age: 21}) { affected_rows }}"},
			want: true,
		},

		{
			name: "hybrid query",
			args: args{"query {profiles {name}}\nmutation hybrid {insert_profiles(objects: {age: 21}) { affected_rows }}"},
			want: true,
		},
		{
			name: "mutation in arg",
			args: args{`query mut{profiles(name: "arg mutation ") { name }}`},
			want: false,
		},
		{
			name: "mutation in body",
			args: args{"query mut{profiles(age: 20) { name mutation }}"},
			want: false,
		},
		{
			name: "close brackets in arg with mutation in body",
			args: args{`query mut{profiles(name: "}") { name mutation }}`},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsMutation(tt.args.query); got != tt.want {
				t.Errorf("isMutation() = %v, want %v", got, tt.want)
			}
		})
	}
}

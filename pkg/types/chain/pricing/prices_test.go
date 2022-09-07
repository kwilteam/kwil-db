package pricing

import (
	"reflect"
	"testing"
)

func Test_parsePrices(t *testing.T) {

	tInput := operations{
		Database: cruds{
			Create: "1",
			Modify: "-1",
			Delete: "-1",
		},
		Table: cruds{
			Create: "2",
			Modify: "3",
			Delete: "1",
		},
		Role: cruds{
			Create: "2",
			Modify: "-1",
			Delete: "1",
		},
		Query: cruds{
			Create: "2",
			Modify: "3",
			Delete: "-1",
		},
	}

	// I have commented the negatives out since they get filtered, but kept them there for reference in case we add them later
	expectedMap := map[int16]int64{
		0: 1, // database create
		//256: -1, // database modify
		//512: -1, // database delete
		1:   2, // table create
		257: 3, // table modify
		513: 1, // table delete
		2:   2, // role create
		//258: -1, // role modify
		514: 1, // role delete
		3:   2, // query create
		259: 3, // query modify
		//515: -1, // query delete
	}

	type args struct {
		p *operations
	}
	tests := []struct {
		name string
		args args
		want *map[int16]int64
	}{
		{
			name: "parsePrices",
			args: args{
				p: &tInput,
			},
			want: &expectedMap,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parsePrices(tt.args.p); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsePrices() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_determineOp(t *testing.T) {
	type args struct {
		op string
	}
	tests := []struct {
		name    string
		args    args
		want    byte
		wantErr bool
	}{
		{
			name: "determineOp_database",
			args: args{
				op: "database",
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "determineOp_table",
			args: args{
				op: "table",
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "determineOp_role",
			args: args{
				op: "role",
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "determineOp_query",
			args: args{
				op: "query",
			},
			want:    3,
			wantErr: false,
		},
		{
			name: "determineOp_invalid",
			args: args{
				op: "invalid",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := determineOp(tt.args.op)
			if (err != nil) != tt.wantErr {
				t.Errorf("determineOp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("determineOp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_determineCRUD(t *testing.T) {
	type args struct {
		crud string
	}
	tests := []struct {
		name    string
		args    args
		want    byte
		wantErr bool
	}{
		{
			name: "determineCRUD_create",
			args: args{
				crud: "create",
			},
			want:    0,
			wantErr: false,
		},
		{
			name: "determineCRUD_modify",
			args: args{
				crud: "modify",
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "determineCRUD_delete",
			args: args{
				crud: "delete",
			},
			want:    2,
			wantErr: false,
		},
		{
			name: "determineCRUD_invalid",
			args: args{
				crud: "invalid",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := determineCRUD(tt.args.crud)
			if (err != nil) != tt.wantErr {
				t.Errorf("determineCRUD() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("determineCRUD() = %v, want %v", got, tt.want)
			}
		})
	}
}

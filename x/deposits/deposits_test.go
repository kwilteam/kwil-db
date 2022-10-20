package deposits

import (
	"reflect"
	"testing"
)

func Test_splitBlocks(t *testing.T) {
	type args struct {
		start     int64
		end       int64
		chunkSize int64
	}
	tests := []struct {
		name string
		args args
		want []chunk
	}{
		{
			name: "split_blocks",
			args: args{
				start:     0,
				end:       350000,
				chunkSize: 100000,
			},
			want: []chunk{
				{
					0,
					99999,
				},
				{
					100000,
					199999,
				},
				{
					200000,
					299999,
				},
				{
					300000,
					349999,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := splitBlocks(tt.args.start, tt.args.end, tt.args.chunkSize); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitBlocks() = %v, want %v", got, tt.want)
			}
		})
	}
}

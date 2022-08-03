package wal

/*import (
	//keepertest "kwil-cosmos/testutil/keeper"
	//"kwil-cosmos/x/kwil/keeper"
	"reflect"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tidwall/wal"
)

func TestNewBlockWal(t *testing.T) {
	//_, ctx := keepertest.KwilKeeper(t)
	type args struct {
		ctx  sdk.Context
		opts *wal.Options
	}
	tests := []struct {
		name    string
		args    args
		want    keeper.Wal
		wantErr bool
	}{
		{
			name: "Creation of Wal",
			args: args{
				ctx: ctx,
			},
			want: keeper.Wal{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := keeper.NewBlockWal(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBlockWal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewBlockWal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConcat(t *testing.T) {
	type args struct {
		strArr []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Concat strings",
			args: args{
				strArr: []string{"kw", "il"},
			},
			want: "kwil",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := keeper.Concat(tt.args.strArr); got != tt.want {
				t.Errorf("Concat() = %v, want %v", got, tt.want)
			}
		})
	}
}
*/

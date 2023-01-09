package simple_checker

import (
	"context"
	"fmt"
	"github.com/alexliesenfeld/health"
	"kwil/x/healthcheck"
	"reflect"
	"testing"
	"time"
)

func TestSimpleChecker_Check(t *testing.T) {
	type fields struct {
		Ck health.Checker
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   healthcheck.Result
	}{
		{
			name: "test status up",
			fields: fields{
				Ck: health.NewChecker(
					health.WithPeriodicCheck(0, 0, health.Check{
						Name: "test",
						Check: func(ctx context.Context) error {
							return nil
						},
					})),
			},
			args: args{
				ctx: context.Background(),
			},
			want: healthcheck.Result{Status: "SERVING"},
		},
		{
			name: "test status down",
			fields: fields{
				Ck: health.NewChecker(
					health.WithPeriodicCheck(0, 0, health.Check{
						Name: "test",
						Check: func(ctx context.Context) error {
							return fmt.Errorf("test error")
						},
					})),
			},
			args: args{
				ctx: context.Background(),
			},
			want: healthcheck.Result{Status: "NOT_SERVING"},
		},
		{
			name: "test status unknown",
			fields: fields{
				// status before first check is "UNKNOWN"
				Ck: health.NewChecker(
					health.WithPeriodicCheck(10*time.Second, 10*time.Second, health.Check{
						Name: "test",
						Check: func(ctx context.Context) error {
							return nil
						},
					})),
			},
			args: args{
				ctx: context.Background(),
			},
			want: healthcheck.Result{Status: "UNKNOWN"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &SimpleChecker{
				Ck: tt.fields.Ck,
			}
			if got := c.Check(tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Check() = %v, want %v", got, tt.want)
			}
		})
	}
}

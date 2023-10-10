package validators

import (
	"testing"
)

func Test_joinReq(t *testing.T) {
	type fields struct {
		pubkey     []byte
		power      int64
		validators map[string]bool
	}
	tests := []struct {
		name         string
		fields       fields
		wantVotes    int
		wantRequired int
		// approverChecks map[string][2]bool
	}{
		{
			name: "empty",
			fields: fields{
				pubkey:     []byte{1, 2, 3},
				power:      1,
				validators: nil,
			},
			wantVotes:    0,
			wantRequired: 0,
		},
		{
			name: "three at threshold",
			fields: fields{
				pubkey: []byte{1, 2, 3},
				power:  1,
				validators: map[string]bool{ // 3 validators => require 2
					"000": true,  // approved
					"001": false, // not yet
					"002": true,
				},
			},
			wantVotes:    2,
			wantRequired: 2,
		},
		{
			name: "two",
			fields: fields{
				pubkey: []byte{1, 2, 3},
				power:  1,
				validators: map[string]bool{ // 2 validators => require 2 (50% < 66.667%)
					"000": true,  // approved
					"001": false, // not yet
				},
			},
			wantVotes:    1,
			wantRequired: 2,
		},
		{
			name: "one",
			fields: fields{
				pubkey: []byte{1, 2, 3},
				power:  1,
				validators: map[string]bool{ // 1 validators => require 1
					"000": true, // approved
				},
			},
			wantVotes:    1,
			wantRequired: 1,
		},
		{
			name: "four, not approved",
			fields: fields{
				pubkey: []byte{1, 2, 3},
				power:  1,
				validators: map[string]bool{ // 4 validators => require 3
					"000": true,
					"001": false,
					"002": true,
					"003": false,
				},
			},
			wantVotes:    2,
			wantRequired: 3,
		},
		{
			name: "four, approved",
			fields: fields{
				pubkey: []byte{1, 2, 3},
				power:  1,
				validators: map[string]bool{ // 4 validators => require 3
					"000": true,
					"001": false,
					"002": true,
					"003": true, // threshold reached
				},
			},
			wantVotes:    3,
			wantRequired: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jr := &joinReq{
				pubkey:     tt.fields.pubkey,
				power:      tt.fields.power,
				validators: tt.fields.validators,
			}
			if votes := jr.votes(); votes != tt.wantVotes {
				t.Errorf("joinReq.votes() = %v, want %v", votes, tt.wantVotes)
			}
			if req := jr.requiredVotes(); req != tt.wantRequired {
				t.Errorf("joinReq.requiredVotes() = %v, want %v", req, tt.wantRequired)
			}
			ok, eligible := jr.approve([]byte("XXX"))
			if ok || eligible {
				t.Errorf("approval allowed from unauthorized validator")
			}
			for pk, already := range jr.validators {
				repeat, eligible := jr.approve([]byte(pk))
				if !eligible {
					t.Fatal("expected validator to be elegible but was not")
				}
				if repeat != already {
					t.Errorf("existence of approval mismatched")
				}
				repeat, _ = jr.approve([]byte(pk))
				if !repeat {
					t.Errorf("repeat approve not recognized")
				}
			}
		})
	}
}

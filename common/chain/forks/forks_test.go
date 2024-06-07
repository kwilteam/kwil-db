package forks_test

import (
	"slices"
	"testing"

	"github.com/kwilteam/kwil-db/common/chain/forks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func intPtr(i uint64) *uint64 {
	return &i
}

func TestForks_Checks(t *testing.T) {
	tests := []struct {
		name           string
		height         uint64
		HaltHeight     *uint64
		Extended       map[string]*uint64
		wantHaltBegins bool
		wantHaltIs     bool
		at             []string
		by             []string
		byHeights      []uint64
	}{
		{
			name:       "genesis, below both, different",
			height:     0,
			HaltHeight: intPtr(10),
			Extended: map[string]*uint64{
				"extended": intPtr(6),
			},
			wantHaltBegins: false,
			wantHaltIs:     false,
			at:             nil,
			by:             nil,
			byHeights:      nil,
		},
		{
			name:       "at extended, below named",
			height:     6,
			HaltHeight: intPtr(10),
			Extended: map[string]*uint64{
				"extended": intPtr(6),
			},
			wantHaltBegins: false,
			wantHaltIs:     false,
			at:             []string{"extended"},
			by:             []string{"extended"},
			byHeights:      []uint64{6},
		},
		{
			name:       "above extended, below named",
			height:     7,
			HaltHeight: intPtr(10),
			Extended: map[string]*uint64{
				"extended": intPtr(6),
			},
			wantHaltBegins: false,
			wantHaltIs:     false,
			at:             nil,
			by:             []string{"extended"},
			byHeights:      []uint64{6},
		},
		{
			name:       "above extended, at named",
			height:     10,
			HaltHeight: intPtr(10),
			Extended: map[string]*uint64{
				"extended": intPtr(6),
			},
			wantHaltBegins: true,
			wantHaltIs:     true,
			at:             []string{"halt"},
			by:             []string{"halt", "extended"},
			byHeights:      []uint64{10, 6},
		},
		{
			name:           "above named",
			height:         11,
			HaltHeight:     intPtr(10),
			Extended:       map[string]*uint64{},
			wantHaltBegins: false,
			wantHaltIs:     true,
			at:             nil,
			by:             []string{"halt"},
			byHeights:      []uint64{10},
		},
		{
			name:           "at named, no extended",
			height:         10,
			HaltHeight:     intPtr(10),
			Extended:       map[string]*uint64{},
			wantHaltBegins: true,
			wantHaltIs:     true,
			at:             []string{"halt"},
			by:             []string{"halt"},
			byHeights:      []uint64{10},
		},
		{
			name:           "named inactive",
			height:         10,
			HaltHeight:     nil,
			Extended:       map[string]*uint64{},
			wantHaltBegins: false,
			wantHaltIs:     false,
			at:             nil,
			by:             nil,
			byHeights:      nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := &forks.Forks{
				HaltHeight: tt.HaltHeight,
				Extended:   tt.Extended,
			}
			assert.Equal(t, tt.wantHaltBegins, fs.BeginsHalt(tt.height))
			assert.Equal(t, tt.wantHaltIs, fs.IsHalt(tt.height))

			assert.Equal(t, tt.at, fs.ActivatesAt(tt.height))
			activatedBy := fs.ActivatedBy(tt.height)
			var activatedNames []string
			var activatedHeights []uint64
			for _, act := range activatedBy {
				activatedNames = append(activatedNames, act.Name)
				activatedHeights = append(activatedHeights, act.Height)
			}
			assert.Equal(t, tt.by, activatedNames)
			assert.Equal(t, tt.byHeights, activatedHeights)
		})
	}
}

func TestForks_Height(t *testing.T) {
	HaltHeight := intPtr(10)
	ExtendedHeight := intPtr(6)

	fs := &forks.Forks{
		HaltHeight: HaltHeight,
		Extended:   map[string]*uint64{"extended": ExtendedHeight},
	}

	hp := fs.ForkHeight(forks.ForkHalt)
	require.NotNil(t, hp)
	assert.Equal(t, *HaltHeight, *hp)

	hp = fs.ForkHeight("notdefined")
	require.Nil(t, hp)

	hp = fs.ForkHeight("extended")
	require.NotNil(t, hp)
	assert.Equal(t, *ExtendedHeight, *hp)

	hp = fs.ForkHeight("unknown")
	require.Nil(t, hp)
}

func TestForks_FromMap(t *testing.T) {
	m := map[string]*uint64{
		forks.ForkHalt: intPtr(10),
		"extended":     intPtr(6),
	}

	fs := forks.NewForks(m)

	require.NotNil(t, fs.HaltHeight)
	assert.Equal(t, *fs.HaltHeight, uint64(10))

	hp := fs.Extended["extended"]
	assert.NotNil(t, hp)
	assert.Equal(t, *hp, uint64(6))

	// named forks are not in the extended map
	hp = fs.Extended[forks.ForkHalt]
	assert.Nil(t, hp)
}

func TestForks_String(t *testing.T) {
	m := map[string]*uint64{
		// forks.ForkHalt: nil, // disabled
		"extended":  intPtr(6),
		"alpha":     intPtr(1),
		"atGenesis": intPtr(0),
	}

	fs := forks.NewForks(m)

	str := fs.String()
	assert.Equal(t, `- halt: <nil> (disabled)
- atGenesis: 0 (genesis)
- alpha: 1
- extended: 6`, str)
}

func TestForks_sort(t *testing.T) {
	activeForks := []*forks.Fork{
		{
			Name:   "a",
			Height: 2,
		},
		{
			Name:   "b",
			Height: 2,
		},
		{
			Name:   "d",
			Height: 6,
		},
		{
			Name:   "c",
			Height: 3,
		},
	}
	wantSorted := []*forks.Fork{activeForks[0], activeForks[1], activeForks[3], activeForks[2]}
	slices.SortStableFunc(activeForks, forks.ForkSortFunc)
	assert.Equal(t, activeForks, wantSorted)
}

func TestForkSortFunc(t *testing.T) {
	tests := []struct {
		name     string
		forks    []*forks.Fork
		expected []*forks.Fork
	}{
		{
			name:     "empty",
			forks:    []*forks.Fork{},
			expected: []*forks.Fork{},
		},
		{
			name:     "single",
			forks:    []*forks.Fork{{Name: "a", Height: 1}},
			expected: []*forks.Fork{{Name: "a", Height: 1}},
		},
		{
			name: "sorted",
			forks: []*forks.Fork{
				{Name: "a", Height: 1},
				{Name: "b", Height: 2},
				{Name: "c", Height: 3},
			},
			expected: []*forks.Fork{
				{Name: "a", Height: 1},
				{Name: "b", Height: 2},
				{Name: "c", Height: 3},
			},
		},
		{
			name: "unsorted",
			forks: []*forks.Fork{
				{Name: "c", Height: 3},
				{Name: "a", Height: 1},
				{Name: "b", Height: 2},
			},
			expected: []*forks.Fork{
				{Name: "a", Height: 1},
				{Name: "b", Height: 2},
				{Name: "c", Height: 3},
			},
		},
		{
			name: "duplicate heights",
			forks: []*forks.Fork{
				{Name: "a", Height: 1},
				{Name: "b", Height: 2},
				{Name: "c", Height: 2},
				{Name: "d", Height: 3},
			},
			expected: []*forks.Fork{
				{Name: "a", Height: 1},
				{Name: "b", Height: 2},
				{Name: "c", Height: 2},
				{Name: "d", Height: 3},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			slices.SortStableFunc(tt.forks, forks.ForkSortFunc)
			assert.Equal(t, tt.expected, tt.forks)
		})
	}
}

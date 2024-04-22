package forks

import (
	"cmp"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
)

const (
	// ForkHalt is kiss-of-death fork, after which no transactions are included
	// in blocks, and all transactions are evicted from mempool. There are no
	// other changes that are sensible for a traditional consensus update. The
	// only conceivable use for this is a network migration where there should
	// be no transactions after the genesis data snapshot is collected for the
	// new network, which is accomplished via the globally available IsHalt and
	// BeginsHalt methods.
	ForkHalt = "halt"
)

// Forks lists the recognized hardforks and their activation heights or times,
// depending on the fork. A zero value means always active for the chain. A nil
// pointer (unset in JSON) indicates it is never activated.
type Forks struct {
	// The named *uint64 fields are natively defined forks recognized by kwild
	// code without any extensions.

	// HaltHeight is the height at which "halt" activates. This stops new transactions.
	HaltHeight *uint64

	// TODO (maybe): support activation time, which might be epoch milliseconds,
	// compared against time stamp of last block.

	// Extended contains any forks that may be defined by extensions.
	// Potentially we could embed a struct with extension forks. To do that the
	// extension package would not be able to import common.
	Extended map[string]*uint64
}

// NewForks creates a new Forks instance from a fork heights map.
func NewForks(forks map[string]*uint64) *Forks {
	var fs Forks
	fs.FromMap(forks)
	return &fs
}

func ptrHeight(ph *uint64) string {
	if ph == nil {
		return "<nil> (disabled)"
	}
	if *ph == 0 {
		return "0 (genesis)"
	}
	return strconv.FormatUint(*ph, 10)
}

// FromMap populates the Forks fields from a fork heights map. Use if a
// pre-allocated Forks instance or to merge multiple definitions.
func (fs *Forks) FromMap(forks map[string]*uint64) {
	extended := maps.Clone(forks)
	if ah, have := extended[ForkHalt]; have {
		fs.HaltHeight = ah
		delete(extended, ForkHalt)
	}

	fs.Extended = extended
}

// ActivatedBy returns a list of Forks names that are active by a certain block
// height.
func (fs *Forks) ActivatedBy(height uint64) []*Fork {
	names, heights := fs.matchForks(func(activationHeight uint64) bool {
		return height >= activationHeight
	})
	forks := make([]*Fork, len(names))
	for i := range names {
		forks[i] = &Fork{Name: names[i], Height: heights[i]}
	}
	return forks
}

// ActivatesAt returns a list of the fork names that activate at a certain
// height. Use ActivatesBy to return forks that activated at or before. The
// order of the slices is not defined and should not be relied upon.
func (fs *Forks) ActivatesAt(height uint64) []string {
	names, _ := fs.matchForks(func(activationHeight uint64) bool {
		return height == activationHeight
	})
	return names
}

// Fork identifies a hardfork and its activation height. The semantics and
// details of the changes associated with the fork are determined by code that
// is affected by the change, which should check if it is activated by height.
type Fork struct {
	Name   string
	Height uint64
}

// String describes the fork in the format "name:height".
func (f Fork) String() string {
	return fmt.Sprintf("%v@%d", f.Name, f.Height)
}

// ForkSortFunc is used with slices.SortFunc or slices.SortStableFunc to sort a
// []*Fork in ascending order by height.
func ForkSortFunc(a, b *Fork) int {
	return int(a.Height) - int(b.Height)
}

// matchForks is the primary helper method for returning for names and
// activation heights according to an arbitrary height comparison function that
// receives a fork's activation height.
func (fs *Forks) matchForks(cmp func(uint64) bool) ([]string, []uint64) {
	var forks []string
	var heights []uint64
	if fs.matchHalt(cmp) {
		forks = append(forks, ForkHalt)
		heights = append(heights, *fs.HaltHeight)
	}
	for fork, ah := range fs.Extended { // note: map range, order undefined
		if ah != nil && cmp(*ah) {
			forks = append(forks, fork)
			heights = append(heights, *ah)
		}
	}
	return forks, heights
}

func (fs *Forks) matchHalt(test func(ah uint64) bool) bool {
	if fh := fs.HaltHeight; fh != nil {
		return test(*fh)
	}
	return false
}

// IsHalt returns true if the "halt" rule changes are in effect *as of* the
// given height.
func (fs *Forks) IsHalt(height uint64) bool {
	return fs.matchHalt(func(fh uint64) bool {
		return height >= fh
	})
}

// BeginsHalt returns true if the "halt" rule changes go into effect *at* the
// given height.
func (fs *Forks) BeginsHalt(height uint64) bool {
	return fs.matchHalt(func(fh uint64) bool {
		return height == fh
	})
}

// ForkHeight returns the activation height of a fork by name, or nil if it
// never activates.
func (fs *Forks) ForkHeight(fork string) *uint64 {
	switch fork {
	case ForkHalt:
		return fs.HaltHeight
	}
	return fs.Extended[fork]
}

// String displays a human readable summary of all defined forks and their
// activation heights. For example:
//
// - halt: <nil> (disabled)
// - atGenesis: 0 (genesis)
// - alpha: 1
// - extended: 6
func (fs Forks) String() string {
	// Extended forks ordered by activation height
	type fk struct {
		name string
		ht   int64
		ph   *uint64
	}
	var fks []fk
	for name, ph := range fs.Extended {
		ht := int64(-1)
		if ph != nil {
			ht = int64(*ph)
		}
		fks = append(fks, fk{name, ht, ph})
	}
	slices.SortFunc(fks, func(a, b fk) int { return cmp.Compare(a.ht, b.ht) })
	var b strings.Builder
	for _, fk := range fks {
		fmt.Fprintf(&b, "\n- %s: %v", fk.name, ptrHeight(fk.ph))
	}
	// named first, in field order
	return fmt.Sprintf("- %v: %v", ForkHalt, ptrHeight(fs.HaltHeight)) +
		b.String()
}

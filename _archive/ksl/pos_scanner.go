package ksl

import (
	"bufio"
	"bytes"

	"github.com/apparentlymart/go-textseg/v13/textseg"
)

type RangeScanner struct {
	filename string
	b        []byte
	cb       bufio.SplitFunc

	pos Pos
	cur Range
	tok []byte
	err error
}

func NewRangeScanner(b []byte, filename string, cb bufio.SplitFunc) *RangeScanner {
	return NewRangeScannerFragment(b, filename, InitialPos, cb)
}

func NewRangeScannerFragment(b []byte, filename string, start Pos, cb bufio.SplitFunc) *RangeScanner {
	return &RangeScanner{
		filename: filename,
		b:        b,
		cb:       cb,
		pos:      start,
	}
}

func (sc *RangeScanner) Scan() bool {
	if sc.pos.Offset >= len(sc.b) || sc.err != nil {
		return false
	}

	advance, token, err := sc.cb(sc.b[sc.pos.Offset:], true)

	if advance == 0 && token == nil && err == nil {
		return false
	}

	if err != nil {
		sc.err = err
		sc.cur = Range{
			Filename: sc.filename,
			Start:    sc.pos,
			End:      sc.pos,
		}
		sc.tok = nil
		return false
	}

	sc.tok = token
	start := sc.pos
	end := sc.pos
	new := sc.pos

	adv := sc.b[sc.pos.Offset : sc.pos.Offset+advance]

	advR := bytes.NewReader(adv)
	gsc := bufio.NewScanner(advR)
	advanced := 0
	gsc.Split(textseg.ScanGraphemeClusters)
	for gsc.Scan() {
		gr := gsc.Bytes()
		new.Offset += len(gr)
		new.Column++

		if len(gr) != 0 && (gr[0] == '\r' || gr[0] == '\n') {
			new.Column = 1
			new.Line++
		}

		if advanced < len(token) {
			end = new
		}
		advanced += len(gr)
	}

	sc.cur = Range{
		Filename: sc.filename,
		Start:    start,
		End:      end,
	}
	sc.pos = new
	return true
}

func (sc *RangeScanner) Range() Range {
	return sc.cur
}

func (sc *RangeScanner) Bytes() []byte {
	return sc.tok
}

func (sc *RangeScanner) Err() error {
	return sc.err
}

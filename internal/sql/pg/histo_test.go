package pg

import (
	"bytes"
	"cmp"
	"math"
	"math/big"
	"math/rand"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/kwilteam/kwil-db/core/types"
)

func Test_histo_explicit_int(t *testing.T) {
	comp := cmp.Compare[int]
	interp := interpNum[int]
	bounds := makeBounds(10, int(-100), 800, comp, interp)
	hInt := makeHisto(bounds, comp)
	for i := 0; i < 1000; i++ {
		hInt.ins(rand.Intn(1200) - 200)
	}

	t.Log(hInt)
}

func Test_bounds_string(t *testing.T) {
	comp := strings.Compare
	interp := interpString
	bounds := makeBounds(100, "title_1", "title_99", comp, interp)
	// Think this through more...
	t.Log(bounds)
}

func Test_histo_num(t *testing.T) {
	bounds := makeBoundsNum[int](10, -100, 800)
	hInt := makeHisto(bounds, cmp.Compare[int])
	for i := 0; i < 1000; i++ {
		hInt.ins(rand.Intn(1200) - 200)
	}

	t.Log(hInt)
}

func Test_histo_float(t *testing.T) {
	bounds := makeBoundsNum[float64](10, -100, 800)
	hInt := makeHisto(bounds, cmp.Compare[float64])
	for i := 0; i < 1000; i++ {
		hInt.ins(float64(rand.Intn(1200) - 200))
	}
	// Without big.Rat impl:
	// bounds = [-100 -10 80 170.00000000000003 260 350 440.00000000000006 530 620 710 800]

	t.Log(hInt)
}

func Test_interpNumF(t *testing.T) {
	tests := []struct {
		name     string
		f        float64
		a        int
		b        int
		expected float64
	}{
		{"Zero interpolation", 0, 10, 20, 10},
		{"Full interpolation", 1, 10, 20, 20},
		{"Mid interpolation", 0.5, 10, 20, 15},
		{"Quarter interpolation", 0.25, 10, 20, 12.5},
		{"Three-quarter interpolation", 0.75, 10, 20, 17.5},
		{"Negative numbers", 0.5, -10, 10, 0},
		{"Same numbers", 0.5, 5, 5, 5},
		{"Large numbers", 0.5, 1000000, 2000000, 1500000},
		{"Small fractional numbers", 0.1, 1, 2, 1.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := interpNumF(tt.f, tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("interpNumF(%v, %v, %v) = %v, want %v", tt.f, tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func Test_interpNumF_float64(t *testing.T) {
	tests := []struct {
		name     string
		f        float64
		a        float64
		b        float64
		expected float64
	}{
		{"Fractional interpolation", 0.3, 1.5, 3.5, 2.1},
		{"Negative fractional interpolation", 0.7, -2.5, -1.5, -1.8},
		{"Zero to one interpolation", 0.5, 0, 1, 0.5},
		{"Very small numbers", 0.5, 1e-10, 2e-10, 1.5e-10},
		{"Very large numbers", 0.5, 1e10, 2e10, 1.5e10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := interpNumF(tt.f, tt.a, tt.b)
			if !almostEqual(result, tt.expected, 1e-9) {
				t.Errorf("interpNumF(%v, %v, %v) = %v, want %v", tt.f, tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}

func Test_interpNum(t *testing.T) {
	tests := []struct {
		name     string
		f        float64
		a        int
		b        int
		expected int
	}{
		{"Extreme interpolation near 1", 0.99, 0, 100, 99},
		{"Extreme interpolation near 0", 0.01, 0, 100, 1},
		{"Interpolation with negative and positive", 0.6, -50, 50, 10},
		{"Interpolation with both negative", 0.4, -100, -50, -80},
		{"Interpolation with large numbers", 0.75, 1000000, 2000000, 1750000},
		{"Zero interpolation with large difference", 0, -1000000, 1000000, -1000000},
		{"Full interpolation with large difference", 1, -1000000, 1000000, 1000000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := interpNum(tt.f, tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("interpNum(%v, %v, %v) = %v, want %v", tt.f, tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func Test_interpNum_float64(t *testing.T) {
	tests := []struct {
		name     string
		f        float64
		a        float64
		b        float64
		expected float64
	}{
		{"Interpolation with small fractional difference", 0.3, 1.1, 1.2, 1.13},
		{"Interpolation with large fractional difference", 0.7, 0.001, 0.1, 0.0703},
		{"Interpolation with negative fractionals", 0.4, -0.5, -0.1, -0.34},
		{"Extreme values near float64 limits", 0.5, -math.MaxFloat64 / 2, math.MaxFloat64 / 2, 0},
		{"Very small positive numbers", 0.6, 1e-15, 1e-14, 6.4e-15},
		{"Very small negative numbers", 0.8, -1e-14, -1e-15, -2.8e-15},
		{"Numbers close to each other", 0.5, 1.00000001, 1.00000002, 1.000000015},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := interpNum(tt.f, tt.a, tt.b)
			if !almostEqual(result, tt.expected, 1e-9) {
				t.Errorf("interpNum(%v, %v, %v) = %v, want %v", tt.f, tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func Test_makeBounds_int(t *testing.T) {
	tests := []struct {
		name     string
		n        int
		a        int
		b        int
		expected []int
	}{
		{
			name:     "Five equal intervals",
			n:        5,
			a:        0,
			b:        100,
			expected: []int{0, 20, 40, 60, 80, 100},
		},
		{
			name:     "two intervals (3 bounds) with negative num",
			n:        2,
			a:        -10,
			b:        10,
			expected: []int{-10, 0, 10},
		},
		{
			name:     "single interval",
			n:        1,
			a:        5,
			b:        10,
			expected: []int{5, 10},
		},
		{
			name:     "too small integer range",
			n:        10,
			a:        0,
			b:        4,
			expected: []int{0, 1, 2, 3, 4}, // integer forces fewer bounds
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeBoundsNum(tt.n, tt.a, tt.b)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("makeBounds(%v, %v, %v, interpNum[int]) = %v, want %v", tt.n, tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func Test_makeBounds_float64(t *testing.T) {
	tests := []struct {
		name     string
		n        int
		a        float64
		b        float64
		expected []float64
	}{
		{
			name:     "Four intervals with fractional values",
			n:        4,
			a:        0.5,
			b:        2.5,
			expected: []float64{0.5, 1.0, 1.5, 2.0, 2.5},
		},
		{
			name:     "Three intervals with very small numbers",
			n:        3,
			a:        1e-6,
			b:        1e-5,
			expected: []float64{1e-6, 4e-6, 7e-6, 1e-5},
		},
		{
			name:     "four intervals with negative fractional values",
			n:        4,
			a:        -1.5,
			b:        1.5,
			expected: []float64{-1.5, -0.75, 0, 0.75, 1.5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeBoundsNum(tt.n, tt.a, tt.b)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("makeBounds(%v, %v, %v, interpNum[float64]) = %v, want %v", tt.n, tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func Test_interpNum_edge_cases(t *testing.T) {
	tests := []struct {
		name     string
		f        float64
		a        int
		b        int
		expected int
	}{
		{"Interpolation with f = 0", 0, 10, 20, 10},
		{"Interpolation with f = 1", 1, 10, 20, 20},
		{"Interpolation with f > 1", 1.5, 10, 20, 25},
		{"Interpolation with f < 0", -0.5, 10, 20, 5},
		{"Interpolation with a = b", 0.5, 15, 15, 15},
		{"Interpolation with very large numbers", 0.5, math.MaxInt32 / 2, math.MaxInt32, 1610612735},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := interpNum(tt.f, tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("interpNum(%v, %v, %v) = %v, want %v", tt.f, tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func Test_interpBig(t *testing.T) {
	tests := []struct {
		name     string
		f        float64
		a        *big.Int
		b        *big.Int
		expected *big.Int
	}{
		{
			name:     "Simple interpolation",
			f:        0.5,
			a:        big.NewInt(100),
			b:        big.NewInt(200),
			expected: big.NewInt(150),
		},
		{
			name:     "Zero interpolation",
			f:        0,
			a:        big.NewInt(1000),
			b:        big.NewInt(2000),
			expected: big.NewInt(1000),
		},
		{
			name:     "Full interpolation",
			f:        1,
			a:        big.NewInt(500),
			b:        big.NewInt(1500),
			expected: big.NewInt(1500),
		},
		{
			name:     "Interpolation with negative numbers",
			f:        0.25,
			a:        big.NewInt(-1000),
			b:        big.NewInt(1000),
			expected: big.NewInt(-500),
		},
		{
			name: "Interpolation with large numbers",
			f:    0.5,
			a:    new(big.Int).Exp(big.NewInt(2), big.NewInt(100), nil),
			b:    new(big.Int).Exp(big.NewInt(2), big.NewInt(101), nil),
			expected: new(big.Int).Add(
				new(big.Int).Exp(big.NewInt(2), big.NewInt(100), nil),
				new(big.Int).Div(new(big.Int).Exp(big.NewInt(2), big.NewInt(100), nil), big.NewInt(2)),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := interpBig(tt.f, tt.a, tt.b)
			if result.Cmp(tt.expected) != 0 {
				t.Errorf("interpBig(%v, %v, %v) = %v, want %v", tt.f, tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func Test_interpBig_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()

	a := big.NewInt(100)
	b := big.NewInt(50)
	interpBig(0.5, a, b)
}

func Test_interpStr(t *testing.T) {
	tests := []struct {
		name     string
		f        float64
		a        string
		b        string
		expected string
	}{
		{
			name:     "simple",
			f:        0.5,
			a:        "abc",
			b:        "xyz",
			expected: "mno",
		},
		{
			name:     "zero f",
			f:        0,
			a:        "hello",
			b:        "world",
			expected: "hello",
		},
		{
			name:     "very simple",
			f:        0.5,
			a:        "a",
			b:        "c",
			expected: "b",
		},
		{
			name:     "full interp",
			f:        1,
			a:        "a",
			b:        "c",
			expected: "c",
		},
		{
			name:     "Interpolation with empty strings",
			f:        0.5,
			a:        "",
			b:        "test",
			expected: ":3::",
		},
		{
			name:     "Interpolation with unicode characters",
			f:        0.5,
			a:        "αβγ",
			b:        "δεζ",
			expected: "γδε",
		},
		// {
		// 	name:     "Interpolation with f > 1",
		// 	f:        1.5,
		// 	a:        "abc",
		// 	b:        "xyz",
		// 	expected: "",
		// },
		// {
		// 	name:     "Interpolation with f < 0",
		// 	f:        -0.5,
		// 	a:        "abc",
		// 	b:        "xyz",
		// 	expected: "UVW",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := interpString(tt.f, tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("interpStr(%v, %q, %q) = %q, want %q", tt.f, tt.a, tt.b, result, tt.expected)
			}

			// t.Log(result)

			if strings.Compare(tt.a, result) == 1 {
				t.Errorf("%v not <= %v", tt.a, result)
			}
			if strings.Compare(result, tt.b) == 1 {
				t.Errorf("%v not >) %v", tt.b, result)
			}
		})
	}
}

func Test_interpBts(t *testing.T) {
	tests := []struct {
		name     string
		f        float64
		a        []byte
		b        []byte
		expected []byte
	}{
		{
			name:     "Simple interpolation",
			f:        0.5,
			a:        []byte{0x00},
			b:        []byte{0xFF},
			expected: []byte{0x7F},
		},
		{
			name:     "Zero interpolation",
			f:        0,
			a:        []byte{0x10, 0x20},
			b:        []byte{0x30, 0x40},
			expected: []byte{0x10, 0x20},
		},
		{
			name:     "Full interpolation",
			f:        1,
			a:        []byte{0x00, 0x00},
			b:        []byte{0xFF, 0xFF},
			expected: []byte{0xFF, 0xFF},
		},
		{
			name:     "Interpolation with different byte lengths",
			f:        0.25,
			a:        []byte{0x01},
			b:        []byte{0x01, 0x00},
			expected: []byte{0x40},
		},
		{
			name:     "Interpolation with large numbers",
			f:        0.75,
			a:        []byte{0x00, 0x00, 0x00, 0x00},
			b:        []byte{0xFF, 0xFF, 0xFF, 0xFF},
			expected: []byte{0xBF, 0xFF, 0xFF, 0xFF},
		},
		{
			name:     "Interpolation with empty byte slice",
			f:        0.5,
			a:        []byte{},
			b:        []byte{0x2},
			expected: []byte{0x1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := interpBts(tt.f, tt.a, tt.b)
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("interpBts(%v, %v, %v) = %v, want %v", tt.f, tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

func Test_interpBts_EdgeCases(t *testing.T) {
	tests := []struct {
		name string
		f    float64
		a    []byte
		b    []byte
		want []byte
	}{
		{
			name: "Interpolation with f > 1",
			f:    1.5,
			a:    []byte{0x00},
			b:    []byte{0xFF},
			want: []byte{1, 126},
		},
		{
			name: "Interpolation with f < 0",
			f:    -0.5,
			a:    []byte{0x00},
			b:    []byte{0xFF},
			want: []byte{127},
		},
		{
			name: "Interpolation with very large byte slices",
			f:    0.5,
			a:    make([]byte, 1000),
			b:    bytes.Repeat([]byte{0xFF}, 1000),
			want: append([]byte{127}, bytes.Repeat([]byte{0xFF}, 999)...),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := interpBts(tt.f, tt.a, tt.b)
			if len(result) == 0 {
				t.Errorf("interpBts(%v, %v, %v) returned empty slice", tt.f, tt.a, tt.b)
			}
			if bytes.Compare(result, tt.a) < 0 || bytes.Compare(result, tt.b) > 0 {
				t.Errorf("interpBts(%v, %v, %v) = %v, which is out of range", tt.f, tt.a, tt.b, result)
			}
			// t.Log(result)
			if !bytes.Equal(result, tt.want) {
				t.Errorf("wanted %x, got %x", tt.want, result)
			}
		})
	}
}

func TestHistoMarshalUnmarshalBinary(t *testing.T) {
	tests := []struct {
		name  string
		histo histo[int64]
	}{
		{
			name: "Empty histogram",
			histo: histo[int64]{
				bounds: []int64{},
				freqs:  []int{},
				comp:   cmp.Compare[int64],
			},
		},
		{
			name: "Histogram with single bound",
			histo: histo[int64]{
				bounds: []int64{10},
				freqs:  []int{5},
				comp:   cmp.Compare[int64],
			},
		},
		{
			name: "Histogram with multiple bounds",
			histo: histo[int64]{
				bounds: []int64{0, 10, 20, 30},
				freqs:  []int{2, 5, 3, 1},
				comp:   cmp.Compare[int64],
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.histo.MarshalBinary()
			if err != nil {
				t.Fatalf("MarshalBinary() error = %v", err)
			}

			var unmarshaled histo[int64]
			err = unmarshaled.UnmarshalBinary(data)
			if err != nil {
				t.Fatalf("UnmarshalBinary() error = %v", err)
			}

			if !reflect.DeepEqual(tt.histo.bounds, unmarshaled.bounds) {
				t.Errorf("Bounds mismatch: got %v, want %v", unmarshaled.bounds, tt.histo.bounds)
			}

			if !reflect.DeepEqual(tt.histo.freqs, unmarshaled.freqs) {
				t.Errorf("Freqs mismatch: got %v, want %v", unmarshaled.freqs, tt.histo.freqs)
			}

			// also verify via the Equal method
			require.True(t, unmarshaled.Equal(tt.histo))

			// The comp function should be restored.
			require.NotNil(t, unmarshaled.comp)
		})
	}
}

func TestHistoMarshalUnmarshalBinaryWithLargeData(t *testing.T) {
	const size = 100000
	h := histo[int64]{
		bounds: make([]int64, size),
		freqs:  make([]int, size),
		comp:   cmp.Compare[int64],
	}

	for i := 0; i < size; i++ {
		h.bounds[i] = int64(i)
		h.freqs[i] = i * 2
	}

	data, err := h.MarshalBinary()
	if err != nil {
		t.Fatalf("MarshalBinary() error = %v", err)
	}

	var unmarshaled histo[int64]
	err = unmarshaled.UnmarshalBinary(data)
	if err != nil {
		t.Fatalf("UnmarshalBinary() error = %v", err)
	}

	if !reflect.DeepEqual(h.bounds, unmarshaled.bounds) {
		t.Errorf("Bounds mismatch for large data")
	}

	if !reflect.DeepEqual(h.freqs, unmarshaled.freqs) {
		t.Errorf("Freqs mismatch for large data")
	}
}

func TestHistoUnmarshalBinaryWithInvalidData(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "Empty data",
			data: []byte{},
		},
		{
			name: "Invalid gob data",
			data: []byte{0x01, 0x02, 0x03},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var h histo[int]
			err := h.UnmarshalBinary(tt.data)
			if err == nil {
				t.Errorf("UnmarshalBinary() expected error for invalid data, got nil")
			}
		})
	}
}

func Test_restoreHistoFuncs(t *testing.T) {
	h := histo[int64]{}
	restoreHistoFuncs(&h)
	t.Logf("%v", h.comp == nil)
}

func TestRestoreHistoFuncs(t *testing.T) {
	cs := sql.ColumnStatistics{
		Histogram: histo[int64]{},
	}
	ht := cs.Histogram.(histo[int64])
	assert.Nil(t, ht.comp)

	restoreStatsHistoFuncs(&cs)
	ht = cs.Histogram.(histo[int64])
	require.NotNil(t, ht.comp)
	assert.Equal(t, ht.comp(1, 2), -1)
}

func Test_restoreHisto(t *testing.T) {
	cs := sql.ColumnStatistics{
		Histogram: histo[int64]{},
	}

	cs.Histogram = setHistoCmpFunc(cs.Histogram)

	ht := cs.Histogram.(histo[int64])
	require.NotNil(t, ht.comp)
	assert.Equal(t, ht.comp(1, 2), -1)
}

func TestInterpUUID(t *testing.T) {
	tests := []struct {
		name     string
		f        float64
		a        types.UUID
		b        types.UUID
		expected types.UUID
	}{
		{
			name:     "result with leading zeros",
			f:        0.5,
			a:        types.UUID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			b:        types.UUID{0x00, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			expected: types.UUID{0x00, 0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		},
		{
			name:     "Zero interpolation",
			f:        0,
			a:        types.UUID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			b:        types.UUID{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			expected: types.UUID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		{
			name:     "Full interpolation",
			f:        1,
			a:        types.UUID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			b:        types.UUID{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			expected: types.UUID{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		},
		{
			name:     "Mid interpolation",
			f:        0.5,
			a:        types.UUID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			b:        types.UUID{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			expected: types.UUID{0x7F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		},
		{
			name:     "Quarter interpolation",
			f:        0.25,
			a:        types.UUID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			b:        types.UUID{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
			expected: types.UUID{0x3F, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF},
		},
		{
			name:     "Interpolation with same UUIDs",
			f:        0.5,
			a:        types.UUID{0x12, 0x34, 0x56, 0x78, 0x90, 0xAB, 0xCD, 0xEF, 0x12, 0x34, 0x56, 0x78, 0x90, 0xAB, 0xCD, 0xEF},
			b:        types.UUID{0x12, 0x34, 0x56, 0x78, 0x90, 0xAB, 0xCD, 0xEF, 0x12, 0x34, 0x56, 0x78, 0x90, 0xAB, 0xCD, 0xEF},
			expected: types.UUID{0x12, 0x34, 0x56, 0x78, 0x90, 0xAB, 0xCD, 0xEF, 0x12, 0x34, 0x56, 0x78, 0x90, 0xAB, 0xCD, 0xEF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := interpUUID(tt.f, tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("interpUUID(%v, %v, %v) = %v, want %v", tt.f, tt.a, tt.b, result, tt.expected)
			}
		})
	}
}

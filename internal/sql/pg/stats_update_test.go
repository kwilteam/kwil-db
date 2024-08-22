package pg

import (
	"math"
	"slices"
	"testing"

	"github.com/kwilteam/kwil-db/common/sql"
	"github.com/stretchr/testify/require"
)

func makeTestVals(num, samplesPerPeriod, amplSteps int, ampl, amplInc float64) []int64 {
	vals := make([]int64, num)
	for i := 0; i < num; i++ {
		p, f := math.Modf(float64(i) / float64(samplesPerPeriod))
		f *= 2 * math.Pi
		pMod := math.Mod(p, float64(amplSteps))
		p = amplInc * pMod // small periodic variation (0,1,2) in amplitude
		vals[i] = int64(math.Round((ampl + p) * math.Sin(f)))
	}
	return vals
}

// Test_updates_demo tests manual updates to a ColumnStatistics with the generic
// up* functions, which is how statistics are kept updated in the logical
// replication stream. Test_scanSineBig tests a full scan with TableStats.
func Test_updates_demo(t *testing.T) {
	stats := &sql.ColumnStatistics{}

	// Make data to ingest that:
	//  - uses the capacity of the MCVs
	//  - has many repeats
	//  - >90% (but not all) of values in MCVs
	// sine wave with 100 samples per periods, 100 periods
	const numUpdates = 10000
	const samplesPerPeriod = 100
	const ampl = 600.0  // larger => more integer discretization
	const amplSteps = 3 // "noise" with small ampl variations between periods
	const amplInc = 2.2 // each step adds a multiple of this to the amplitude

	// Build the full set of values
	vals := makeTestVals(numUpdates, samplesPerPeriod, amplSteps, ampl, amplInc)

	// ensure the test data exceeds the MCV cap
	fullCounts := make(map[int64]int)
	for _, v := range vals {
		fullCounts[v]++
	}
	require.Greater(t, len(fullCounts), statsCap)

	// insert one at a time
	for _, v := range vals {
		require.NoError(t, upColStatsWithInsert(stats, v))
	}

	maxVal := slices.Max(vals)

	// min/max must be captured even if not in the MCVs
	require.Equal(t, -maxVal, stats.Min)
	require.Equal(t, maxVal, stats.Max)

	hist := stats.Histogram.(histo[int64]) // t.Log("histogram:", hist)
	histCount0 := hist.TotalCount()

	// insert an outlier
	require.NoError(t, upColStatsWithInsert(stats, int64(math.MaxInt64)))
	require.Equal(t, int64(math.MaxInt64), stats.Max) // 9223372036854775807
	require.Equal(t, histCount0+1, hist.TotalCount()) // one more in the histogram

	// The MCVals slice should be sorted (ascending).
	mcVals := convSliceAsserted[int64](stats.MCVals)
	require.Equal(t, len(mcVals), statsCap)
	require.True(t, slices.IsSorted(mcVals))
	// outlier must not be in MCVs, which is full
	require.False(t, slices.Contains(mcVals, int64(math.MaxInt64)))

	// remove the outlier
	require.NoError(t, upColStatsWithDelete(stats, int64(math.MaxInt64)))
	require.Equal(t, unknown{}, stats.Max)
	require.Equal(t, histCount0, hist.TotalCount()) // back to the original counts

	ltVal := int64(0) // WHERE v < 0, about half the values for a sine wave
	loc, found := slices.BinarySearch(mcVals, ltVal)

	require.True(t, found)
	require.Equal(t, ltVal, mcVals[loc])

	// for "WHERE v <= 0", use loc++ (if found)
	loc++

	sumMCFreqs := func() int {
		var mcvSum, negativeFreqTotal int
		freqs := make([]float64, len(stats.MCFreqs))
		for i, f := range stats.MCFreqs {
			freqs[i] = float64(f) / float64(numUpdates)
			mcvSum += f
			if i < loc {
				negativeFreqTotal += f
			}
		}

		// t.Log("total freq of all MCVs:", float64(mcvSum)/float64(numUpdates))
		// t.Log("sum of freqs where v<=0:", float64(negativeFreqTotal)/float64(numUpdates))
		return mcvSum
	}

	mcvSum := sumMCFreqs()
	totalCounted := mcvSum + hist.TotalCount()

	// every value must be counted in either the MCVs array or the histogram
	require.Equal(t, numUpdates, totalCounted)

	// remove vals one at a time
	for _, v := range vals {
		require.NoError(t, upColStatsWithDelete(stats, v))
	}

	mcvSum = sumMCFreqs()
	require.Equal(t, 0, mcvSum)
	require.Equal(t, 0, hist.TotalCount())
}

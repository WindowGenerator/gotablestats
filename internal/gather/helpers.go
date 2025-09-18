package gather

import (
	"math"
	"sort"

	"github.com/WindowGenerator/gotablestats/internal/stat"
)

// calculateAggregates computes statistical aggregates for numeric data
func calculateAggregates(values []float64) *stat.AggregateStats {
	if len(values) == 0 {
		return &stat.AggregateStats{}
	}

	// Sort values for percentile calculations
	sortedValues := make([]float64, len(values))
	copy(sortedValues, values)
	sort.Float64s(sortedValues)

	// Basic stats
	count := int64(len(values))
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(count)

	// Variance and standard deviation
	variance := 0.0
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	variance /= float64(count)
	stdDev := math.Sqrt(variance)

	// Percentiles
	percentiles := make(map[int]float64)
	percentilePoints := []int{25, 50, 75, 90, 95, 99}

	for _, p := range percentilePoints {
		percentiles[p] = calculatePercentile(sortedValues, p)
	}

	return &stat.AggregateStats{
		Count:       count,
		Sum:         sum,
		Mean:        mean,
		Median:      percentiles[50],
		StdDev:      stdDev,
		Variance:    variance,
		Percentiles: percentiles,
	}
}

func calculatePercentile(sortedValues []float64, percentile int) float64 {
	if len(sortedValues) == 0 {
		return 0
	}

	index := float64(percentile) / 100.0 * float64(len(sortedValues)-1)

	if index == float64(int(index)) {
		return sortedValues[int(index)]
	}

	lower := int(math.Floor(index))
	upper := int(math.Ceil(index))
	weight := index - float64(lower)

	return sortedValues[lower]*(1-weight) + sortedValues[upper]*weight
}

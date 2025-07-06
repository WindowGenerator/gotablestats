package stats

import (
	"fmt"
	"math"
	"sort"
)

// calculateAggregates computes statistical aggregates for numeric data
func calculateAggregates(values []float64) *AggregateStats {
	if len(values) == 0 {
		return &AggregateStats{}
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

	return &AggregateStats{
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

func PrintStats(stats *TableStats, format string) {
	fmt.Printf("=== %s File Statistics ===\n", format)
	fmt.Printf("Sampled Rows: %d\n", stats.RowCount)
	fmt.Printf("Estimated Total Rows: %d\n", stats.EstimatedRows)
	fmt.Printf("Columns: %d\n", stats.ColumnCount)
	//	fmt.Printf("Sampling Config: %d samples from %d positions\n",
	//		stats.SamplingConfig.SampleSize, stats.SamplingConfig.RandomPositions)
	fmt.Printf("Column Names: %v\n", stats.ColumnNames)

	fmt.Println("\nColumn Details:")
	for _, colName := range stats.ColumnNames {
		fmt.Printf("  %s:\n", colName)
		fmt.Printf("    Type: %s\n", stats.ColumnTypes[colName])
		fmt.Printf("    Null Count: %d (%.2f%%)\n",
			stats.NullCounts[colName], stats.NullPercentage[colName])
		fmt.Printf("    Min: %v\n", stats.MinValues[colName])
		fmt.Printf("    Max: %v\n", stats.MaxValues[colName])

		// Print aggregates for numeric columns
		if agg, exists := stats.Aggregates[colName]; exists {
			fmt.Printf("    Aggregates:\n")
			fmt.Printf("      Count: %d\n", agg.Count)
			fmt.Printf("      Sum: %.2f\n", agg.Sum)
			fmt.Printf("      Mean: %.2f\n", agg.Mean)
			fmt.Printf("      Median: %.2f\n", agg.Median)
			fmt.Printf("      Std Dev: %.2f\n", agg.StdDev)
			fmt.Printf("      Percentiles: 25th=%.2f, 75th=%.2f, 95th=%.2f, 99th=%.2f\n",
				agg.Percentiles[25], agg.Percentiles[75],
				agg.Percentiles[95], agg.Percentiles[99])
		}
	}

	if len(stats.SampleData) > 0 {
		fmt.Println("\nSample Data:")
		for i, row := range stats.SampleData {
			fmt.Printf("  Row %d: %v\n", i+1, row)
		}
	}
	fmt.Println()
}

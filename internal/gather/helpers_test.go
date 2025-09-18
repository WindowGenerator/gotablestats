package gather

import (
	"math"
	"testing"

	"github.com/WindowGenerator/gotablestats/internal/stat"
)

func TestCalculateAggregates(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		expected *stat.AggregateStats
	}{
		{
			name:   "empty slice",
			values: []float64{},
			expected: &stat.AggregateStats{
				Count:       0,
				Sum:         0,
				Mean:        0,
				Median:      0,
				StdDev:      0,
				Variance:    0,
				Percentiles: nil,
			},
		},
		{
			name:   "single value",
			values: []float64{5.0},
			expected: &stat.AggregateStats{
				Count:    1,
				Sum:      5.0,
				Mean:     5.0,
				Median:   5.0,
				StdDev:   0.0,
				Variance: 0.0,
				Percentiles: map[int]float64{
					25: 5.0,
					50: 5.0,
					75: 5.0,
					90: 5.0,
					95: 5.0,
					99: 5.0,
				},
			},
		},
		{
			name:   "basic case",
			values: []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			expected: &stat.AggregateStats{
				Count:    5,
				Sum:      15.0,
				Mean:     3.0,
				Median:   3.0,
				StdDev:   math.Sqrt(2.0), // variance = 2.0
				Variance: 2.0,
				Percentiles: map[int]float64{
					25: 2.0,
					50: 3.0,
					75: 4.0,
					90: 4.6,
					95: 4.8,
					99: 4.96,
				},
			},
		},
		{
			name:   "unsorted values",
			values: []float64{5.0, 1.0, 3.0, 2.0, 4.0},
			expected: &stat.AggregateStats{
				Count:    5,
				Sum:      15.0,
				Mean:     3.0,
				Median:   3.0,
				StdDev:   math.Sqrt(2.0),
				Variance: 2.0,
				Percentiles: map[int]float64{
					25: 2.0,
					50: 3.0,
					75: 4.0,
					90: 4.6,
					95: 4.8,
					99: 4.96,
				},
			},
		},
		{
			name:   "duplicate values",
			values: []float64{2.0, 2.0, 2.0, 2.0},
			expected: &stat.AggregateStats{
				Count:    4,
				Sum:      8.0,
				Mean:     2.0,
				Median:   2.0,
				StdDev:   0.0,
				Variance: 0.0,
				Percentiles: map[int]float64{
					25: 2.0,
					50: 2.0,
					75: 2.0,
					90: 2.0,
					95: 2.0,
					99: 2.0,
				},
			},
		},
		{
			name:   "negative values",
			values: []float64{-2.0, -1.0, 0.0, 1.0, 2.0},
			expected: &stat.AggregateStats{
				Count:    5,
				Sum:      0.0,
				Mean:     0.0,
				Median:   0.0,
				StdDev:   math.Sqrt(2.0),
				Variance: 2.0,
				Percentiles: map[int]float64{
					25: -1.0,
					50: 0.0,
					75: 1.0,
					90: 1.6,
					95: 1.8,
					99: 1.96,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateAggregates(tt.values)

			// Check basic fields
			if result.Count != tt.expected.Count {
				t.Errorf("Count = %d, want %d", result.Count, tt.expected.Count)
			}
			if !floatEqual(result.Sum, tt.expected.Sum) {
				t.Errorf("Sum = %f, want %f", result.Sum, tt.expected.Sum)
			}
			if !floatEqual(result.Mean, tt.expected.Mean) {
				t.Errorf("Mean = %f, want %f", result.Mean, tt.expected.Mean)
			}
			if !floatEqual(result.Median, tt.expected.Median) {
				t.Errorf("Median = %f, want %f", result.Median, tt.expected.Median)
			}
			if !floatEqual(result.StdDev, tt.expected.StdDev) {
				t.Errorf("StdDev = %f, want %f", result.StdDev, tt.expected.StdDev)
			}
			if !floatEqual(result.Variance, tt.expected.Variance) {
				t.Errorf("Variance = %f, want %f", result.Variance, tt.expected.Variance)
			}

			// Check percentiles for non-empty cases
			if len(tt.values) > 0 {
				if result.Percentiles == nil {
					t.Error("Percentiles should not be nil for non-empty values")
					return
				}
				for p, expectedVal := range tt.expected.Percentiles {
					if actualVal, exists := result.Percentiles[p]; !exists {
						t.Errorf("Percentile %d missing", p)
					} else if !floatEqual(actualVal, expectedVal) {
						t.Errorf("Percentile %d = %f, want %f", p, actualVal, expectedVal)
					}
				}
			}
		})
	}
}

func TestCalculatePercentile(t *testing.T) {
	tests := []struct {
		name       string
		values     []float64
		percentile int
		expected   float64
	}{
		{
			name:       "empty slice",
			values:     []float64{},
			percentile: 50,
			expected:   0.0,
		},
		{
			name:       "single value",
			values:     []float64{5.0},
			percentile: 50,
			expected:   5.0,
		},
		{
			name:       "median of odd count",
			values:     []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			percentile: 50,
			expected:   3.0,
		},
		{
			name:       "median of even count",
			values:     []float64{1.0, 2.0, 3.0, 4.0},
			percentile: 50,
			expected:   2.5,
		},
		{
			name:       "25th percentile",
			values:     []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			percentile: 25,
			expected:   2.0,
		},
		{
			name:       "75th percentile",
			values:     []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			percentile: 75,
			expected:   4.0,
		},
		{
			name:       "90th percentile",
			values:     []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			percentile: 90,
			expected:   4.6,
		},
		{
			name:       "0th percentile (minimum)",
			values:     []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			percentile: 0,
			expected:   1.0,
		},
		{
			name:       "100th percentile (maximum)",
			values:     []float64{1.0, 2.0, 3.0, 4.0, 5.0},
			percentile: 100,
			expected:   5.0,
		},
		{
			name:       "interpolation case",
			values:     []float64{10.0, 20.0, 30.0, 40.0},
			percentile: 33,
			expected:   19.9, // 33% of 3 indices = 0.99, interpolate between index 0 and 1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculatePercentile(tt.values, tt.percentile)
			if !floatEqual(result, tt.expected) {
				t.Errorf("calculatePercentile(%v, %d) = %f, want %f",
					tt.values, tt.percentile, result, tt.expected)
			}
		})
	}
}

// Helper function to compare floats with tolerance
func floatEqual(a, b float64) bool {
	const tolerance = 1e-9
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	return math.Abs(a-b) < tolerance
}

// Benchmark tests
func BenchmarkCalculateAggregates(b *testing.B) {
	values := make([]float64, 1000)
	for i := range values {
		values[i] = float64(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculateAggregates(values)
	}
}

func BenchmarkCalculatePercentile(b *testing.B) {
	values := make([]float64, 1000)
	for i := range values {
		values[i] = float64(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculatePercentile(values, 50)
	}
}

// Test edge cases
func TestCalculateAggregatesEdgeCases(t *testing.T) {
	// Test with very large numbers
	t.Run("large numbers", func(t *testing.T) {
		values := []float64{1e10, 2e10, 3e10}
		result := calculateAggregates(values)
		if result.Count != 3 {
			t.Errorf("Count = %d, want 3", result.Count)
		}
		if result.Sum != 6e10 {
			t.Errorf("Sum = %f, want %f", result.Sum, 6e10)
		}
	})

	// Test with very small numbers
	t.Run("small numbers", func(t *testing.T) {
		values := []float64{1e-10, 2e-10, 3e-10}
		result := calculateAggregates(values)
		if result.Count != 3 {
			t.Errorf("Count = %d, want 3", result.Count)
		}
		expected := 6e-10
		if !floatEqual(result.Sum, expected) {
			t.Errorf("Sum = %e, want %e", result.Sum, expected)
		}
	})

	// Test with infinity values
	t.Run("infinity values", func(t *testing.T) {
		values := []float64{math.Inf(1), 1.0, 2.0}
		result := calculateAggregates(values)
		if !math.IsInf(result.Sum, 1) {
			t.Errorf("Sum should be +Inf, got %f", result.Sum)
		}
		if !math.IsInf(result.Mean, 1) {
			t.Errorf("Mean should be +Inf, got %f", result.Mean)
		}
	})

	// Test with NaN values
	t.Run("NaN values", func(t *testing.T) {
		values := []float64{math.NaN(), 1.0, 2.0}
		result := calculateAggregates(values)
		if !math.IsNaN(result.Sum) {
			t.Errorf("Sum should be NaN, got %f", result.Sum)
		}
		if !math.IsNaN(result.Mean) {
			t.Errorf("Mean should be NaN, got %f", result.Mean)
		}
	})
}

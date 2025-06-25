package stats

import (
	"bytes"
	"io"
	"math"
	"os"
	"strings"
	"testing"
)

func TestCalculateAggregates(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		expected *AggregateStats
	}{
		{
			name:   "empty slice",
			values: []float64{},
			expected: &AggregateStats{
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
			expected: &AggregateStats{
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
			expected: &AggregateStats{
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
			expected: &AggregateStats{
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
			expected: &AggregateStats{
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
			expected: &AggregateStats{
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

func TestPrintStats(t *testing.T) {
	// Create a sample TableStats struct
	stats := &TableStats{
		RowCount:      1000,
		EstimatedRows: 5000,
		ColumnCount:   3,
		ColumnNames:   []string{"id", "name", "age"},
		ColumnTypes: map[string]string{
			"id":   "integer",
			"name": "string",
			"age":  "float",
		},
		NullCounts: map[string]int64{
			"id":   0,
			"name": 10,
			"age":  5,
		},
		NullPercentage: map[string]float64{
			"id":   0.0,
			"name": 1.0,
			"age":  0.5,
		},
		MinValues: map[string]interface{}{
			"id":   1,
			"name": "Alice",
			"age":  18.5,
		},
		MaxValues: map[string]interface{}{
			"id":   1000,
			"name": "Zoe",
			"age":  65.2,
		},
		Aggregates: map[string]*AggregateStats{
			"age": {
				Count:    995,
				Sum:      25000.0,
				Mean:     25.13,
				Median:   24.5,
				StdDev:   12.5,
				Variance: 156.25,
				Percentiles: map[int]float64{
					25: 20.0,
					50: 24.5,
					75: 30.0,
					90: 40.0,
					95: 45.0,
					99: 50.0,
				},
			},
		},
		SampleData: [][]string{
			{"1", "Alice", "25.5"},
			{"2", "Bob", "30.0"},
			{"3", "Charlie", "22.3"},
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call the function
	PrintStats(stats, "CSV")

	// Restore stdout and get output
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Test various parts of the output
	expectedStrings := []string{
		"=== CSV File Statistics ===",
		"Sampled Rows: 1000",
		"Estimated Total Rows: 5000",
		"Columns: 3",
		"Column Names: [id name age]",
		"id:",
		"Type: integer",
		"name:",
		"Type: string",
		"age:",
		"Type: float",
		"Null Count: 0 (0.00%)",
		"Null Count: 10 (1.00%)",
		"Null Count: 5 (0.50%)",
		"Min: 1",
		"Max: 1000",
		"Min: Alice",
		"Max: Zoe",
		"Min: 18.5",
		"Max: 65.2",
		"Aggregates:",
		"Count: 995",
		"Sum: 25000.00",
		"Mean: 25.13",
		"Median: 24.50",
		"Std Dev: 12.50",
		"Percentiles: 25th=20.00, 75th=30.00, 95th=45.00, 99th=50.00",
		"Sample Data:",
		"Row 1: [1 Alice 25.5]",
		"Row 2: [2 Bob 30.0]",
		"Row 3: [3 Charlie 22.3]",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Output should contain '%s', but it doesn't.\nFull output:\n%s", expected, output)
		}
	}
}

func TestPrintStatsWithoutAggregatesAndSampleData(t *testing.T) {
	// Test with minimal data
	stats := &TableStats{
		RowCount:      100,
		EstimatedRows: 100,
		ColumnCount:   1,
		ColumnNames:   []string{"name"},
		ColumnTypes: map[string]string{
			"name": "string",
		},
		NullCounts: map[string]int64{
			"name": 0,
		},
		NullPercentage: map[string]float64{
			"name": 0.0,
		},
		MinValues: map[string]interface{}{
			"name": "Alice",
		},
		MaxValues: map[string]interface{}{
			"name": "Zoe",
		},
		Aggregates: map[string]*AggregateStats{}, // Empty
		SampleData: [][]string{},                 // Empty
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	PrintStats(stats, "JSON")

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Should not contain aggregate or sample data sections
	if strings.Contains(output, "Aggregates:") {
		t.Error("Output should not contain 'Aggregates:' section when no aggregates exist")
	}
	if strings.Contains(output, "Sample Data:") {
		t.Error("Output should not contain 'Sample Data:' section when no sample data exists")
	}

	// Should contain basic info
	expectedStrings := []string{
		"=== JSON File Statistics ===",
		"Sampled Rows: 100",
		"name:",
		"Type: string",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Output should contain '%s'", expected)
		}
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

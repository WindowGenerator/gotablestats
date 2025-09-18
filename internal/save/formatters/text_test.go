package formatters

import (
	"strings"
	"testing"

	"github.com/WindowGenerator/gotablestats/internal/stat"
)

func TestTextFormatter(t *testing.T) {
	// Create a sample stats.TableStats struct
	stats := &stat.TableStats{
		RowCount:      1000,
		EstimatedRows: 5000,
		ColumnCount:   3,
		ColumnNames:   []string{"id", "name", "age"},
		ColumnsStats: map[string]*stat.ColumnStats{
			"id": &stat.ColumnStats{
				Type:           "integer",
				NullCount:      0,
				NullPercentage: 0.0,
				MinValue:       1,
				MaxValue:       1000,
			},
			"name": &stat.ColumnStats{
				Type:           "string",
				NullCount:      10,
				NullPercentage: 1.0,
				MinValue:       "Alice",
				MaxValue:       "Zoe",
			},
			"age": &stat.ColumnStats{
				Type:           "float",
				NullCount:      5,
				NullPercentage: 0.5,
				MinValue:       18.5,
				MaxValue:       65.2,
				Aggregate: &stat.AggregateStats{
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
		},
		SampleData: [][]string{
			{"1", "Alice", "25.5"},
			{"2", "Bob", "30.0"},
			{"3", "Charlie", "22.3"},
		},
	}

	// Call the function
	formatter := NewTextFormatter()
	output, _ := formatter.Format(stats)

	// Test various parts of the output
	expectedStrings := []string{
		"=== File Statistics ===",
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
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("Output should contain '%s', but it doesn't.\nFull output:\n%s", expected, output)
		}
	}
}

func TestTextFormatterWithoutAggregatesAndSampleData(t *testing.T) {
	// Test with minimal data
	stats := &stat.TableStats{
		RowCount:      100,
		EstimatedRows: 100,
		ColumnCount:   1,
		ColumnNames:   []string{"name"},
		ColumnsStats: map[string]*stat.ColumnStats{
			"name": &stat.ColumnStats{
				Type:           "string",
				NullCount:      0,
				NullPercentage: 0.0,
				MinValue:       "Alice",
				MaxValue:       "Zoe",
			},
		},
		SampleData: [][]string{}, // Empty
	}

	formatter := NewTextFormatter()
	output, _ := formatter.Format(stats)

	// Should not contain aggregate or sample data sections
	if strings.Contains(output, "Aggregates:") {
		t.Error("Output should not contain 'Aggregates:' section when no aggregates exist")
	}
	if strings.Contains(output, "Sample Data:") {
		t.Error("Output should not contain 'Sample Data:' section when no sample data exists")
	}

	// Should contain basic info
	expectedStrings := []string{
		"=== File Statistics ===",
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

package formatters

import (
	"fmt"

	"github.com/WindowGenerator/gotablestats/internal/stat"
)

type TextFormatter struct{}

func NewTextFormatter() *TextFormatter {
	return &TextFormatter{}
}

func (formatter *TextFormatter) Format(stats *stat.TableStats) (string, error) {
	text := "=== File Statistics ===\n"
	text += fmt.Sprintf("Sampled Rows: %d\n", stats.RowCount)
	text += fmt.Sprintf("Estimated Total Rows: %d\n", stats.EstimatedRows)
	text += fmt.Sprintf("Columns: %d\n", stats.ColumnCount)
	text += fmt.Sprintf("Column Names: %v\n", stats.ColumnNames)

	text += "\nColumn Details:"
	for colName, colStats := range stats.ColumnsStats {
		text += fmt.Sprintf("  %s:\n", colName)
		text += fmt.Sprintf("    Type: %s\n", colStats.Type)
		text += fmt.Sprintf("    Null Count: %d (%.2f%%)\n", colStats.NullCount, colStats.NullPercentage)
		text += fmt.Sprintf("    Min: %v\n", colStats.MinValue)
		text += fmt.Sprintf("    Max: %v\n", colStats.MaxValue)

		// Print aggregates for numeric columns
		if colStats.Aggregate != nil {
			agg := colStats.Aggregate

			text += "    Aggregates:\n"
			text += fmt.Sprintf("      Count: %d\n", agg.Count)
			text += fmt.Sprintf("      Sum: %.2f\n", agg.Sum)
			text += fmt.Sprintf("      Mean: %.2f\n", agg.Mean)
			text += fmt.Sprintf("      Median: %.2f\n", agg.Median)
			text += fmt.Sprintf("      Std Dev: %.2f\n", agg.StdDev)
			text += fmt.Sprintf("      Percentiles: 25th=%.2f, 75th=%.2f, 95th=%.2f, 99th=%.2f\n",
				agg.Percentiles[25], agg.Percentiles[75],
				agg.Percentiles[95], agg.Percentiles[99])
		}
	}

	return text, nil
}

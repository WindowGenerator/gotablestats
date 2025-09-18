package gather

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/WindowGenerator/gotablestats/internal/stat"
)

// Test helper functions

func createTempCSV(t *testing.T, content string, delimiter rune) *os.File {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.csv")

	file, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	writer := csv.NewWriter(file)
	writer.Comma = delimiter

	lines := splitLines(content)
	for _, line := range lines {
		if line != "" {
			fields := splitCSVLine(line, delimiter)
			writer.Write(fields)
		}
	}
	writer.Flush()
	file.Seek(0, io.SeekStart)

	return file
}

func splitLines(content string) []string {
	lines := []string{}
	current := ""
	for _, char := range content {
		if char == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func splitCSVLine(line string, delimiter rune) []string {
	fields := []string{}
	current := ""
	for _, char := range line {
		if char == delimiter {
			fields = append(fields, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	fields = append(fields, current)
	return fields
}

func createLargeCSV(t *testing.T, rows int) *os.File {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "large.csv")

	file, err := os.Create(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	writer := csv.NewWriter(file)

	// Write header
	writer.Write([]string{"id", "name", "value", "category"})

	// Write data rows
	for i := 1; i <= rows; i++ {
		writer.Write([]string{
			fmt.Sprintf("%d", i),
			fmt.Sprintf("name_%d", i),
			fmt.Sprintf("%.2f", float64(i)*1.5),
			fmt.Sprintf("cat_%d", i%5),
		})
	}
	writer.Flush()
	file.Seek(0, io.SeekStart)

	return file
}

// Tests for NewCSVStatsGatherer

func TestNewCSVStatsGatherer(t *testing.T) {
	reader := NewCSVStatsGatherer(',')

	if reader.Delimiter != ',' {
		t.Errorf("Expected delimiter ',', got %c", reader.Delimiter)
	}

	reader2 := NewCSVStatsGatherer(';')
	if reader2.Delimiter != ';' {
		t.Errorf("Expected delimiter ';', got %c", reader2.Delimiter)
	}
}

// Tests for ReadTable

func TestReadTable_BasicCSV(t *testing.T) {
	csvContent := `name,age,salary
John,25,50000
Jane,30,60000
Bob,35,55000`

	tmpFile := createTempCSV(t, csvContent, ',')
	defer tmpFile.Close()

	reader := NewCSVStatsGatherer(',')
	config := stat.SamplingConfig{
		MaxFileSize:     1024 * 1024, // 1MB
		SampleSize:      1000,
		RandomPositions: 5,
	}

	stats, err := reader.Stats(tmpFile, config)
	if err != nil {
		t.Fatalf("ReadTable failed: %v", err)
	}

	// Check basic stats
	if stats.ColumnCount != 3 {
		t.Errorf("Expected 3 columns, got %d", stats.ColumnCount)
	}

	if stats.RowCount != 3 {
		t.Errorf("Expected 3 rows, got %d", stats.RowCount)
	}

	expectedColumns := []string{"name", "age", "salary"}
	if !reflect.DeepEqual(stats.ColumnNames, expectedColumns) {
		t.Errorf("Expected columns %v, got %v", expectedColumns, stats.ColumnNames)
	}

	// Check column types
	if stats.ColumnsStats["name"].Type != "string" {
		t.Errorf("Expected name column to be string, got %s", stats.ColumnsStats["name"].Type)
	}

	if stats.ColumnsStats["age"].Type != "int64" {
		t.Errorf("Expected age column to be int64, got %s", stats.ColumnsStats["age"].Type)
	}

	if stats.ColumnsStats["salary"].Type != "int64" {
		t.Errorf("Expected salary column to be int64, got %s", stats.ColumnsStats["salary"].Type)
	}
}

func TestReadTable_WithFloats(t *testing.T) {
	csvContent := `name,height,weight
Alice,5.6,120.5
Bob,6.0,180.0
Charlie,5.8,165.25`

	tmpFile := createTempCSV(t, csvContent, ',')
	defer tmpFile.Close()

	reader := NewCSVStatsGatherer(',')
	config := stat.SamplingConfig{
		MaxFileSize:     1024 * 1024,
		SampleSize:      1000,
		RandomPositions: 5,
	}

	stats, err := reader.Stats(tmpFile, config)
	if err != nil {
		t.Fatalf("ReadTable failed: %v", err)
	}

	// Check float column types
	if stats.ColumnsStats["height"].Type != "float64" {
		t.Errorf("Expected height column to be float64, got %s", stats.ColumnsStats["height"].Type)
	}

	if stats.ColumnsStats["weight"].Type != "float64" {
		t.Errorf("Expected weight column to be float64, got %s", stats.ColumnsStats["weight"].Type)
	}

	// Check aggregates exist for numeric columns
	if stats.ColumnsStats["height"].Aggregate == nil {
		t.Error("Expected aggregates for height column")
	}

	if stats.ColumnsStats["weight"].Aggregate == nil {
		t.Error("Expected aggregates for weight column")
	}
}

func TestReadTable_WithDate(t *testing.T) {
	csvContent := `name,datetime,date
Alice,1937-01-01T12:00:27.87+00:20,1902-05-19
Bob,1990-12-31T15:59:00-08:00,2007-03-12
Charlie,1996-12-19T16:39:57-08:00,2011-11-02`

	tmpFile := createTempCSV(t, csvContent, ',')
	defer tmpFile.Close()

	reader := NewCSVStatsGatherer(',')
	config := stat.SamplingConfig{
		MaxFileSize:     1024 * 1024,
		SampleSize:      1000,
		RandomPositions: 5,
	}

	stats, err := reader.Stats(tmpFile, config)
	if err != nil {
		t.Fatalf("ReadTable failed: %v", err)
	}

	if stats.ColumnsStats["datetime"].Type != "datetime" {
		t.Errorf("Expected datetime column to be datetime, got %s", stats.ColumnsStats["datetime"].Type)
	}

	if stats.ColumnsStats["date"].Type != "date" {
		t.Errorf("Expected date column to be date, got %s", stats.ColumnsStats["date"].Type)
	}
}

func TestReadTable_WithBool(t *testing.T) {
	csvContent := `name,done
Alice,true
Bob,false
Charlie,true`

	tmpFile := createTempCSV(t, csvContent, ',')
	defer tmpFile.Close()

	reader := NewCSVStatsGatherer(',')
	config := stat.SamplingConfig{
		MaxFileSize:     1024 * 1024,
		SampleSize:      1000,
		RandomPositions: 5,
	}

	stats, err := reader.Stats(tmpFile, config)
	if err != nil {
		t.Fatalf("ReadTable failed: %v", err)
	}

	if stats.ColumnsStats["done"].Type != "bool" {
		t.Errorf("Expected done column to be bool, got %s", stats.ColumnsStats["done"].Type)
	}
}

func TestReadTable_WithNulls(t *testing.T) {
	csvContent := `name,age,city
John,25,NYC
Jane,,Boston
Bob,35,
Alice,null,Chicago`

	tmpFile := createTempCSV(t, csvContent, ',')
	defer tmpFile.Close()

	reader := NewCSVStatsGatherer(',')
	config := stat.SamplingConfig{
		MaxFileSize:     1024 * 1024,
		SampleSize:      1000,
		RandomPositions: 5,
		NullValues:      []string{"null", ""},
	}

	stats, err := reader.Stats(tmpFile, config)
	if err != nil {
		t.Fatalf("ReadTable failed: %v", err)
	}

	// Check null counts
	if stats.ColumnsStats["age"].NullCount != 2 { // Jane and Alice
		t.Errorf("Expected 2 nulls in age column, got %d", stats.ColumnsStats["age"].NullCount)
	}

	if stats.ColumnsStats["city"].NullCount != 1 { // Bob
		t.Errorf("Expected 1 null in city column, got %d", stats.ColumnsStats["city"].NullCount)
	}

	// Check null percentages
	expectedAgeNullPct := 50.0 // 2 out of 4 rows
	if stats.ColumnsStats["age"].NullPercentage != expectedAgeNullPct {
		t.Errorf("Expected %.1f%% nulls in age, got %.1f%%", expectedAgeNullPct, stats.ColumnsStats["age"].NullPercentage)
	}
}

func TestReadTable_CustomDelimiter(t *testing.T) {
	csvContent := `name;age;department
John;25;Engineering
Jane;30;Marketing`

	tmpFile := createTempCSV(t, csvContent, ';')
	defer tmpFile.Close()

	reader := NewCSVStatsGatherer(';')
	config := stat.SamplingConfig{
		MaxFileSize:     1024 * 1024,
		SampleSize:      1000,
		RandomPositions: 5,
	}

	stats, err := reader.Stats(tmpFile, config)
	if err != nil {
		t.Fatalf("ReadTable failed: %v", err)
	}

	if stats.ColumnCount != 3 {
		t.Errorf("Expected 3 columns, got %d", stats.ColumnCount)
	}

	expectedColumns := []string{"name", "age", "department"}
	if !reflect.DeepEqual(stats.ColumnNames, expectedColumns) {
		t.Errorf("Expected columns %v, got %v", expectedColumns, stats.ColumnNames)
	}
}

func TestReadTable_EmptyFile(t *testing.T) {
	tmpFile := createTempCSV(t, "", ',')
	defer tmpFile.Close()

	reader := NewCSVStatsGatherer(',')
	config := stat.SamplingConfig{
		MaxFileSize:     1024 * 1024,
		SampleSize:      1000,
		RandomPositions: 5,
	}

	_, err := reader.Stats(tmpFile, config)
	if err == nil {
		t.Error("Expected error for empty file")
	}
}

func TestReadTable_HeaderOnly(t *testing.T) {
	csvContent := `name,age,city`

	tmpFile := createTempCSV(t, csvContent, ',')

	reader := NewCSVStatsGatherer(',')
	config := stat.SamplingConfig{
		MaxFileSize:     1024 * 1024,
		SampleSize:      1000,
		RandomPositions: 5,
	}

	stats, err := reader.Stats(tmpFile, config)
	if err != nil {
		t.Fatalf("ReadTable failed: %v", err)
	}

	if stats.RowCount != 0 {
		t.Errorf("Expected 0 data rows, got %d", stats.RowCount)
	}

	if stats.ColumnCount != 3 {
		t.Errorf("Expected 3 columns, got %d", stats.ColumnCount)
	}
}

func TestReadTable_SampleData(t *testing.T) {
	csvContent := `name,value
A,1
B,2
C,3
D,4
E,5
F,6`

	tmpFile := createTempCSV(t, csvContent, ',')
	defer tmpFile.Close()

	reader := NewCSVStatsGatherer(',')
	config := stat.SamplingConfig{
		MaxFileSize:     1024 * 1024,
		SampleSize:      1000,
		RandomPositions: 5,
	}

	stats, err := reader.Stats(tmpFile, config)
	if err != nil {
		t.Fatalf("ReadTable failed: %v", err)
	}

	// Should have 5 sample rows (default sample size)
	if len(stats.SampleData) != 5 {
		t.Errorf("Expected 5 sample rows, got %d", len(stats.SampleData))
	}

	// First sample row should be the first data row
	if len(stats.SampleData[0]) != 2 {
		t.Errorf("Expected 2 columns in sample data, got %d", len(stats.SampleData[0]))
	}
}

// Tests for large file sampling

func TestReadTable_LargeFileSampling(t *testing.T) {
	// Create a file larger than MaxFileSize to trigger sampling
	tmpFile := createLargeCSV(t, 10000)
	defer tmpFile.Close()

	reader := NewCSVStatsGatherer(',')
	config := stat.SamplingConfig{
		MaxFileSize:     1000, // Very small to force sampling
		SampleSize:      100,
		RandomPositions: 5,
	}

	stats, err := reader.Stats(tmpFile, config)
	if err != nil {
		t.Fatalf("ReadTable failed: %v", err)
	}

	// Should have sampled data, not full dataset
	if stats.RowCount > int64(config.SampleSize) {
		t.Errorf("Expected at most %d sampled rows, got %d", config.SampleSize, stats.RowCount)
	}

	// Estimated rows should be much higher than actual sampled rows
	if stats.EstimatedRows <= stats.RowCount {
		t.Errorf("Expected estimated rows (%d) to be higher than sampled rows (%d)",
			stats.EstimatedRows, stats.RowCount)
	}
}

// Tests for column analysis

func TestAnalyzeColumn_MinMaxValues(t *testing.T) {
	csvContent := `name,age,score
Alice,25,85.5
Bob,30,92.0
Charlie,22,78.5`

	tmpFile := createTempCSV(t, csvContent, ',')
	defer tmpFile.Close()

	reader := NewCSVStatsGatherer(',')
	config := stat.SamplingConfig{
		MaxFileSize:     1024 * 1024,
		SampleSize:      1000,
		RandomPositions: 5,
	}

	stats, err := reader.Stats(tmpFile, config)
	if err != nil {
		t.Fatalf("ReadTable failed: %v", err)
	}

	// Check min/max for string column
	if stats.ColumnsStats["name"].MinValue != "Alice" {
		t.Errorf("Expected min name 'Alice', got %v", stats.ColumnsStats["name"].MinValue)
	}
	if stats.ColumnsStats["name"].MaxValue != "Charlie" {
		t.Errorf("Expected max name 'Charlie', got %v", stats.ColumnsStats["name"].MaxValue)
	}

	// Check min/max for numeric columns
	if stats.ColumnsStats["age"].MinValue != float64(22) {
		t.Errorf("Expected min age 22, got %v", stats.ColumnsStats["age"].MinValue)
	}
	if stats.ColumnsStats["age"].MaxValue != float64(30) {
		t.Errorf("Expected max age 30, got %v", stats.ColumnsStats["age"].MaxValue)
	}
}

func TestAnalyzeColumn_MixedTypes(t *testing.T) {
	// Column starts as numeric but has non-numeric values
	csvContent := `id,mixed_col,mixed_col2
1,123,33.4
2,456,442
3,abc,1990-12-31T15:59:00-08:00
4,1990-12-31T15:59:00-08:00,1990-12-31
5,123,true`

	tmpFile := createTempCSV(t, csvContent, ',')
	defer tmpFile.Close()

	reader := NewCSVStatsGatherer(',')
	config := stat.SamplingConfig{
		MaxFileSize:     1024 * 1024,
		SampleSize:      1000,
		RandomPositions: 5,
	}

	stats, err := reader.Stats(tmpFile, config)
	if err != nil {
		t.Fatalf("ReadTable failed: %v", err)
	}

	// Mixed column should be treated as string
	if stats.ColumnsStats["mixed_col"].Type != "string" {
		t.Errorf("Expected mixed_col to be string, got %s", stats.ColumnsStats["mixed_col"].Type)
	}

	if stats.ColumnsStats["mixed_col2"].Type != "string" {
		t.Errorf("Expected mixed_col2 to be string, got %s", stats.ColumnsStats["mixed_col2"].Type)
	}

	// Should not have aggregates for string columns
	if stats.ColumnsStats["mixed_col"].Aggregate != nil {
		t.Error("Expected no aggregates for string column")
	}

	if stats.ColumnsStats["mixed_col2"].Aggregate != nil {
		t.Error("Expected no aggregates for string column")
	}
}

// Test helper for sampling methods

func TestSampleRecords(t *testing.T) {
	// Create a reasonably sized file
	tmpFile := createLargeCSV(t, 1000)
	defer tmpFile.Close()

	fileInfo, _ := tmpFile.Stat()

	reader := NewCSVStatsGatherer(',')
	config := stat.SamplingConfig{
		SampleSize:      50,
		RandomPositions: 5,
	}

	records, _, err := reader.sampleRecords(tmpFile, fileInfo.Size(), config)
	if err != nil {
		t.Fatalf("sampleRecords failed: %v", err)
	}

	if len(records) > config.SampleSize {
		t.Errorf("Expected at most %d records, got %d", config.SampleSize, len(records))
	}

	// Each record should have 4 columns (from createLargeCSV)
	for i, record := range records {
		if len(record) != 4 {
			t.Errorf("Record %d has %d columns, expected 4", i, len(record))
		}
	}
}

func TestEstimateRowCount(t *testing.T) {
	reader := NewCSVStatsGatherer(',')
	config := stat.SamplingConfig{
		RandomPositions: 5,
		SampleSize:      100,
	}

	fileSize := int64(100000)
	var readerBytes int64 = 1000

	estimate := reader.estimateRowCount(fileSize, readerBytes, config)

	// Estimate should be reasonable (non-zero and not negative)
	if estimate <= 0 {
		t.Errorf("Expected positive row count estimate, got %d", estimate)
	}

	// Estimate should be proportional to file size
	estimate2 := reader.estimateRowCount(fileSize*2, readerBytes, config)
	if estimate2 < estimate {
		t.Errorf("Expected larger estimate for larger file, got %d >= %d", estimate2, estimate)
	}
}

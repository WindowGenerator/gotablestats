package stats

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

// CSVReader implements TableReader for CSV files with probabilistic sampling
type CSVReader struct {
	Delimiter rune
}

func NewCSVReader(delimiter rune) *CSVReader {
	return &CSVReader{
		Delimiter: delimiter,
	}
}

func (r *CSVReader) GetFormatName() string {
	return "CSV"
}

func (r *CSVReader) ReadTable(filePath string, config SamplingConfig) (*TableStats, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	fileSize := fileInfo.Size()

	// Read header first
	csvReader := csv.NewReader(file)
	csvReader.Comma = r.Delimiter

	header, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	stats := &TableStats{
		ColumnCount:    len(header),
		ColumnNames:    header,
		ColumnTypes:    make(map[string]string),
		NullCounts:     make(map[string]int64),
		NullPercentage: make(map[string]float64),
		MinValues:      make(map[string]interface{}),
		MaxValues:      make(map[string]interface{}),
		SampleData:     make([][]string, 0),
		Aggregates:     make(map[string]*AggregateStats),
		SamplingConfig: config,
	}

	var records [][]string
	var readerBytes int64

	// Decide sampling strategy based on file size
	if fileSize <= config.MaxFileSize {
		// Small file - read entirely
		allRecords, err := csvReader.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV: %w", err)
		}
		records = allRecords
		stats.RowCount = int64(len(records))
		stats.EstimatedRows = stats.RowCount
	} else {
		// Large file - use probabilistic sampling
		records, readerBytes, err = r.sampleRecords(file, fileSize, config)
		if err != nil {
			return nil, fmt.Errorf("failed to sample records: %w", err)
		}
		stats.RowCount = int64(len(records))
		// Estimate total rows based on sampling
		stats.EstimatedRows = r.estimateRowCount(fileSize, readerBytes, config)
	}

	if len(records) == 0 {
		return stats, nil
	}

	// Get sample data
	sampleSize := 5
	if len(records) < sampleSize {
		sampleSize = len(records)
	}
	stats.SampleData = records[:sampleSize]

	// Analyze each column
	for colIdx, colName := range stats.ColumnNames {
		r.analyzeColumn(records, colIdx, colName, stats)
	}

	return stats, nil
}

func (r *CSVReader) sampleRecords(file *os.File, fileSize int64, config SamplingConfig) ([][]string, int64, error) {
	var allRecords [][]string
	recordsPerPosition := config.SampleSize / config.RandomPositions
	if recordsPerPosition < 1 {
		recordsPerPosition = 1
	}

	var readerBytes int64 = 0

	for i := 0; i < config.RandomPositions; i++ {
		// Generate random position (skip first 1% to avoid header area)
		minPos := fileSize / 100
		randomPos := minPos + rand.Int63n(fileSize-minPos)

		_, err := file.Seek(randomPos, io.SeekStart)
		if err != nil {
			return nil, 0, err
		}

		records, err := r.readFromPosition(file, recordsPerPosition)
		if err != nil {
			continue // Skip failed positions
		}
		current, err := file.Seek(0, io.SeekCurrent)
		if err != nil {
			return nil, 0, err
		}

		readerBytes += current - randomPos
		allRecords = append(allRecords, records...)

		if len(allRecords) >= config.SampleSize {
			break
		}
	}

	// Trim to exact sample size
	if len(allRecords) > config.SampleSize {
		allRecords = allRecords[:config.SampleSize]
	}

	return allRecords, readerBytes, nil
}

func (r *CSVReader) readFromPosition(file *os.File, maxRecords int) ([][]string, error) {
	reader := bufio.NewReader(file)

	// Skip to next complete line (in case we're in the middle of a line)
	_, _, err := reader.ReadLine()
	if err != nil && err != io.EOF {
		return nil, err
	}

	// Read records from this position
	csvReader := csv.NewReader(reader)
	csvReader.Comma = r.Delimiter

	var records [][]string
	for i := 0; i < maxRecords; i++ {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue // Skip malformed records
		}
		records = append(records, record)
	}

	return records, nil
}

func (r *CSVReader) estimateRowCount(fileSize int64, readerBytes int64, config SamplingConfig) int64 {
	// Simple estimation based on file size and sample density
	avgBytesPerRecord := readerBytes / int64(config.SampleSize)
	estimatedRows := fileSize / avgBytesPerRecord
	return estimatedRows
}

func toStringComparable(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return fmt.Sprintf("%020.6f", val)
	default:
		panic("can't parse vinput value. Please contact with maintainerce")
	}
}

func (r *CSVReader) analyzeColumn(records [][]string, colIdx int, colName string, stats *TableStats) {
	var nullCount int64
	var minVal, maxVal interface{}
	var isNumeric bool = true
	var isFloat bool = false
	var numericValues []float64

	for _, record := range records {
		if colIdx >= len(record) {
			nullCount++
			continue
		}

		value := strings.TrimSpace(record[colIdx])
		if value == "" || value == "NULL" || value == "null" {
			nullCount++
			continue
		}

		// Try to determine type and collect numeric values
		if isNumeric {
			if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
				numericValues = append(numericValues, floatVal)
				if strings.Contains(value, ".") {
					isFloat = true
				}
				if minVal == nil || floatVal < minVal.(float64) {
					minVal = floatVal
				}
				if maxVal == nil || floatVal > maxVal.(float64) {
					maxVal = floatVal
				}
			} else {
				isNumeric = false
				isFloat = false
				// Switch to string comparison and clear numeric values
				numericValues = nil

				if minVal == nil || value < toStringComparable(minVal) {
					minVal = value
				}
				if maxVal == nil || value > toStringComparable(maxVal) {
					maxVal = value
				}
			}
		} else {
			// String comparison
			if minVal == nil || value < minVal.(string) {
				minVal = value
			}
			if maxVal == nil || value > maxVal.(string) {
				maxVal = value
			}
		}
	}

	// Set column type
	if isNumeric {
		if isFloat {
			stats.ColumnTypes[colName] = "float64"
		} else {
			stats.ColumnTypes[colName] = "int64"
		}

		// Calculate aggregates for numeric columns
		if len(numericValues) > 0 {
			stats.Aggregates[colName] = calculateAggregates(numericValues)
		}
	} else {
		stats.ColumnTypes[colName] = "string"
	}

	stats.NullCounts[colName] = nullCount
	stats.NullPercentage[colName] = float64(nullCount) / float64(len(records)) * 100
	stats.MinValues[colName] = minVal
	stats.MaxValues[colName] = maxVal
}

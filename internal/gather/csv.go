package gather

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"math/rand"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/WindowGenerator/gotablestats/internal/stat"
)

// CSVStatsGatherer implements TableReader for CSV files with probabilistic sampling
type CSVStatsGatherer struct {
	Delimiter rune
}

func NewCSVStatsGatherer(delimiter rune) *CSVStatsGatherer {
	return &CSVStatsGatherer{
		Delimiter: delimiter,
	}
}

func (r *CSVStatsGatherer) Stats(file *os.File, config stat.SamplingConfig) (*stat.TableStats, error) {
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

	stats := &stat.TableStats{
		ColumnCount:  len(header),
		ColumnNames:  header,
		ColumnsStats: make(map[string]*stat.ColumnStats),
		SampleData:   make([][]string, 0),
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
		stats.ColumnsStats[colName] = &stat.ColumnStats{}
		r.analyzeColumn(records, colIdx, colName, stats.ColumnsStats[colName], config.NullValues)
	}

	return stats, nil
}

func (r *CSVStatsGatherer) sampleRecords(file *os.File, fileSize int64, config stat.SamplingConfig) ([][]string, int64, error) {
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

func (r *CSVStatsGatherer) readFromPosition(file *os.File, maxRecords int) ([][]string, error) {
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

func (r *CSVStatsGatherer) estimateRowCount(fileSize int64, readerBytes int64, config stat.SamplingConfig) int64 {
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

func (r *CSVStatsGatherer) analyzeColumn(records [][]string, colIdx int, colName string, columnStats *stat.ColumnStats, nullValues []string) {
	var nullCount int64
	var minVal, maxVal interface{}

	var isNumeric bool = true
	var isFloat bool = false
	var isDate bool = true
	var isDateTime bool = true
	var isBool bool = true

	var numericValues []float64

	for idx, record := range records {
		if colIdx >= len(record) {
			nullCount++
			continue
		}

		value := strings.TrimSpace(record[colIdx])
		if slices.Contains(nullValues, value) {
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
				continue
			} else {
				if idx != 0 {
					isDate = false
					isDateTime = false
					isBool = false
				}

				isNumeric = false
				isFloat = false
			}
		}
		if isDateTime {
			if _, err := time.Parse(time.RFC3339, value); err == nil {
				continue
			}
			isDateTime = false

			if idx != 0 {
				isDate = false
				isBool = false
			}
		}

		if isDate {
			if _, err := time.Parse(time.DateOnly, value); err == nil {
				continue
			}
			isDate = false

			if idx != 0 {
				isBool = false
			}
		}

		if isBool {
			if boolValue := strings.ToLower(value); boolValue == "true" || boolValue == "false" {
				continue
			}
			isBool = false
		}

		// String comparison
		if minVal == nil || value < toStringComparable(minVal) {
			minVal = value
		}
		if maxVal == nil || value > toStringComparable(maxVal) {
			maxVal = value
		}
	}

	// Set column type
	if isNumeric {
		if isFloat {
			columnStats.Type = "float64"
		} else {
			columnStats.Type = "int64"
		}

		// Calculate aggregates for numeric columns
		if len(numericValues) > 0 {
			columnStats.Aggregate = calculateAggregates(numericValues)
		}
	} else if isDateTime {
		columnStats.Type = "datetime"
	} else if isDate {
		columnStats.Type = "date"
	} else if isBool {
		columnStats.Type = "bool"
	} else {
		columnStats.Type = "string"
	}

	columnStats.NullCount = nullCount
	columnStats.NullPercentage = (float64(nullCount) / float64(len(records))) * 100
	columnStats.MinValue = minVal
	columnStats.MaxValue = maxVal
}

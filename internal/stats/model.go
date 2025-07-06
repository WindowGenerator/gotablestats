package stats

// AggregateStats represents statistical aggregations
type AggregateStats struct {
	Count       int64
	Sum         float64
	Mean        float64
	Median      float64
	StdDev      float64
	Variance    float64
	Percentiles map[int]float64 // 25th, 50th, 75th, 90th, 95th, 99th
}

// TableStats represents the statistics we want to collect
type TableStats struct {
	RowCount       int64
	EstimatedRows  int64 // Estimated total rows based on sampling
	ColumnCount    int
	ColumnNames    []string
	ColumnTypes    map[string]string
	NullCounts     map[string]int64
	NullPercentage map[string]float64
	MinValues      map[string]interface{}
	MaxValues      map[string]interface{}
	SampleData     [][]string
	Aggregates     map[string]*AggregateStats // For numeric columns
	SamplingConfig SamplingConfig
}

// SamplingConfig controls the sampling behavior
type SamplingConfig struct {
	SampleSize      int     // Number of rows to sample
	RandomPositions int     // Number of random positions to seek to
	Confidence      float64 // Confidence level for estimates
	MaxFileSize     int64   // Max file size to process entirely
}

// DefaultSamplingConfig returns sensible defaults
func DefaultSamplingConfig() SamplingConfig {
	return SamplingConfig{
		SampleSize:      1000,
		RandomPositions: 10,
		Confidence:      0.95,
		MaxFileSize:     100 * 1024 * 1024, // 100MB
	}
}

// TableReader defines the strategy interface for reading different table formats
type TableReader interface {
	ReadTable(filePath string, config SamplingConfig) (*TableStats, error)
	GetFormatName() string
}

// StatisticsGenerator is the context that uses the strategy
type StatisticsGenerator struct {
	reader TableReader
	config SamplingConfig
}

// NewStatisticsGenerator creates a new statistics generator with a specific reader strategy
func NewStatisticsGenerator(reader TableReader, config SamplingConfig) *StatisticsGenerator {
	return &StatisticsGenerator{
		reader: reader,
		config: config,
	}
}

// SetReader allows changing the strategy at runtime
func (sg *StatisticsGenerator) SetReader(reader TableReader) {
	sg.reader = reader
}

// GenerateStats generates statistics using the current reader strategy
func (sg *StatisticsGenerator) GenerateStats(filePath string) (*TableStats, error) {
	return sg.reader.ReadTable(filePath, sg.config)
}

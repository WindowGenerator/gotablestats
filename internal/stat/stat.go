package stat

type AggregateStats struct {
	Count       int64           `json:"count"`
	Sum         float64         `json:"sum"`
	Mean        float64         `json:"mean"`
	Median      float64         `json:"median"`
	StdDev      float64         `json:"stddev"`
	Variance    float64         `json:"variance"`
	Percentiles map[int]float64 `json:"percentiles"` // 25th, 50th, 75th, 90th, 95th, 99th
}

type ColumnStats struct {
	Type           string          `json:"type"`
	NullCount      int64           `json:"null_count"`
	NullPercentage float64         `json:"null_percentage"`
	MinValue       any             `json:"min"`
	MaxValue       any             `json:"max"`
	Aggregate      *AggregateStats `json:"aggregate"`
}

// TableStats represents the statistics we want to collect
type TableStats struct {
	RowCount      int64                   `json:"row_count"`
	EstimatedRows int64                   `json:"estimated_rows"` // Estimated total rows based on sampling
	ColumnCount   int                     `json:"column_count"`
	ColumnNames   []string                `json:"column_names"`
	SampleData    [][]string              `json:"sample_data"`
	ColumnsStats  map[string]*ColumnStats `json:"column_stats"`
}

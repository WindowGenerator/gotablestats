package stat

type SamplingConfig struct {
	SampleSize      int
	RandomPositions int
	Confidence      float64
	MaxFileSize     int64
	NullValues      []string
}

func DefaultSamplingConfig() SamplingConfig {
	return SamplingConfig{
		SampleSize:      1000,
		RandomPositions: 10,
		Confidence:      0.95,
		MaxFileSize:     100 * 1024 * 1024, // 100MB
	}
}

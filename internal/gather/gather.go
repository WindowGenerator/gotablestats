package gather

import (
	"os"

	"github.com/WindowGenerator/gotablestats/internal/stat"
)

type StatsGatherer interface {
	Stats(file *os.File, config stat.SamplingConfig) (*stat.TableStats, error)
}

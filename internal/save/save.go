package save

import "github.com/WindowGenerator/gotablestats/internal/stat"

type SaveFormat string

const (
	SaveFormatText SaveFormat = "text"
	SaveFormatJson SaveFormat = "json"
)

type StatsSaver interface {
	Save(stats *stat.TableStats) error
}

type StatsFormatter interface {
	Format(stats *stat.TableStats) (string, error)
}

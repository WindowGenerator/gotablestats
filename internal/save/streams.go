package save

import (
	"os"

	"github.com/WindowGenerator/gotablestats/internal/stat"
)

type StreamSaver struct {
	stream    *os.File
	formatter StatsFormatter
}

func NewStdoutSaver(formatter StatsFormatter) *StreamSaver {
	return &StreamSaver{
		stream:    os.Stdout,
		formatter: formatter,
	}
}

func NewFileSaver(file *os.File, formatter StatsFormatter) *StreamSaver {
	return &StreamSaver{
		stream:    file,
		formatter: formatter,
	}
}

func (saver *StreamSaver) Save(stats *stat.TableStats) error {
	formattedStats, err := saver.formatter.Format(stats)
	if err != nil {
		return err
	}
	_, err = saver.stream.WriteString(formattedStats)
	return err
}

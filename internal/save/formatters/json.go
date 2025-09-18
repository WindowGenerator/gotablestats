package formatters

import (
	"encoding/json"

	"github.com/WindowGenerator/gotablestats/internal/stat"
)

type JSONFormatter struct{}

func NewJsonFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

func (formatter *JSONFormatter) Format(stats *stat.TableStats) (string, error) {
	jsonData, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

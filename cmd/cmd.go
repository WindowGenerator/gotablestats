package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/WindowGenerator/gotablestats/internal/gather"
	"github.com/WindowGenerator/gotablestats/internal/save"
	"github.com/WindowGenerator/gotablestats/internal/save/formatters"
	"github.com/WindowGenerator/gotablestats/internal/stat"
	"github.com/spf13/cobra"
)

var (
	inputFile  string
	outputFile string
	format     string
	sampleSize int
	positions  int
	confidence float64
	maxSize    int64
	nullValues []string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gotablestats",
	Short: "A tool for analyzing table statistics from CSV/TSV files",
	Long: `gotablestats is a CLI tool that processes CSV and TSV files to generate
statistical analysis with sampling capabilities for large files.

The tool automatically detects file format based on extension and provides
detailed statistics about your data including column types, distributions,
and quality metrics.`,
	Example: `  gotablestats -input data.csv
  gotablestats -input large.tsv -sample-size 5000 -positions 10
  gotablestats -input data.csv -confidence 0.99`,
	Run: func(cmd *cobra.Command, args []string) {
		if inputFile == "" {
			fmt.Fprintf(os.Stderr, "Error: Input file is required\n")
			cmd.Help()
			os.Exit(1)
		}

		// Create config from CLI args
		config := stat.SamplingConfig{
			SampleSize:      sampleSize,
			RandomPositions: positions,
			Confidence:      confidence,
			NullValues:      nullValues,
			MaxFileSize:     maxSize,
		}

		// Validate config
		if err := validateConfig(config); err != nil {
			log.Fatal(err)
		}

		// Process file
		start := time.Now()
		stats_, err := processFile(inputFile, config)
		if err != nil {
			log.Fatalf("Error processing file: %v", err)
		}
		processTime := time.Since(start).String()
		log.Printf("Process time: %v", processTime)

		err = saveStats(
			outputFile,
			format,
			stats_,
		)
		if err != nil {
			log.Fatal(err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Define flags
	rootCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input file (CSV or TSV) (required)")
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file")
	rootCmd.Flags().StringVarP(&format, "format", "f", "text", "Output format")
	rootCmd.Flags().IntVarP(&sampleSize, "sample-size", "s", 1000, "Number of rows to sample")
	rootCmd.Flags().IntVarP(&positions, "positions", "p", 5, "Number of random positions")
	rootCmd.Flags().Float64VarP(&confidence, "confidence", "c", 0.95, "Confidence level (0-1)")
	rootCmd.Flags().Int64VarP(&maxSize, "max-size", "m", 100*1024*1024, "Max file size for full processing (bytes)")
	rootCmd.Flags().StringSliceVarP(&nullValues, "nulls", "n", []string{"", "null", "NULL"}, "Comma-separated list of values to treat as NULL/missing")

	// Mark required flags
	rootCmd.MarkFlagRequired("input")
}

func validateConfig(config stat.SamplingConfig) error {
	if config.SampleSize <= 0 {
		return fmt.Errorf("sample size must be positive")
	}
	if config.RandomPositions <= 0 {
		return fmt.Errorf("random positions must be positive")
	}
	if config.Confidence <= 0 || config.Confidence >= 1 {
		return fmt.Errorf("confidence must be between 0 and 1")
	}
	return nil
}

func processFile(filePath string, config stat.SamplingConfig) (*stat.TableStats, error) {
	_, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot access file: %v", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(filePath))
	var reader gather.StatsGatherer

	switch ext {
	case ".csv":
		reader = gather.NewCSVStatsGatherer(',')
	case ".tsv":
		reader = gather.NewCSVStatsGatherer('\t')
	case ".parquet":
		reader = gather.NewParquetReader()

	default:
		return nil, fmt.Errorf("cannot auto-detect delimiter for %s, unsupported file type", ext)
	}

	return reader.Stats(file, config)
}

func saveStats(filePath string, format string, stats *stat.TableStats) error {
	var formatter save.StatsFormatter
	var saver save.StatsSaver

	saveFormat := (save.SaveFormat)(format)

	switch saveFormat {
	case save.SaveFormatText:
		formatter = formatters.NewTextFormatter()
	case save.SaveFormatJson:
		formatter = formatters.NewJsonFormatter()
	default:
		return fmt.Errorf("there is no such formatter %s, for output statistics", format)
	}

	if filePath == "" {
		saver = save.NewStdoutSaver(formatter)
	} else {
		file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()
		saver = save.NewFileSaver(file, formatter)
	}

	return saver.Save(stats)
}

package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/WindowGenerator/gotablestats/internal/stats"
	"github.com/spf13/cobra"
)

var (
	inputFile  string
	sampleSize int
	positions  int
	confidence float64
	maxSize    int64
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
		config := stats.SamplingConfig{
			SampleSize:      sampleSize,
			RandomPositions: positions,
			Confidence:      confidence,
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

		stats.PrintStats(stats_, "")
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
	rootCmd.Flags().IntVarP(&sampleSize, "sample-size", "s", 1000, "Number of rows to sample")
	rootCmd.Flags().IntVarP(&positions, "positions", "p", 5, "Number of random positions")
	rootCmd.Flags().Float64VarP(&confidence, "confidence", "c", 0.95, "Confidence level (0-1)")
	rootCmd.Flags().Int64VarP(&maxSize, "max-size", "m", 100*1024*1024, "Max file size for full processing (bytes)")

	// Mark required flags
	rootCmd.MarkFlagRequired("input")
}

func validateConfig(config stats.SamplingConfig) error {
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

func processFile(filePath string, config stats.SamplingConfig) (*stats.TableStats, error) {
	_, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("cannot access file: %v", err)
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	var reader stats.TableReader

	switch ext {
	case ".csv":
		reader = &stats.CSVReader{
			Delimiter: ',',
		}
	case ".tsv":
		reader = &stats.TSVReader{}
	default:
		return nil, fmt.Errorf("cannot auto-detect delimiter for %s, unsupported file type", ext)
	}

	return reader.ReadTable(filePath, config)
}

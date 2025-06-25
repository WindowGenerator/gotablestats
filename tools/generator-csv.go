package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
)

// Data arrays for random generation
var (
	departments = []string{"Engineering", "Marketing", "Sales", "HR", "Finance", "Operations", "Legal", "IT"}
	categories  = []string{"A", "B", "C", "D", "E"}
	domains     = []string{"gmail.com", "yahoo.com", "company.com", "outlook.com", "hotmail.com"}
	firstNames  = []string{"John", "Jane", "Michael", "Sarah", "David", "Lisa", "Robert", "Emily", "James", "Ashley", "Chris", "Jessica", "Daniel", "Amanda", "Matthew", "Nicole", "William", "Jennifer", "Richard", "Michelle", "Joseph", "Kimberly", "Thomas", "Amy", "Charles", "Angela", "Christopher", "Brenda", "Mark", "Emma"}
	lastNames   = []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez", "Wilson", "Anderson", "Thomas", "Taylor", "Moore", "Jackson", "Martin", "Lee", "Perez", "Thompson", "White", "Harris", "Sanchez", "Clark", "Ramirez", "Lewis", "Robinson"}
)

type Config struct {
	Rows     int
	Filename string
	Workers  int
}

type Record struct {
	ID         int
	Name       string
	Email      string
	Age        int
	Salary     int
	Department string
	JoinDate   string
	Active     bool
	Score      float64
	Category   string
}

func main() {
	// Parse command line flags
	var config Config
	flag.IntVar(&config.Rows, "rows", 1000000, "Number of rows to generate")
	flag.StringVar(&config.Filename, "file", "big_data.csv", "Output filename")
	flag.IntVar(&config.Workers, "workers", 4, "Number of worker goroutines")
	flag.Parse()

	fmt.Printf("Generating CSV with %d rows...\n", config.Rows)
	fmt.Printf("Output file: %s\n", config.Filename)
	fmt.Printf("Workers: %d\n", config.Workers)

	startTime := time.Now()

	// Generate CSV
	if err := generateCSV(config); err != nil {
		fmt.Printf("Error generating CSV: %v\n", err)
		os.Exit(1)
	}

	duration := time.Since(startTime)

	// Get file stats
	fileInfo, err := os.Stat(config.Filename)
	if err != nil {
		fmt.Printf("Error getting file stats: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\n✅ CSV generation complete!")
	fmt.Printf("📄 File: %s\n", config.Filename)
	fmt.Printf("📊 Rows: %d (plus header)\n", config.Rows)
	fmt.Printf("💾 Size: %.2f MB\n", float64(fileInfo.Size())/1024/1024)
	fmt.Printf("⏱️  Time: %v\n", duration)
	fmt.Printf("🚀 Speed: %.0f rows/second\n", float64(config.Rows)/duration.Seconds())

	// Show sample data
	fmt.Println("\nSample data:")
	showSample(config.Filename)
}

func generateCSV(config Config) error {
	// Create output file
	file, err := os.Create(config.Filename)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write header
	header := []string{"id", "name", "email", "age", "salary", "department", "join_date", "active", "score", "category"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("writing header: %w", err)
	}

	// Channel for generated records
	recordChan := make(chan []Record, 100)
	done := make(chan bool)

	// Start workers
	for i := 0; i < config.Workers; i++ {
		go recordGenerator(recordChan, done)
	}

	// Progress tracking
	progressInterval := config.Rows / 100
	if progressInterval == 0 {
		progressInterval = 1
	}

	batchSize := 10000
	recordsGenerated := 0

	// Generate and write records
	for recordsGenerated < config.Rows {
		remainingRows := config.Rows - recordsGenerated
		currentBatchSize := batchSize
		if remainingRows < batchSize {
			currentBatchSize = remainingRows
		}

		// Request batch generation
		recordChan <- make([]Record, currentBatchSize)

		// Get generated batch
		batch := <-recordChan

		// Write batch to CSV
		for i, record := range batch {
			row := []string{
				strconv.Itoa(recordsGenerated + i + 1),
				record.Name,
				record.Email,
				strconv.Itoa(record.Age),
				strconv.Itoa(record.Salary),
				record.Department,
				record.JoinDate,
				strconv.FormatBool(record.Active),
				fmt.Sprintf("%.2f", record.Score),
				record.Category,
			}
			if err := writer.Write(row); err != nil {
				return fmt.Errorf("writing record: %w", err)
			}
		}

		recordsGenerated += len(batch)

		// Show progress
		if recordsGenerated%progressInterval == 0 || recordsGenerated == config.Rows {
			percentage := recordsGenerated * 100 / config.Rows
			fmt.Printf("Progress: %d%% (%d/%d rows)\n", percentage, recordsGenerated, config.Rows)
		}

		// Flush periodically
		if recordsGenerated%50000 == 0 {
			writer.Flush()
		}
	}

	// Stop workers
	for i := 0; i < config.Workers; i++ {
		done <- true
	}

	return nil
}

func recordGenerator(recordChan chan []Record, done chan bool) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for {
		select {
		case batch := <-recordChan:
			// Generate records for the batch
			for i := range batch {
				batch[i] = generateRecord(rng)
			}
			recordChan <- batch
		case <-done:
			return
		}
	}
}

func generateRecord(rng *rand.Rand) Record {
	firstName := firstNames[rng.Intn(len(firstNames))]
	lastName := lastNames[rng.Intn(len(lastNames))]

	return Record{
		Name: fmt.Sprintf("%s %s", firstName, lastName),
		Email: fmt.Sprintf("%s.%s%d@%s",
			toLowerCase(firstName),
			toLowerCase(lastName),
			rng.Intn(9999),
			domains[rng.Intn(len(domains))]),
		Age:        22 + rng.Intn(44),        // 22-65
		Salary:     30000 + rng.Intn(120000), // 30k-150k
		Department: departments[rng.Intn(len(departments))],
		JoinDate:   generateRandomDate(rng),
		Active:     rng.Intn(2) == 0,
		Score:      rng.Float64() * 100, // 0-100
		Category:   categories[rng.Intn(len(categories))],
	}
}

func generateRandomDate(rng *rand.Rand) string {
	year := 2015 + rng.Intn(9) // 2015-2023
	month := 1 + rng.Intn(12)
	day := 1 + rng.Intn(28)
	return fmt.Sprintf("%d-%02d-%02d", year, month, day)
}

func toLowerCase(s string) string {
	result := make([]byte, len(s))
	for i, c := range []byte(s) {
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

func showSample(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening file for sample: %v\n", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	for i := 0; i < 5; i++ {
		record, err := reader.Read()
		if err != nil {
			break
		}
		fmt.Printf("%s\n", record)
	}
}

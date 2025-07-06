
# gotablestats

**gotablestats** is a fast and flexible command-line tool for generating detailed statistical summaries from CSV and TSV files. It supports sampling for large datasets, making it ideal for quick exploratory data analysis (EDA) without loading entire files into memory.

## Features

- 📊 Detects column data types and distributions
- 📁 Supports CSV and TSV files (auto-detect by extension)
- 🔍 Smart sampling with configurable sample size and confidence level
- 📈 Provides quality metrics for your tabular data
- ⚡ Efficient processing for large files with file size limit

## Installation

Precompiled `go-critic` binaries can be found at [releases](https://github.com/WindowGenerator/gotablestats/releases) page.

You can build `gotablestats` from source:

```bash
git clone https://github.com/WindowGenerator/gotablestats.git
cd gotablestats
go build -o gotablestats
```

Then move it to your `$PATH`, for example:

```bash
mv gotablestats /usr/local/bin/
```

## Usage

```bash
gotablestats --input <file> [flags]
```

### Required

* `-i, --input`: Input file (CSV or TSV)

### Optional Flags

| Flag                | Default     | Description                                                |
| ------------------- | ----------- | ---------------------------------------------------------- |
| `-s, --sample-size` | `1000`      | Number of rows to sample                                   |
| `-p, --positions`   | `5`         | Number of random positions to select during sampling       |
| `-c, --confidence`  | `0.95`      | Confidence level for statistical inference (0–1)           |
| `-m, --max-size`    | `104857600` | Max file size in bytes for full processing (default 100MB) |

### Examples

```bash
# Basic usage on a CSV file
gotablestats --input data.csv

# Use a larger sample size for a TSV file
gotablestats -i data.tsv -s 5000

# Adjust confidence level and number of sampling positions
gotablestats -i data.csv -c 0.99 -p 10

# Avoid full processing if file exceeds 50MB
gotablestats -i huge.csv -m 52428800
```

## Output

The tool prints a human-readable report to stdout, including:

* Column names and inferred data types
* Value distribution (e.g., min/max, unique count)
* Missing value stats
* Quality checks based on sampling

## How It Works

* Determines file format from extension (`.csv` or `.tsv`)
* Samples rows from random positions to ensure fair representation
* Computes descriptive statistics and structural info
* Avoids memory overload by limiting file size for full parsing

## Limitations

* Currently supports only `.csv` and `.tsv` formats
* Assumes UTF-8 encoding
* Designed for tabular files where the first row is a header

## Roadmap

See the [ROADMAP.md](./ROADMAP.md) for upcoming features and development plans.

## Development

This CLI tool is built using:

* [Go](https://golang.org)
* [Cobra](https://github.com/spf13/cobra) for CLI scaffolding
* Custom readers and statistical analyzers in the `internal/stats` package

## License

MIT © 2025 [Window Generator](https://github.com/WindowGenerator)


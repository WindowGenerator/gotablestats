# gotablestats Roadmap

A forward-looking roadmap for planned improvements to `gotablestats` — a lightweight and efficient CLI tool for analyzing tabular data.

---

## 🤩 Parquet File Support

### Track

Input Format Extension

### Motivation

Expand usability by supporting columnar data formats common in data pipelines and analytics workflows.

### Description

Implement functionality to read `.parquet` files using an efficient Go-based library such as `github.com/xitongsys/parquet-go`. Ensure seamless integration with existing analytics, including type inference, sampling, and value distribution.

---

## 📊 Extended Analytics

### Track

Analytical Features

### Motivation

Increase the analytical value by calculating more advanced metrics useful for deeper data quality assessment and data science workflows.

### Description

Add statistics such as:

- [X] Standard deviation and variance
- [ ] Skewness and kurtosis
- [ ] Correlation between numeric columns
- [ ] Basic histograms or frequency bins

These should be optional and controlled via flags to maintain performance for lightweight use.

---

## 📂 Save Analytics Output

### Track

Output & Integration

### Motivation

Enable integration with other tools and workflows by supporting persistent output formats.

### Description

Implement file-based output options including:

- [ ] JSON
- [ ] CSV (summary per column)
- [ ] Markdown (for human-readable reports)

Allow users to specify an output file path with `--output` and optionally choose format with `--format`.

---

## 🛠️ Custom Null Value Configuration

### Track

Data Cleaning Customization

### Motivation

Different datasets may use different placeholders for missing or null values; allowing customization improves accuracy.

### Description

Introduce a flag like `--nulls` that accepts a comma-separated list of values (e.g., `"NA,null,missing"`) which should be interpreted as missing. This will improve completeness and quality metrics.

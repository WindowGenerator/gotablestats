# gotablestats Roadmap

A forward-looking roadmap for planned improvements to `gotablestats` — a lightweight and efficient CLI tool for analyzing tabular data.

---

## 🤩 Parquet File Support

### Motivation

Expand usability by supporting columnar data formats common in data pipelines and analytics workflows.

### Description

Implement functionality to read `.parquet` files using an efficient Go-based library such as `github.com/xitongsys/parquet-go`. Ensure seamless integration with existing analytics, including type inference, sampling, and value distribution.

---

## 📊 Extended Analytics

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

### Motivation

Different datasets may use different placeholders for missing or null values; allowing customization improves accuracy.

### Description

Introduce a flag like `--nulls` that accepts a comma-separated list of values (e.g., `"NA,null,missing"`) which should be interpreted as missing. This will improve completeness and quality metrics.

---

## 🗒️ New Data Types Support

### Motivation
Improve accuracy and usefulness of analysis by supporting common data types not currently recognized.

### Description
Implement support for additional types:
- [ ] `Date`
- [ ] `Datetime`
- [ ] `Boolean`
- [ ] `URL like`
- [ ] `Email like?`

Enhance type inference logic and provide type-specific statistics such as weekday distribution for dates or true/false counts for booleans.

---

## ⬆️ New Quantile Statistics (Q1, Q3)


### Motivation
Provide better insight into distribution shape and potential outliers by reporting quartiles.

### Description
Extend statistical output with first (Q1) and third (Q3) quartiles in addition to the existing min, max, and mean metrics.
Support optional flag to enable/disable quartile computation.

---

## ⚖️ Measurement Error Estimation

### Motivation
Provide transparency about uncertainty in sampled statistics, especially on incomplete or sampled datasets.

### Description
Estimate and display standard error or confidence intervals for key metrics like mean, proportion of nulls, and unique value count.
Allow tuning precision via sample size and confidence flags.

---

## 🚀 Multiprocessing and Multithreading Settings

### Motivation
Improve speed and scalability when working with large datasets on modern multi-core systems.

### Description
Introduce optional CLI flags or config-based control for:
- [ ] Number of goroutines (threads)
- [ ] Use of multiprocessing for file chunks or column groups
- [ ] Tuning performance vs memory usage trade-offs

Provide benchmarks and sensible defaults for typical dataset sizes.

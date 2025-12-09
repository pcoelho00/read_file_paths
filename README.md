# File Path Scanner

A high-performance Go CLI tool that recursively scans a directory and generates a CSV file containing file paths and their string lengths.

## Features

- **High Performance**: Uses `filepath.WalkDir` to minimize system calls and implements a concurrent producer-consumer pattern for maximum throughput.
- **Low Memory Footprint**: Streams results to disk in configurable batches instead of holding everything in RAM.
- **Visual Feedback**: Includes a real-time terminal spinner and live file counter.
- **Cross-Platform**: Compiles and runs on Linux, Windows, and macOS.

## Installation

Ensure you have Go installed, then build the binary:

```bash
go build -o file_paths main.go
```

## Usage

```bash
./file_paths <directory> [batch_size]
```

### Arguments

- `<directory>`: **(Required)** The absolute or relative path to the directory you want to scan.
- `[batch_size]`: **(Optional)** The number of records to group together before writing to disk. Defaults to `100`. Larger batches (e.g., 1000-5000) may improve performance on very large file systems.

### Examples

Scan the current directory:
```bash
./file_paths .
```

Scan a specific folder with a custom batch size of 500:
```bash
./file_paths /home/user/projects 500
```

## Output

The tool creates a `file_paths.csv` file in your current working directory with the following format:

```csv
file_path,path_length
/home/user/projects/main.go,25
/home/user/projects/README.md,27
/home/user/projects/data/config.json,34
```

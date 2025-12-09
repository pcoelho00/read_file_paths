package main

import (
	"encoding/csv"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const defaultBatchSize = 100

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <directory> [batch_size]\n", os.Args[0])
		os.Exit(1)
	}

	dirPath := os.Args[1]

	batchSize := defaultBatchSize
	if len(os.Args) >= 3 {
		size, err := strconv.Atoi(os.Args[2])
		if err != nil || size <= 0 {
			fmt.Fprintf(os.Stderr, "Error: batch_size must be a positive integer\n")
			os.Exit(1)
		}
		batchSize = size
	}

	// Verify the path is a directory
	info, err := os.Stat(dirPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error accessing path: %v\n", err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: %s is not a directory\n", dirPath)
		os.Exit(1)
	}

	outputFile, err := os.Create("file_paths.csv")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating CSV file: %v\n", err)
		os.Exit(1)
	}
	defer outputFile.Close()

	writer := csv.NewWriter(outputFile)
	defer writer.Flush()

	if err := writer.Write([]string{"file_path", "path_length"}); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing CSV header: %v\n", err)
		os.Exit(1)
	}

	// Channels and Sync
	// Buffer allows producer to continue scanning while consumer is writing
	pathChan := make(chan string, 1000)
	var fileCount int64 // Atomic counter
	var scanErr error
	var wg sync.WaitGroup

	// 1. Spinner Goroutine
	done := make(chan bool)
	wg.Add(1)
	go func() {
		defer wg.Done()
		spinChars := []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}
		i := 0
		for {
			select {
			case <-done:
				return
			default:
				// Atomic load for thread safety
				count := atomic.LoadInt64(&fileCount)
				fmt.Printf("\r%c Scanning... %d files found", spinChars[i%len(spinChars)], count)
				i++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// 2. Producer Goroutine (Scanner)
	// Runs concurrently with the writer
	go func() {
		defer close(pathChan)
		// optimization: Use WalkDir instead of Walk (avoids extra os.Stat calls)
		scanErr = filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			// d.IsDir() checks the directory entry directly, no extra syscall needed
			if !d.IsDir() {
				pathChan <- path
			}
			return nil
		})
	}()

	// 3. Consumer (Writer)
	// Main goroutine consumes paths from channel and writes to CSV
	batch := make([][]string, 0, batchSize)
	for path := range pathChan {
		pathLength := len(path)
		batch = append(batch, []string{path, strconv.Itoa(pathLength)})

		if len(batch) >= batchSize {
			if err := writer.WriteAll(batch); err != nil {
				fmt.Fprintf(os.Stderr, "\nError writing batch: %v\n", err)
				os.Exit(1)
			}
			// Atomic add
			atomic.AddInt64(&fileCount, int64(len(batch)))
			batch = batch[:0] // Reset batch
		}
	}

	// Write remaining records
	if len(batch) > 0 {
		if err := writer.WriteAll(batch); err != nil {
			fmt.Fprintf(os.Stderr, "\nError writing final batch: %v\n", err)
			os.Exit(1)
		}
		atomic.AddInt64(&fileCount, int64(len(batch)))
	}

	// Stop spinner
	done <- true
	wg.Wait()
	fmt.Print("\r\033[K") // Clear line

	if scanErr != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", scanErr)
		os.Exit(1)
	}

	fmt.Printf("Done! Processed %d files.\n", atomic.LoadInt64(&fileCount))
	fmt.Println("CSV file created: file_paths.csv")
}

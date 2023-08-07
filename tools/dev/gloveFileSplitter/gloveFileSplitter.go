package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	prnt := "glove_segments_6B_300d_1000"
	// Open the input file
	inputFile, err := os.Open("glove.6B.300d.txt")
	if err != nil {
		fmt.Println("Error opening input file:", err)
		return
	}
	defer inputFile.Close()

	// Create a scanner to read the input file
	scanner := bufio.NewScanner(inputFile)

	// Specify the number of lines per output file
	linesPerFile := 1000

	// Initialize variables
	lineCount := 0
	fileCount := 1

	// Create the first output file
	outputFile, err := os.Create(filepath.Join(".", prnt, fmt.Sprintf("output%d.txt", fileCount)))
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outputFile.Close()

	// Create a writer for the current output file
	writer := bufio.NewWriter(outputFile)

	// Iterate over the lines of the input file
	for scanner.Scan() {
		line := scanner.Text()
		// Write the line to the current output file
		fmt.Fprintln(writer, line)

		// Increment the line count
		lineCount++

		// Check if the current output file has reached the maximum lines
		if lineCount >= linesPerFile {
			// Flush the writer to write any remaining data to the output file
			writer.Flush()

			// Increment the file count
			fileCount++

			// Create the next output file
			outputFile, err = os.Create(filepath.Join(".", prnt, fmt.Sprintf("output%d.txt", fileCount)))
			if err != nil {
				fmt.Println("Error creating output file:", err)
				return
			}
			defer outputFile.Close()

			// Create a writer for the next output file
			writer = bufio.NewWriter(outputFile)

			// Reset the line count
			lineCount = 0
		}
	}

	// Flush the writer to write any remaining data to the last output file
	writer.Flush()

	// Check if there was an error during scanning
	if err := scanner.Err(); err != nil {
		fmt.Println("Error scanning input file:", err)
		return
	}

	fmt.Println("File splitting completed.")
}

package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run replace_errors.go <file>")
		os.Exit(1)
	}

	filename := os.Args[1]

	// Read file
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	// Replace http.Error with s.respondWithError
	re := regexp.MustCompile(`http\.Error\(w,\s*(.+?),\s*(http\.Status\w+)\)`)

	for i, line := range lines {
		if strings.Contains(line, "http.Error(w,") {
			lines[i] = re.ReplaceAllString(line, `s.respondWithError(w, $2, $1)`)
		}
	}

	// Write file
	output, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating output file: %v\n", err)
		os.Exit(1)
	}
	defer output.Close()

	writer := bufio.NewWriter(output)
	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}
	writer.Flush()

	fmt.Printf("Successfully updated %s\n", filename)
}

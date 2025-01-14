package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Custom errors
type ValidationError struct {
	message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error: %s", e.message)
}

// Converter interface
type Converter interface {
	Convert(input string) (*Table, error)
	Format(table *Table) (string, error)
}

type Table struct {
	Headers []string
	Rows    [][]string
}

// MarkdownConverter implements Converter
type MarkdownConverter struct{}

func NewMarkdownConverter() *MarkdownConverter {
	return &MarkdownConverter{}
}

func (m *MarkdownConverter) Convert(content string) (*Table, error) {
	return parseMarkdownTable(content)
}

func (m *MarkdownConverter) Format(table *Table) (string, error) {
	var sb strings.Builder
	writer := csv.NewWriter(&sb)

	if err := writer.Write(table.Headers); err != nil {
		return "", fmt.Errorf("failed to write headers: %w", err)
	}

	if err := writer.WriteAll(table.Rows); err != nil {
		return "", fmt.Errorf("failed to write rows: %w", err)
	}

	writer.Flush()
	return sb.String(), nil
}

// CSVConverter implements Converter
type CSVConverter struct{}

func NewCSVConverter() *CSVConverter {
	return &CSVConverter{}
}

func (c *CSVConverter) Convert(content string) (*Table, error) {
	reader := csv.NewReader(strings.NewReader(content))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) == 0 {
		return nil, &ValidationError{message: "empty CSV file"}
	}

	return &Table{
		Headers: records[0],
		Rows:    records[1:],
	}, nil
}

func (c *CSVConverter) Format(table *Table) (string, error) {
	// Calculate column widths
	colWidths := make([]int, len(table.Headers))
	for i, header := range table.Headers {
		colWidths[i] = len(header)
	}
	for _, row := range table.Rows {
		for i, cell := range row {
			if len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	var sb strings.Builder

	// Write headers
	sb.WriteString("|")
	for i, header := range table.Headers {
		sb.WriteString(fmt.Sprintf(" %-*s |", colWidths[i], header))
	}
	sb.WriteString("\n|")

	// Write separator
	for _, width := range colWidths {
		sb.WriteString(strings.Repeat("-", width+2))
		sb.WriteString("|")
	}
	sb.WriteString("\n")

	// Write data rows
	for _, row := range table.Rows {
		sb.WriteString("|")
		for i, cell := range row {
			sb.WriteString(fmt.Sprintf(" %-*s |", colWidths[i], cell))
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

func parseMarkdownTable(content string) (*Table, error) {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) < 3 {
		return nil, fmt.Errorf("invalid markdown table: minimum 3 lines required")
	}

	// Parse headers
	headers := parseRow(lines[0])
	if len(headers) == 0 {
		return nil, fmt.Errorf("invalid markdown table: no headers found")
	}

	// Validate separator line
	separator := parseRow(lines[1])
	if len(separator) != len(headers) {
		return nil, fmt.Errorf("invalid markdown table: separator line doesn't match headers")
	}
	for _, sep := range separator {
		if !isValidSeparator(sep) {
			return nil, fmt.Errorf("invalid markdown table: invalid separator line")
		}
	}

	// Parse data rows
	var rows [][]string
	for _, line := range lines[2:] {
		if strings.TrimSpace(line) == "" {
			continue
		}
		row := parseRow(line)
		if len(row) != len(headers) {
			return nil, fmt.Errorf("invalid markdown table: inconsistent column count in row")
		}
		rows = append(rows, row)
	}

	return &Table{Headers: headers, Rows: rows}, nil
}

func parseRow(line string) []string {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "|") || !strings.HasSuffix(line, "|") {
		return nil
	}
	line = line[1 : len(line)-1]
	cells := strings.Split(line, "|")
	for i, cell := range cells {
		cells[i] = strings.TrimSpace(cell)
	}
	return cells
}

func isValidSeparator(sep string) bool {
	sep = strings.TrimSpace(sep)
	if len(sep) == 0 {
		return false
	}
	for _, ch := range sep {
		if ch != '-' && ch != ':' {
			return false
		}
	}
	return true
}

func getDefaultOutputPath(inputPath string, toCSV bool) string {
	ext := filepath.Ext(inputPath)
	basePath := inputPath[:len(inputPath)-len(ext)]
	if toCSV {
		return basePath + ".csv"
	}
	return basePath + ".md"
}

// Add Config struct
type Config struct {
	InputPath  string
	OutputPath string
	ToCSV      bool
}

// Update main function
func main() {
	config := parseFlags()

	content, err := os.ReadFile(config.InputPath)
	if err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		os.Exit(1)
	}

	var converter Converter
	if config.ToCSV {
		converter = NewMarkdownConverter()
	} else {
		converter = NewCSVConverter()
	}

	// Convert input to table
	table, err := converter.Convert(string(content))
	if err != nil {
		fmt.Printf("Error converting input: %v\n", err)
		os.Exit(1)
	}

	// Format table to output
	output, err := converter.Format(table)
	if err != nil {
		fmt.Printf("Error formatting output: %v\n", err)
		os.Exit(1)
	}

	// Write output
	if err := os.WriteFile(config.OutputPath, []byte(output), 0644); err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully converted to %s: %s\n",
		map[bool]string{true: "CSV", false: "markdown"}[config.ToCSV],
		config.OutputPath)
}

func parseFlags() Config {
	toCSV := flag.Bool("to-csv", false, "Convert from markdown to CSV")
	toMD := flag.Bool("to-md", false, "Convert from CSV to markdown")
	inFile := flag.String("in", "", "Input file path")
	outFile := flag.String("out", "", "Output file path (optional)")

	flag.Parse()

	if *inFile == "" {
		fmt.Println("Error: Input file path is required")
		flag.Usage()
		os.Exit(1)
	}

	if *toCSV == *toMD {
		fmt.Println("Error: Must specify exactly one of --to-csv or --to-md")
		flag.Usage()
		os.Exit(1)
	}

	outputPath := *outFile
	if outputPath == "" {
		outputPath = getDefaultOutputPath(*inFile, *toCSV)
	}

	return Config{
		InputPath:  *inFile,
		OutputPath: outputPath,
		ToCSV:      *toCSV,
	}
}

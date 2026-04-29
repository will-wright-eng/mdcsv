package main

import (
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Table struct {
	Headers []string
	Rows    [][]string
}

type Parser interface {
	Parse(input string) (*Table, error)
}

type Formatter interface {
	Format(table *Table) (string, error)
}

var parsers = map[string]Parser{
	"md":  MarkdownParser{},
	"csv": CSVParser{},
}

var formatters = map[string]Formatter{
	"md":  MarkdownFormatter{},
	"csv": CSVFormatter{},
}

type MarkdownParser struct{}

func (MarkdownParser) Parse(content string) (*Table, error) {
	lines := strings.Split(strings.TrimSpace(content), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("invalid markdown table: minimum 2 lines required")
	}

	headers := parseRow(lines[0])
	if len(headers) == 0 {
		return nil, fmt.Errorf("invalid markdown table: no headers found")
	}

	separator := parseRow(lines[1])
	if len(separator) != len(headers) {
		return nil, fmt.Errorf("invalid markdown table: separator line doesn't match headers")
	}
	for _, sep := range separator {
		if !isValidSeparator(sep) {
			return nil, fmt.Errorf("invalid markdown table: invalid separator line")
		}
	}

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

type CSVParser struct{}

func (CSVParser) Parse(content string) (*Table, error) {
	reader := csv.NewReader(strings.NewReader(content))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}
	if len(records) == 0 {
		return nil, errors.New("empty CSV input")
	}
	return &Table{
		Headers: records[0],
		Rows:    records[1:],
	}, nil
}

type CSVFormatter struct{}

func (CSVFormatter) Format(table *Table) (string, error) {
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

type MarkdownFormatter struct{}

func (MarkdownFormatter) Format(table *Table) (string, error) {
	colWidths := make([]int, len(table.Headers))
	for i, header := range table.Headers {
		colWidths[i] = len(header)
	}
	for _, row := range table.Rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	var sb strings.Builder

	sb.WriteString("|")
	for i, header := range table.Headers {
		sb.WriteString(fmt.Sprintf(" %-*s |", colWidths[i], header))
	}
	sb.WriteString("\n|")

	for _, width := range colWidths {
		sb.WriteString(strings.Repeat("-", width+2))
		sb.WriteString("|")
	}
	sb.WriteString("\n")

	for _, row := range table.Rows {
		sb.WriteString("|")
		for i, cell := range row {
			sb.WriteString(fmt.Sprintf(" %-*s |", colWidths[i], cell))
		}
		sb.WriteString("\n")
	}

	return sb.String(), nil
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

type Config struct {
	From   string
	To     string
	Input  string
	Output string
}

func inferFormat(path string) (string, bool) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md":
		return "md", true
	case ".csv":
		return "csv", true
	}
	return "", false
}

func resolveConfig(from, to, input, output string) (Config, error) {
	cfg := Config{From: from, To: to, Input: input, Output: output}

	if cfg.From == "" {
		if cfg.Input != "" && cfg.Input != "-" {
			f, ok := inferFormat(cfg.Input)
			if !ok {
				return cfg, fmt.Errorf("cannot infer input format from %q; pass -f md|csv", cfg.Input)
			}
			cfg.From = f
		} else {
			return cfg, errors.New("input format required: pass -f md|csv (cannot infer from stdin)")
		}
	}

	if cfg.To == "" {
		if cfg.Output != "" && cfg.Output != "-" {
			t, ok := inferFormat(cfg.Output)
			if !ok {
				return cfg, fmt.Errorf("cannot infer output format from %q; pass -t md|csv", cfg.Output)
			}
			cfg.To = t
		} else {
			return cfg, errors.New("output format required: pass -t md|csv (cannot infer from stdout)")
		}
	}

	if _, ok := parsers[cfg.From]; !ok {
		return cfg, fmt.Errorf("unknown input format %q (want md or csv)", cfg.From)
	}
	if _, ok := formatters[cfg.To]; !ok {
		return cfg, fmt.Errorf("unknown output format %q (want md or csv)", cfg.To)
	}

	return cfg, nil
}

const usage = `usage: mdcsv [-f FROM] [-t TO] [-o FILE] [FILE]

Convert between markdown tables and CSV. Reads stdin and writes stdout
by default; FILE or '-' can also stand in for stdin, and -o '-' for
stdout.

Flags:
  -f, --from FORMAT    input format: md or csv
  -t, --to FORMAT      output format: md or csv
  -o, --output FILE    output file (default: stdout)
  -h, --help           this message

Formats are inferred from file extensions when -f / -t are omitted.
`

func parseFlags(args []string) (Config, error) {
	fs := flag.NewFlagSet("mdcsv", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}

	var from, to, output string
	fs.StringVar(&from, "f", "", "")
	fs.StringVar(&from, "from", "", "")
	fs.StringVar(&to, "t", "", "")
	fs.StringVar(&to, "to", "", "")
	fs.StringVar(&output, "o", "", "")
	fs.StringVar(&output, "output", "", "")

	// flag.Parse stops at the first non-flag argument, so we re-parse
	// after each positional to allow `mdcsv data.md -o out.csv` style.
	input := ""
	remaining := args
	for {
		if err := fs.Parse(remaining); err != nil {
			return Config{}, err
		}
		rest := fs.Args()
		if len(rest) == 0 {
			break
		}
		if input != "" {
			return Config{}, fmt.Errorf("expected at most one input file, got more than one")
		}
		input = rest[0]
		remaining = rest[1:]
	}

	return resolveConfig(from, to, input, output)
}

func run(cfg Config, stdin io.Reader, stdout io.Writer) error {
	var (
		content []byte
		err     error
	)
	if cfg.Input == "" || cfg.Input == "-" {
		content, err = io.ReadAll(stdin)
	} else {
		content, err = os.ReadFile(cfg.Input)
	}
	if err != nil {
		return fmt.Errorf("reading input: %w", err)
	}

	table, err := parsers[cfg.From].Parse(string(content))
	if err != nil {
		return fmt.Errorf("parsing input: %w", err)
	}

	out, err := formatters[cfg.To].Format(table)
	if err != nil {
		return fmt.Errorf("formatting output: %w", err)
	}

	if cfg.Output == "" || cfg.Output == "-" {
		_, err = io.WriteString(stdout, out)
		return err
	}
	return os.WriteFile(cfg.Output, []byte(out), 0644)
}

func main() {
	args := os.Args[1:]
	for _, a := range args {
		if a == "-h" || a == "--help" {
			fmt.Print(usage)
			return
		}
	}

	cfg, err := parseFlags(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "mdcsv:", err)
		fmt.Fprintln(os.Stderr)
		fmt.Fprint(os.Stderr, usage)
		os.Exit(1)
	}

	if err := run(cfg, os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "mdcsv:", err)
		os.Exit(1)
	}
}

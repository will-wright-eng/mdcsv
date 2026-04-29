package main

import (
	"reflect"
	"strings"
	"testing"
)

func TestMarkdownParser(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Table
		wantErr bool
	}{
		{
			name: "valid table",
			input: `| a | b |
|---|---|
| 1 | 2 |
| 3 | 4 |`,
			want: &Table{
				Headers: []string{"a", "b"},
				Rows:    [][]string{{"1", "2"}, {"3", "4"}},
			},
		},
		{
			name: "headers only",
			input: `| a | b |
|---|---|`,
			want: &Table{
				Headers: []string{"a", "b"},
				Rows:    nil,
			},
		},
		{
			name: "padded cells are trimmed",
			input: `|   foo  |  bar |
|--------|------|
|   1    |   2  |`,
			want: &Table{
				Headers: []string{"foo", "bar"},
				Rows:    [][]string{{"1", "2"}},
			},
		},
		{
			name: "alignment markers accepted",
			input: `| a | b |
|:---|---:|
| 1 | 2 |`,
			want: &Table{
				Headers: []string{"a", "b"},
				Rows:    [][]string{{"1", "2"}},
			},
		},
		{
			name:    "missing separator",
			input:   `| a | b |`,
			wantErr: true,
		},
		{
			name: "separator column count mismatch",
			input: `| a | b |
|---|
| 1 | 2 |`,
			wantErr: true,
		},
		{
			name: "row column count mismatch",
			input: `| a | b |
|---|---|
| 1 |`,
			wantErr: true,
		},
		{
			name: "invalid separator chars",
			input: `| a | b |
| x | y |
| 1 | 2 |`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MarkdownParser{}.Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestCSVParser(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Table
		wantErr bool
	}{
		{
			name:  "valid csv",
			input: "a,b\n1,2\n3,4\n",
			want: &Table{
				Headers: []string{"a", "b"},
				Rows:    [][]string{{"1", "2"}, {"3", "4"}},
			},
		},
		{
			name:  "quoted commas",
			input: "a,b\n\"1,5\",2\n",
			want: &Table{
				Headers: []string{"a", "b"},
				Rows:    [][]string{{"1,5", "2"}},
			},
		},
		{
			name:    "empty",
			input:   "",
			wantErr: true,
		},
		{
			name:    "ragged rows",
			input:   "a,b\n1\n",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CSVParser{}.Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestMarkdownFormatter(t *testing.T) {
	tests := []struct {
		name  string
		table *Table
		want  string
	}{
		{
			name: "pads to widest cell",
			table: &Table{
				Headers: []string{"a", "b"},
				Rows:    [][]string{{"1", "22"}, {"333", "4"}},
			},
			want: "| a   | b  |\n|-----|----|\n| 1   | 22 |\n| 333 | 4  |\n",
		},
		{
			name: "header is widest",
			table: &Table{
				Headers: []string{"name", "id"},
				Rows:    [][]string{{"a", "1"}},
			},
			want: "| name | id |\n|------|----|\n| a    | 1  |\n",
		},
		{
			name: "no rows",
			table: &Table{
				Headers: []string{"a", "b"},
				Rows:    nil,
			},
			want: "| a | b |\n|---|---|\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MarkdownFormatter{}.Format(tt.table)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got:\n%q\nwant:\n%q", got, tt.want)
			}
		})
	}
}

func TestCSVFormatter(t *testing.T) {
	tests := []struct {
		name  string
		table *Table
		want  string
	}{
		{
			name: "simple",
			table: &Table{
				Headers: []string{"a", "b"},
				Rows:    [][]string{{"1", "2"}},
			},
			want: "a,b\n1,2\n",
		},
		{
			name: "quotes commas",
			table: &Table{
				Headers: []string{"a", "b"},
				Rows:    [][]string{{"1,5", "2"}},
			},
			want: "a,b\n\"1,5\",2\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CSVFormatter{}.Format(tt.table)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolveConfig(t *testing.T) {
	tests := []struct {
		name        string
		from        string
		to          string
		input       string
		output      string
		wantFrom    string
		wantTo      string
		wantErr     bool
		errContains string
	}{
		{
			name: "explicit flags only",
			from: "md", to: "csv",
			wantFrom: "md", wantTo: "csv",
		},
		{
			name:  "both inferred from extensions",
			input: "in.md", output: "out.csv",
			wantFrom: "md", wantTo: "csv",
		},
		{
			name: "flag wins over extension",
			from: "csv", to: "md",
			input: "in.md", output: "out.csv",
			wantFrom: "csv", wantTo: "md",
		},
		{
			name:  "stdin pipe with -o for output inference",
			from:  "md",
			input: "-", output: "out.csv",
			wantFrom: "md", wantTo: "csv",
		},
		{
			name:  "input file with -t for output (stdout)",
			to:    "csv",
			input: "in.md",
			wantFrom: "md", wantTo: "csv",
		},
		{
			name: "md to md is allowed",
			from: "md", to: "md",
			wantFrom: "md", wantTo: "md",
		},
		{
			name:        "stdin without -f errors",
			to:          "csv",
			wantErr:     true,
			errContains: "input format",
		},
		{
			name:        "stdout without -t and no -o errors",
			from:        "md",
			input:       "in.md",
			wantErr:     true,
			errContains: "output format",
		},
		{
			name:        "unknown input extension errors",
			input:       "data.txt",
			to:          "csv",
			wantErr:     true,
			errContains: "cannot infer input format",
		},
		{
			name:        "unknown output extension errors",
			from:        "md",
			output:      "data.txt",
			wantErr:     true,
			errContains: "cannot infer output format",
		},
		{
			name:        "unknown format value errors",
			from:        "json",
			to:          "csv",
			wantErr:     true,
			errContains: "unknown input format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := resolveConfig(tt.from, tt.to, tt.input, tt.output)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errContains)
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("err = %v, want substring %q", err, tt.errContains)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if cfg.From != tt.wantFrom {
				t.Errorf("From = %q, want %q", cfg.From, tt.wantFrom)
			}
			if cfg.To != tt.wantTo {
				t.Errorf("To = %q, want %q", cfg.To, tt.wantTo)
			}
		})
	}
}

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    Config
		wantErr bool
	}{
		{
			name: "short flags",
			args: []string{"-f", "md", "-t", "csv"},
			want: Config{From: "md", To: "csv"},
		},
		{
			name: "long flags",
			args: []string{"--from", "md", "--to", "csv"},
			want: Config{From: "md", To: "csv"},
		},
		{
			name: "positional file with extension inference",
			args: []string{"-t", "csv", "in.md"},
			want: Config{From: "md", To: "csv", Input: "in.md"},
		},
		{
			name: "output flag with inference",
			args: []string{"-f", "md", "-o", "out.csv"},
			want: Config{From: "md", To: "csv", Output: "out.csv"},
		},
		{
			name: "positional file before flags",
			args: []string{"in.md", "-o", "out.csv"},
			want: Config{From: "md", To: "csv", Input: "in.md", Output: "out.csv"},
		},
		{
			name: "flags around positional file",
			args: []string{"-f", "md", "in.md", "-o", "out.csv"},
			want: Config{From: "md", To: "csv", Input: "in.md", Output: "out.csv"},
		},
		{
			name:    "two positional files errors",
			args:    []string{"-f", "md", "-t", "csv", "a.md", "b.md"},
			wantErr: true,
		},
		{
			name:    "unknown flag errors",
			args:    []string{"--bogus"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseFlags(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestRunRoundTrip(t *testing.T) {
	cfg, err := resolveConfig("md", "csv", "-", "-")
	if err != nil {
		t.Fatal(err)
	}
	in := strings.NewReader("| a | b |\n|---|---|\n| 1 | 2 |\n")
	var out strings.Builder
	if err := run(cfg, in, &out); err != nil {
		t.Fatal(err)
	}
	want := "a,b\n1,2\n"
	if out.String() != want {
		t.Errorf("got %q, want %q", out.String(), want)
	}
}

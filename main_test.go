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
		input       string
		output      string
		reformat    bool
		want        Config
		wantErr     bool
		errContains string
	}{
		{
			name:     "reformat md keeps md",
			input:    "in.md",
			reformat: true,
			want:     Config{From: "md", To: "md", Input: "in.md", Reformat: true},
		},
		{
			name:     "reformat csv keeps csv",
			input:    "in.csv",
			reformat: true,
			want:     Config{From: "csv", To: "csv", Input: "in.csv", Reformat: true},
		},
		{
			name:  "reformat with matching output extension",
			input: "in.md", output: "out.md",
			reformat: true,
			want:     Config{From: "md", To: "md", Input: "in.md", Output: "out.md", Reformat: true},
		},
		{
			name:  "reformat with conflicting output extension errors",
			input: "in.md", output: "out.csv",
			reformat:    true,
			wantErr:     true,
			errContains: "--reformat keeps the md format",
		},
		{
			name:     "reformat on stdin defers to run",
			reformat: true,
			want:     Config{Reformat: true},
		},
		{
			name:  "md file converts to csv",
			input: "in.md",
			want:  Config{From: "md", To: "csv", Input: "in.md"},
		},
		{
			name:  "csv file converts to md",
			input: "in.csv",
			want:  Config{From: "csv", To: "md", Input: "in.csv"},
		},
		{
			name: "stdin defers format to run",
			want: Config{},
		},
		{
			name:  "explicit stdin alias defers format to run",
			input: "-",
			want:  Config{Input: "-"},
		},
		{
			name:   "stdin with output extension",
			output: "out.md",
			want:   Config{To: "md", Output: "out.md"},
		},
		{
			name:  "output extension wins over counterpart",
			input: "in.md", output: "out.md",
			want: Config{From: "md", To: "md", Input: "in.md", Output: "out.md"},
		},
		{
			name:  "stdout alias keeps counterpart",
			input: "in.csv", output: "-",
			want: Config{From: "csv", To: "md", Input: "in.csv", Output: "-"},
		},
		{
			name:        "unknown input extension errors",
			input:       "data.txt",
			wantErr:     true,
			errContains: "cannot infer input format",
		},
		{
			name:        "unknown output extension errors",
			input:       "in.md",
			output:      "data.txt",
			wantErr:     true,
			errContains: "cannot infer output format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveConfig(tt.input, tt.output, tt.reformat)
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
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
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
			name: "no args reads stdin",
			args: nil,
			want: Config{},
		},
		{
			name: "positional file alone",
			args: []string{"in.md"},
			want: Config{From: "md", To: "csv", Input: "in.md"},
		},
		{
			name: "short output flag",
			args: []string{"-o", "out.md"},
			want: Config{To: "md", Output: "out.md"},
		},
		{
			name: "long output flag",
			args: []string{"--output", "out.csv"},
			want: Config{To: "csv", Output: "out.csv"},
		},
		{
			name: "positional file before flags",
			args: []string{"in.md", "-o", "out.csv"},
			want: Config{From: "md", To: "csv", Input: "in.md", Output: "out.csv"},
		},
		{
			name: "flags around positional file",
			args: []string{"-o", "out.csv", "in.md", "-o", "out.md"},
			want: Config{From: "md", To: "md", Input: "in.md", Output: "out.md"},
		},
		{
			name: "short reformat flag",
			args: []string{"-r", "in.md"},
			want: Config{From: "md", To: "md", Input: "in.md", Reformat: true},
		},
		{
			name: "long reformat flag after positional",
			args: []string{"in.csv", "--reformat"},
			want: Config{From: "csv", To: "csv", Input: "in.csv", Reformat: true},
		},
		{
			name:    "reformat with conflicting output extension errors",
			args:    []string{"-r", "in.md", "-o", "out.csv"},
			wantErr: true,
		},
		{
			name:    "two positional files errors",
			args:    []string{"a.md", "b.md"},
			wantErr: true,
		},
		{
			name:    "removed -f flag errors",
			args:    []string{"-f", "md"},
			wantErr: true,
		},
		{
			name:    "removed -t flag errors",
			args:    []string{"-t", "csv", "in.md"},
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

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{"markdown table", "| a | b |\n|---|---|\n", "md"},
		{"leading blank lines", "\n\n  | a |\n|---|\n", "md"},
		{"csv", "a,b\n1,2\n", "csv"},
		{"csv cell containing pipe later", "a,b\n|x|,2\n", "csv"},
		{"empty", "", "csv"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectFormat(tt.content); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTargetFormat(t *testing.T) {
	tests := []struct {
		name      string
		from      string
		outputExt string
		reformat  bool
		want      string
		wantErr   bool
	}{
		{name: "md counterpart", from: "md", want: "csv"},
		{name: "csv counterpart", from: "csv", want: "md"},
		{name: "output extension wins", from: "md", outputExt: "md", want: "md"},
		{name: "reformat keeps md", from: "md", reformat: true, want: "md"},
		{name: "reformat keeps csv", from: "csv", reformat: true, want: "csv"},
		{name: "reformat agrees with output extension", from: "csv", outputExt: "csv", reformat: true, want: "csv"},
		{name: "reformat conflicts with output extension", from: "csv", outputExt: "md", reformat: true, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := targetFormat(tt.from, tt.outputExt, tt.reformat)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRunStdin(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		reformat bool
		in       string
		want     string
		wantErr  bool
	}{
		{
			name:     "reformat aligns md columns",
			reformat: true,
			in:       "|a|bb|\n|-|-|\n|111|2|\n",
			want:     "| a   | bb |\n|-----|----|\n| 111 | 2  |\n",
		},
		{
			name:     "reformat normalizes csv quoting and line endings",
			reformat: true,
			in:       "\"a\",\"b\"\r\n\"1,5\",2\r\n",
			want:     "a,b\n\"1,5\",2\n",
		},
		{
			name:     "reformat rejects ragged csv",
			reformat: true,
			in:       "a,b\n1\n",
			wantErr:  true,
		},
		{
			name:     "reformat with conflicting output extension errors",
			output:   "out.md",
			reformat: true,
			in:       "a,b\n1,2\n",
			wantErr:  true,
		},
		{
			name: "md on stdin becomes csv",
			in:   "| a | b |\n|---|---|\n| 1 | 2 |\n",
			want: "a,b\n1,2\n",
		},
		{
			name: "csv on stdin becomes md",
			in:   "a,b\n1,2\n",
			want: "| a | b |\n|---|---|\n| 1 | 2 |\n",
		},
		{
			name:   "stdout alias keeps counterpart",
			output: "-",
			in:     "a,b\n1,2\n",
			want:   "| a | b |\n|---|---|\n| 1 | 2 |\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := resolveConfig("-", tt.output, tt.reformat)
			if err != nil {
				t.Fatal(err)
			}
			var out strings.Builder
			err = run(cfg, strings.NewReader(tt.in), &out)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if out.String() != tt.want {
				t.Errorf("got %q, want %q", out.String(), tt.want)
			}
		})
	}
}

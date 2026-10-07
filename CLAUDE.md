# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```sh
# Build
make build          # outputs to dist/mdcsv
go build -o dist/mdcsv ./main.go

# Test
go test ./...
go test -run TestMarkdownParser  # run a single test

# Format
make fmt            # gofmt -w on all .go files

# Install system-wide
make install        # copies dist/mdcsv to /usr/local/bin
```

## Architecture

All code lives in a single `main.go` (per DESIGN.md: stay there until a concrete reason — e.g. a third format — emerges).

**Core types:**
- `Table` — the canonical intermediate: `Headers []string`, `Rows [][]string`
- `Parser` interface: `Parse(string) (*Table, error)`
- `Formatter` interface: `Format(*Table) (string, error)`

**Registries** (keyed by format string `"md"` or `"csv"`):
- `parsers map[string]Parser` — `MarkdownParser`, `CSVParser`
- `formatters map[string]Formatter` — `MarkdownFormatter`, `CSVFormatter`

**Config resolution pipeline:** `parseFlags` → `resolveConfig` → `run`

Input format: explicit `-f` wins, then the input file extension, then error (stdin without `-f` is always an error). Output format: the `-o` file extension if given, otherwise the counterpart of the input format (`md`→`csv`, `csv`→`md`). There is no output-format flag.

**I/O:** input from stdin or a positional file arg; output to stdout or `-o FILE`. Both sides accept `-` as an explicit stdin/stdout alias.

## CLI surface

```
mdcsv [-f FORMAT] [-o FILE] [FILE]
```

Supported formats: `md`, `csv`. `md→md` (reformat/align columns) is reached by writing to a `.md` output path, e.g. `mdcsv messy.md -o clean.md`.

## License

Copyright (C) 2026 will-wright-eng

Licensed under the GNU General Public License, version 3 or (at your option) any
later version. See [LICENSE](LICENSE).

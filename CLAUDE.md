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

There are no format flags. Input format comes from the input file extension (`.md`/`.csv`, anything else errors); stdin is sniffed in `run` via `detectFormat` (content starting with `|` is `md`, otherwise `csv`). Output format is chosen by `targetFormat`: the `-o` file extension if given, otherwise the input format under `--reformat`, otherwise the counterpart (`md`→`csv`, `csv`→`md`). `--reformat` with a conflicting `-o` extension is an error. `Config.From`/`To` stay empty for stdin until `run` resolves them.

**I/O:** input from stdin or a positional file arg; output to stdout or `-o FILE`. Both sides accept `-` as an explicit stdin/stdout alias.

## CLI surface

```
mdcsv [-r] [-o FILE] [FILE]
```

Supported formats: `md`, `csv`. `-r`/`--reformat` keeps the input format: `md→md` aligns columns; `csv→csv` validates row shape and normalizes quoting and line endings via `encoding/csv`. Writing to an output path whose extension matches the input (e.g. `mdcsv messy.md -o clean.md`) does the same without the flag.

## License

Copyright (C) 2026 will-wright-eng

Licensed under the GNU General Public License, version 3 or (at your option) any
later version. See [LICENSE](LICENSE).

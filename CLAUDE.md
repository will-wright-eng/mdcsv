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

Format inference is per-side and priority-ordered: explicit flag wins, then file extension, then error. Stdin without `-f` and stdout without `-t`/`-o` are always errors.

**I/O:** input from stdin or a positional file arg; output to stdout or `-o FILE`. Both sides accept `-` as an explicit stdin/stdout alias.

## CLI surface

```
mdcsv [-f FROM] [-t TO] [-o FILE] [FILE]
```

Supported formats: `md`, `csv`. `md→md` is valid (reformats/aligns columns).

# mdcsv — CLI Redesign

## Motivation

The current CLI uses mutually-exclusive boolean flags (`--to-csv` / `--to-md`)
and required `--in` / `--out` path flags. This blocks the main use case for a
small format-conversion tool — being part of a Unix pipeline — and encodes one
piece of information (the conversion direction) across two flags.

This redesign aligns the tool with POSIX conventions (stdin/stdout by default,
single source of truth for format selection) and adds a third mode: reformat
a markdown table in place (md → md), which falls out naturally once `from`
and `to` are independent.

## Goals

- Pipe-friendly: read stdin, write stdout by default.
- Single, explicit way to express the conversion (`-f`, `-t`).
- Support md → md as a first-class mode (pretty-print / column-align).
- Keep the surface area small — one binary, no subcommands.

## Non-goals

- Multiple input files / batch mode.
- Format auto-detection from content (extension-based only).
- Streaming for tables larger than memory.
- `csv → csv` normalization. The registry technically supports it, but no
  custom quoting/whitespace handling beyond `encoding/csv` defaults is in
  scope for this phase.

## CLI surface

```
mdcsv [-f FROM] [-t TO] [-o FILE] [FILE]
```

| Flag             | Description                                          |
| ---------------- | ---------------------------------------------------- |
| `-f`, `--from`   | Input format: `md` or `csv`.                         |
| `-t`, `--to`     | Output format: `md` or `csv`.                        |
| `-o`, `--output` | Output file. Default or `-`: stdout.                 |
| `FILE`           | Input file. Default or `-`: stdin.                   |
| `-h`, `--help`   | Usage.                                               |

### Examples

```sh
# Pipe: md on stdin, csv on stdout
cat data.md | mdcsv -t csv > data.csv

# File in, file out, formats inferred from extensions
mdcsv data.md -o data.csv

# Reformat a markdown table in place (md -> md)
mdcsv -f md -t md messy.md -o clean.md

# Reformat via pipe — same format on both sides
cat messy.md | mdcsv -f md -t md
```

## Format inference

Each side (input, output) resolves independently from its own source, in
this priority order:

1. Explicit flag (`-f` for input, `-t` for output) — always wins.
2. File extension on that side (`FILE` for input, `-o` for output).
3. No source available (stdin without `-f`, stdout without `-t` and no
   `-o`) → error with usage message.

| Extension           | Format |
| ------------------- | ------ |
| `.md`               | `md`   |
| `.csv`              | `csv`  |
| anything else       | error — ask for explicit `-f` / `-t` |

This means inference is asymmetric and per-side: `cat in.md | mdcsv -o
out.csv` errors (stdin has no `-f`), but `cat in.md | mdcsv -f md -o
out.csv` works (`-t` inferred from `-o`). Likewise `mdcsv in.md > out.csv`
errors (stdout has no `-t`), but `mdcsv in.md -t csv > out.csv` works.

If neither `-f` nor `-t` can be resolved, exit non-zero with a usage
message naming which side is missing.

## md → md formatting

Same pipeline as the other modes: `parse → Table → format`. The markdown
formatter is the column-aligned writer that already exists for csv → md
(currently `CSVConverter.Format`). The mode is reached whenever `from == to == md`;
no special case in `main`.

Behavior:
- Pads each column to the widest cell (header or row).
- Normalizes the separator row to match column widths.
- Trims surrounding whitespace per cell.
- Preserves row order; does not sort or dedupe.

Out of scope for v1: alignment markers (`:---`, `---:`, `:---:`). The current
parser accepts them but the formatter emits a plain `---` separator. A follow-up
can preserve alignment if needed.

## Architecture

The existing `Converter` interface conflates two responsibilities:

```go
type Converter interface {
    Convert(input string) (*Table, error)   // parse FROM format
    Format(table *Table) (string, error)    // emit TO format
}
```

Split into two single-purpose interfaces so `from` and `to` are independent:

```go
type Parser interface {
    Parse(input string) (*Table, error)
}

type Formatter interface {
    Format(table *Table) (string, error)
}
```

Registry keyed by format name:

```go
parsers := map[string]Parser{
    "md":  &MarkdownParser{},
    "csv": &CSVParser{},
}
formatters := map[string]Formatter{
    "md":  &MarkdownFormatter{},
    "csv": &CSVFormatter{},
}
```

`main` resolves `from` / `to`, looks up the parser and formatter, and runs
`format(parse(input))`. md → md works without further changes.

## I/O

- Input: `os.Stdin` if `FILE` is absent or `-`, else `os.ReadFile(FILE)`.
- Output: `os.Stdout` if `-o` is absent or `-`, else
  `os.WriteFile(path, ..., 0644)`.
- Output ends with exactly one trailing newline (existing formatters
  already emit `\n` after the last row; don't double or strip).
- Errors go to `os.Stderr`; exit code `1` on any failure.
- Drop the trailing "Successfully converted…" line — it's noise on stdout
  and not idiomatic for filter tools.

## Breaking changes

| Old                              | New                              |
| -------------------------------- | -------------------------------- |
| `--to-csv` / `--to-md`           | `-f md -t csv` / `-f csv -t md`  |
| `--in PATH` (required)           | positional `PATH` (optional)     |
| `--out PATH` (auto-defaulted)    | `-o PATH` (default: stdout)      |
| Default output path inferred     | Removed — use `-o` or redirect   |

Repo has no external users yet, so a clean cut is preferred over a
compatibility shim.

## Code layout

Stay in a single `main.go`. The redesign roughly doubles the type count
(parsers + formatters + registry + format inference), but the total is
still small enough that splitting into multiple files would add navigation
cost without payoff. Revisit if a concrete reason emerges (e.g. a third
format).

## Tests

Add `main_test.go` with:

- Table-driven parser tests (`MarkdownParser`, `CSVParser`): valid input
  round-trips to the expected `Table`; malformed input returns an error.
- Table-driven formatter tests (`MarkdownFormatter`, `CSVFormatter`):
  given a `Table`, output matches the expected string byte-for-byte
  (column alignment matters for the markdown formatter).
- Format inference tests covering the priority rules in "Format
  inference" — flag wins over extension, each side independent, stdin
  without `-f` errors, etc.

Add a `make smoke` target that exercises the three in-scope mode
combinations end-to-end via shell pipelines (md→csv, csv→md, md→md),
asserting output matches a checked-in fixture. This catches wiring bugs
that unit tests miss (flag parsing, stdin/stdout paths, exit codes).

## Implementation plan

1. Split `Converter` into `Parser` / `Formatter`; register by format name.
2. Rewrite `parseFlags` for the new surface; add format inference.
3. Wire stdin/stdout I/O paths in `main`.
4. Remove `getDefaultOutputPath` and the success line.
5. Add `main_test.go` covering parsers, formatters, and inference.
6. Update `Makefile` — drop old invocations, add `make smoke`.

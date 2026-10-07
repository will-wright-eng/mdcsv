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
- Conversion direction is implied by the input format: `md` → `csv` and
  `csv` → `md`. No format flags at all.
- Support md → md as a first-class mode (pretty-print / column-align).
- Keep the surface area small — one binary, no subcommands.

## Non-goals

- Multiple input files / batch mode.
- Robust content-based format detection. Stdin uses a one-character
  sniff (`|` → markdown, else CSV); files are extension-based only.
- Streaming for tables larger than memory.
- `csv → csv` normalization. The registry technically supports it, but no
  custom quoting/whitespace handling beyond `encoding/csv` defaults is in
  scope for this phase.

## CLI surface

```
mdcsv [-o FILE] [FILE]
```

| Flag             | Description                                                |
| ---------------- | ---------------------------------------------------------- |
| `-o`, `--output` | Output file. Default or `-`: stdout.                       |
| `FILE`           | Input file. Default or `-`: stdin.                         |
| `-h`, `--help`   | Usage.                                                     |

There are no format flags. With only two formats the direction is
implied: markdown in, CSV out; CSV in, markdown out.

### Examples

```sh
# File in, csv on stdout
mdcsv data.md > data.csv

# Pipe: format sniffed from stdin content
cat data.md | mdcsv > data.csv

# File in, file out
mdcsv data.csv -o data.md

# Reformat a markdown table (md -> md): the .md output path overrides
# the default direction
mdcsv messy.md -o clean.md
```

## Format inference

Input format:

1. `FILE` extension, when a file is given.
2. Stdin: content sniff. Input whose first non-blank character is `|` is
   markdown; anything else is CSV. Deliberately naive — a CSV whose first
   cell starts with `|` is misdetected, and the markdown parser then
   reports the error.

Output format, in priority order:

1. `-o` file extension, when `-o` names a file.
2. The counterpart of the input format: `md` → `csv`, `csv` → `md`.

| Extension           | Format |
| ------------------- | ------ |
| `.md`               | `md`   |
| `.csv`              | `csv`  |
| anything else       | error — rename the file or use a `.md`/`.csv` path for `-o` |

Consequences: `mdcsv in.md > out.csv` and `cat in.md | mdcsv > out.csv`
both work with no flags. md → md through a pipe is not expressible;
reformatting requires a `.md` output path (`mdcsv messy.md -o clean.md`).

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
| `--to-csv` / `--to-md`           | implied by input format          |
| `--in PATH` (required)           | positional `PATH` (optional)     |
| `-f` / `-t` (interim redesign)   | removed — extension or stdin sniff |
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
- Format inference tests covering the rules in "Format inference" —
  extension sets the direction, `-o` extension overrides, stdin is
  sniffed, unknown extensions error, etc.

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

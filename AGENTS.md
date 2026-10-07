# AGENTS.md

Guidance for coding agents working in this repository.

## Commands

```sh
make build                      # dist/mdcsv, version stamped from git describe
make test                       # go test ./...
go test -run TestMarkdownParser # single test
make smoke                      # end-to-end checks against testdata/ fixtures
make fmt                        # gofmt -w .
make snapshot                   # goreleaser dry run into dist/, no publish
```

## Architecture

All code lives in `main.go`; stay there until a concrete reason (e.g. a third
format) emerges. `docs/DESIGN.md` is the design rationale.

- `Table{Headers, Rows}` is the intermediate. `Parser` and `Formatter` are
  single-method interfaces registered in the `parsers`/`formatters` maps keyed
  by `"md"` or `"csv"`. Every mode is `format(parse(input))`.
- Pipeline: `parseFlags` → `resolveConfig` → `run`. There are no format flags.
  Input format comes from the file extension; stdin is sniffed in `run` by
  `detectFormat` (leading `|` → md, else csv). `targetFormat` picks the output:
  `-o` extension, else input format under `--reformat`, else the counterpart.
- `-h`/`-v` are handled in `main` before flag parsing. `parseFlags` re-parses
  after each positional so `mdcsv in.md -o out.csv` works.

## Conventions

- Conventional commits; releases are cut by GoReleaser on a `v*` tag push.
- `*.csv` is gitignored except under `testdata/`. `make smoke` diffs CLI
  output against `testdata/simple.{md,csv}`, so keep those fixtures in sync
  with formatter changes.

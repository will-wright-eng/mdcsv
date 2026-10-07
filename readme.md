# mdcsv

Convert between markdown tables and CSV from the command line. Reads stdin,
writes stdout, and infers the direction from the input: `.md` in → CSV out,
`.csv` in → markdown out.

## Install

```sh
go install github.com/will-wright-eng/mdcsv@latest
```

Or download a binary from the [releases page](https://github.com/will-wright-eng/mdcsv/releases).

## Usage

```
mdcsv [-r] [-o FILE] [FILE]
```

```sh
mdcsv data.md > data.csv        # markdown → CSV
mdcsv data.csv -o data.md       # CSV → markdown
cat data.md | mdcsv             # stdin: leading '|' is markdown, else CSV
mdcsv -r messy.md               # reformat: align markdown columns
mdcsv -r data.csv               # reformat: validate and normalize CSV
```

| Flag               | Description                                              |
| ------------------ | -------------------------------------------------------- |
| `-r`, `--reformat` | Keep the input format instead of converting              |
| `-o`, `--output`   | Output file (default stdout); `.md`/`.csv` sets format   |
| `-v`, `--version`  | Print version                                            |

## Development

```sh
make build   # dist/mdcsv
make test    # unit tests
make smoke   # end-to-end checks against testdata/
```

See [AGENTS.md](AGENTS.md) for architecture notes and [docs/DESIGN.md](docs/DESIGN.md)
for the design rationale.

## License

GPL-3.0-or-later. See [LICENSE](LICENSE).

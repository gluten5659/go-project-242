### Hexlet tests and linter status

[![Actions Status](https://github.com/gluten5659/go-project-242/actions/workflows/hexlet-check.yml/badge.svg)](https://github.com/gluten5659/go-project-242/actions)

## Description

CLI tool that calculates file and directory sizes.

## Install

```
make build
```

## Usage

```
bin/hexlet-path-size [options] <path>
```

## Examples

```
bin/hexlet-path-size file.txt
5B      file.txt

bin/hexlet-path-size --human largefile.bin
2.0KB   largefile.bin

bin/hexlet-path-size -r src/
1024B   src/

bin/hexlet-path-size -r -a -H src/
1.5KB   src/
```

## Flags

- `--human, -H` converts bytes into readable format (KB, MB, GB)
- `--recursive, -r` includes subdirectories
- `--all, -a` includes hidden files

## Behavior

A few rules are worth calling out explicitly so the output never feels
surprising.

### Hidden files

By default a directory walk skips entries whose name starts with `.`. The
`-a` flag turns that filtering off and counts hidden entries as well.

The filter only kicks in while walking *into* a directory. If you point the
tool directly at a hidden path (for example `bin/hexlet-path-size .env`),
its size is always reported — the `-a` flag is about what gets included
during a walk, not about what you are allowed to ask for.

### Symbolic links

Symbolic links are reported by the size of the link entry itself (the
length of the target path stored in the link), not by the size of the file
they point to. The tool never follows symlinks, so a link inside a
directory contributes its own bytes to the total and that is all.

This keeps the result stable when links point outside the tree, are
broken, or form cycles.

### Exit codes

The tool returns distinct exit codes so it can be used from scripts:

| Code | Meaning                                          |
|------|--------------------------------------------------|
| 0    | success                                          |
| 1    | other / unexpected error                         |
| 64   | usage error (wrong number of arguments, bad flag)|
| 65   | unsupported file type (sockets, pipes, devices)  |
| 66   | path does not exist                              |
| 77   | permission denied                                |

## Development

```
make build
make test
make lint
```

Run tests with coverage:

```
go test -cover ./...
```

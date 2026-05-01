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

#!/usr/bin/env bash

prompt='$ '
type_delay="${TYPE_DELAY:-0.04}"
pause_after="${PAUSE:-1.2}"

sample_tree='demo-tree'

setup_sample_tree() {
	mkdir -p "$sample_tree"
	printf 'visible content\n' > "$sample_tree/file.txt"
	printf 'secret\n' > "$sample_tree/.hidden.txt"
}

cleanup_sample_tree() {
	rm -rf "$sample_tree"
}

type_and_run() {
	local command="$1"
	printf '%s' "$prompt"
	for (( index = 0; index < ${#command}; index++ )); do
		printf '%s' "${command:index:1}"
		sleep "$type_delay"
	done
	printf '\n'
	sleep 0.3
	eval "$command"
	printf '\n'
	sleep "$pause_after"
}

trap cleanup_sample_tree EXIT
setup_sample_tree

type_and_run 'bin/hexlet-path-size --help'
type_and_run 'bin/hexlet-path-size README.md'
type_and_run 'bin/hexlet-path-size --human README.md'
type_and_run 'bin/hexlet-path-size -r internal/'
type_and_run "ls -a $sample_tree"
type_and_run "bin/hexlet-path-size -r $sample_tree"
type_and_run "bin/hexlet-path-size -r -a $sample_tree"
type_and_run 'bin/hexlet-path-size no-such-file; echo "exit code: $?"'

.PHONY: build test lint fmt lint-fix demo

build:
	go build -o bin/hexlet-path-size ./cmd/hexlet-path-size
test:
	go test -race ./...
lint:
	go tool golangci-lint run
fmt:
	go tool golangci-lint fmt
lint-fix:
	go tool golangci-lint run --fix
demo: build
	asciinema record --headless --window-size 100x30 --overwrite -i 2 -c "bash demo/demo.sh" demo/demo.cast
	agg --speed 1.4 --idle-time-limit 1.5 demo/demo.cast demo/demo.gif
	rm -f demo/demo.cast

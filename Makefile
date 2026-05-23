.PHONY: build test lint fmt lint-fix

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

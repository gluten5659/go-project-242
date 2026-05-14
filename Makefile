.PHONY: build test lint

build:
	go build -o bin/hexlet-path-size ./cmd/hexlet-path-size
test:
	go test -race ./...
lint:
	go tool golangci-lint run

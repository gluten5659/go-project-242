build:
	go build -o bin/hexlet-path-size ./cmd/hexlet-path-size
test:
	go test ./...
lint:
	golangci-lint run

.PHONY: build run dev clean test lint fmt

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

dev:
	go run ./cmd/server

clean:
	rm -rf bin/ tmp/

test:
	go test ./... -v -count=1

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .

.PHONY: build run generate test

build: generate
	@go build -v -o 2sp ./cmd/2sp

build-all: generate
	@go build -v ./...

run: generate
	@go run ./cmd/2sp

generate:
	@go generate ./...

test: generate
	@gotestsum

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

demo:
	@go run ./cmd/2sp --anonymous --demo --name=Alice

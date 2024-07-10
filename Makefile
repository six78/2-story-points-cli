.PHONY: build run generate test

build:
	@go build -v -o 2sp ./cmd/2sp

build-all:
	@go build -v ./...

run:
	@go run -buildvcs=true ./cmd/2sp

generate:
	@go generate ./...

test: generate
	@gotestsum

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

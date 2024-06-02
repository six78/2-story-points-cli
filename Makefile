.PHONY: build run generate test

build:
	@go build -v -o 2sp ./cmd/2sp/main.go

build-all:
	@go build -v ./...

run:
	@go run cmd/2sp/main.go

generate:
	@go generate ./...

test: generate
	@gotestsum
.PHONY: build test

setup:
	go get ./...

build:
	rm -rf build
	go build -o build/cli cmd/cli/*.go

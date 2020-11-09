.PHONY: build test

setup:
	go get ./...

build:
	rm -rf build
	go build -o build/cli cmd/cli/*.go

#
# Utility targets for testing out the cli.

dev.run.change-in:
	make build && ./build/cli evaluate change-in --input "hello.yml" --output "hello.yml.compiled" --logs "logs.jsonl"

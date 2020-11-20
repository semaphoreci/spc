.PHONY: build test

setup:
	go get ./...

build:
	rm -rf build
	go build -o build/cli cmd/cli/*.go

gen.pipeline.models:
	#
	# schema-generate can only use JSON files, converting it with Ruby
	#   Input: schemas/v1.0.yml
	#  Output: /tmp/v1.0.json
	#
	ruby -rjson -ryaml -e "File.write('/tmp/v1.0.json', JSON.pretty_generate(YAML.load_file('schemas/v1.0.yml')))"
	go get -u github.com/a-h/generate/...
	schema-generate /tmp/v1.0.json > pkg/pipelines/models.go
	sed -i 's/^package main/package pipelines/' pkg/pipelines/models.go

#
# Utility targets for testing out the cli.

dev.run.change-in:
	make build && ./build/cli evaluate change-in --input "test/fixtures/hello.yml" --output "/tmp/hello.yml.compiled" --logs "/tmp/logs.jsonl"

test:
	gotestsum --format short-verbose

e2e: build
	ruby test/e2e/change_in_simple.rb
	ruby test/e2e/change_in_with_default_branch.rb
	ruby test/e2e/change_in_multiple_paths.rb
	ruby test/e2e/change_in_missing_branch.rb

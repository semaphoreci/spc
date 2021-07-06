.PHONY: build test

lint:
	revive -formatter friendly -config lint.toml ./...

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
	make build && ./build/cli compile --input "test/fixtures/hello.yml" --output "/tmp/hello.yml.compiled" --logs "/tmp/logs.jsonl"

test:
	gotestsum --format short-verbose

e2e: build
	ruby $(TEST)

#
# Automation of CLI tagging.
#
# When a tag is release, a new release will appear on Github.
#
tag.major:
	git fetch --tags
	latest=$$(git tag | sort --version-sort | tail -n 1); new=$$(echo $$latest | cut -c 2- | awk -F '.' '{ print "v" $$1+1 ".0.0" }');          echo $$new; git tag $$new; git push origin $$new

tag.minor:
	git fetch --tags
	latest=$$(git tag | sort --version-sort | tail -n 1); new=$$(echo $$latest | cut -c 2- | awk -F '.' '{ print "v" $$1 "." $$2 + 1 ".0" }');  echo $$new; git tag $$new; git push origin $$new

tag.patch:
	git fetch --tags
	latest=$$(git tag | sort --version-sort | tail -n 1); new=$$(echo $$latest | cut -c 2- | awk -F '.' '{ print "v" $$1 "." $$2 "." $$3+1 }'); echo $$new; git tag $$new; git push origin $$new

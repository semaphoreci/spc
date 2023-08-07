.PHONY: build test

SECURITY_TOOLBOX_BRANCH ?= master
SECURITY_TOOLBOX_TMP_DIR ?= /tmp/security-toolbox

check.prepare:
	rm -rf $(SECURITY_TOOLBOX_TMP_DIR)
	git clone git@github.com:renderedtext/security-toolbox.git $(SECURITY_TOOLBOX_TMP_DIR) && (cd $(SECURITY_TOOLBOX_TMP_DIR) && git checkout $(SECURITY_TOOLBOX_BRANCH) && cd -)

check.static: check.prepare
	docker run -it -v $$(pwd):/app \
		-v $(SECURITY_TOOLBOX_TMP_DIR):$(SECURITY_TOOLBOX_TMP_DIR) \
		registry.semaphoreci.com/ruby:2.7 \
		bash -c 'cd /app && $(SECURITY_TOOLBOX_TMP_DIR)/code --language go -d'

check.deps: check.prepare
	docker run -it -v $$(pwd):/app \
		-v $(SECURITY_TOOLBOX_TMP_DIR):$(SECURITY_TOOLBOX_TMP_DIR) \
		registry.semaphoreci.com/ruby:2.7 \
		bash -c 'cd /app && $(SECURITY_TOOLBOX_TMP_DIR)/dependencies --language go -d'


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

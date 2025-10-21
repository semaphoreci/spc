# Repository Guidelines

## Project Structure & Module Organization
Semaphore Pipeline Compiler (SPC) is a Go CLI. CLI entrypoint lives in `cmd/cli/main.go`. Core packages sit under `pkg/` (e.g., `pkg/commands`, `pkg/pipelines`, `pkg/when`) and share test doubles in `pkg/.../*_test.go`. Shared YAML schemas are in `schemas/` and feed generated models in `pkg/pipelines/models.go`. The `test/` tree stores fixtures (`test/fixtures/...`) and Ruby helpers for e2e runs, while compiled artifacts are written to `build/`.

For quick triage or deeper architectural context, consult `DOCUMENTATION.md`—it maps CLI entry points, pipeline passes, and common task recipes.

## Build, Test, and Development Commands
- `make setup` installs Go module dependencies.
- `make build` or `go build ./...` creates `build/cli`.
- `make dev.run.change-in` rebuilds then exercises the sample pipeline in `test/fixtures/hello.yml`.
- `make lint` runs Revive with `lint.toml`.
- `make test` executes `gotestsum --format short-verbose`.
- `make gen.pipeline.models` regenerates models after updating `schemas/v1.0.yml`.

## Coding Style & Naming Conventions
Format Go code with `gofmt` (tabs for indentation, trailing newlines) and keep imports curated via `goimports` if available. Revive rules in `lint.toml` enforce receiver naming, error handling, and comment quality—run lint before pushing. Package names stay lowercase, files use snake_case, and exported identifiers should have doc comments connecting them to CLI behaviour.

## Testing Guidelines
Unit tests live alongside code as `*_test.go` and rely on `gotest.tools/v3` assertions. Favor table-driven cases and mirror production package names. Use fixtures from `test/fixtures/` to cover YAML edge cases. Run `make test` locally before opening a PR. For manual regression checks, run `make dev.run.change-in` or target the Ruby harness with `make e2e TEST=test/e2e.rb`.

## Commit & Pull Request Guidelines
Recent history (`git log`) shows Conventional-style prefixes (`fix:`, `feat:`, `chore:`) for clarity—continue that pattern and keep subjects under 72 characters. Reference issues or PRs with `(#id)` when applicable. Pull requests should summarise intent, list validation steps (`make test`, `make lint`), attach relevant pipeline outputs or screenshots, and link to Semaphore change requests.

## Security & Release Tips
Security checks rely on the internal toolbox; run `make check.static` and `make check.deps` when touching dependencies or shell execution paths. Version bumps are automated: use `make tag.patch|minor|major` from a clean main branch, which triggers GoReleaser via Semaphore. Never commit secrets; configuration belongs in environment variables or the security toolbox.

# SPC Architecture Cheat Sheet

## Mission & Entry Points
This repository builds the Semaphore Pipeline Compiler (`spc`), a Go CLI that normalizes pipeline YAML before Semaphore runs it. The binary lives under `cmd/cli/main.go` and bootstraps the Cobra command tree in `pkg/cli`. First-level commands include `spc compile`, `spc evaluate change-in`, and `spc list-diff`, each orchestrating pipeline transforms or Git queries.

## Pipeline Processing Flow
`pkg/pipelines` wraps the raw YAML (loaded through `LoadFromFile`) in a `Pipeline` struct backed by `github.com/Jeffail/gabs`. The default compilation flow called from `spc compile` performs three passes:
1. **Commands file expansion** via `ExtractCommandsFromCommandsFiles` which scans YAML for `commands_file` references and uses `pkg/commands.File` to inline shell commands (supporting relative or repo-absolute paths).
2. **Template evaluation** with `EvaluateTemplates`, which walks the document, discovers `{{ expression }}` placeholders, and delegates substitution to `pkg/templates` utilities while logging progress through `pkg/consolelogger`.
3. **`change_in` evaluation** through `EvaluateChangeIns`, reusing the `pkg/when` parser binary (`whencli`) plus helpers under `pkg/when/changein` to decide whether blocks/promotions should remain.
The resulting structure can be emitted back to JSON (`ToJSON`) or YAML (`ToYAML`).

## Supporting Packages
- `pkg/cli`: Cobra command definitions, flag helpers (`util.go`), and shared error handling that exits on known `pkg/logs` errors.
- `pkg/logs`: central logging wiring for compiler output files; carries typed errors such as `ErrorChangeInMissingBranch`.
- `pkg/git`: wrappers over Git commands (`diff_set.go`, `git.go`) to compute commit ranges and run `git fetch` / `git diff --name-only`. Used by list-diff and the change_in evaluator.
- `pkg/environment`: adapters around Semaphore environment variables with local fallbacks (e.g., resolves current branch, repo slug, commit range).
- `pkg/templates`, `pkg/when`: pure-Go helpers to detect and evaluate template expressions and `when` syntax; the latter shells out to the external parser.
- `pkg/consolelogger`: indentation-aware stdout writer that gives numbered logs during template and change_in passes.

## Data & Schemas
Shared pipeline schema definitions sit in `schemas/` (currently `v1.0.yml`). Regenerate the strongly typed Go model in `pkg/pipelines/models.go` with `make gen.pipeline.models`, which converts YAML to JSON (Ruby helper) and pipes it through `schema-generate`.

## Tests & Fixtures
Unit tests live alongside code (`*_test.go`) and rely on `gotest.tools/v3`. YAML fixtures are under `test/fixtures/`; Ruby-driven end-to-end tests reside in `test/e2e` with a harness in `test/e2e.rb`. Run `make test` for Go coverage, `make dev.run.change-in` to compile the sample hello pipeline, and `make e2e TEST=test/e2e.rb` for smoke testing.

## Build & Release Tooling
Key Make targets:
- `make setup`: pull Go module dependencies.
- `make build`: compile `build/cli`.
- `make lint`: run Revive using `lint.toml`.
- `make check.static`/`make check.deps`: invoke the security toolbox Docker image.
- `make tag.(patch|minor|major)`: automate semantic version bumps and push release tags (Semaphore + GoReleaser handle artifacts).

## Common Task Recipes
- **Add a CLI command**: create a file in `pkg/cli`, attach it inside an `init()` function, and expose the handler from existing packages or a new `pkg/...` domain module.
- **Inject a new pipeline transform**: extend `pkg/pipelines.Pipeline` with a method, call it from `compileCmd` in the desired sequence, and provide fixtures under `test/fixtures/` with table-driven tests.
- **Debug change_in issues**: ensure `when` binary is on `$PATH` (`checkWhenInstalled` exits if missing), run `spc list-diff` to verify commit range inputs, and inspect logs written via `--logs`.

## Operational Notes
- All repo paths resolve relative to the Git root; absolute `commands_file` values are treated as `/path/from/repo/root`.
- Exit behaviour differs for known semantic errors: `ErrorChangeInMissingBranch` and `ErrorInvalidWhenExpression` exit gracefully, everything else panics to highlight unexpected states.
- The CLI assumes Go 1.20+ (see `go.mod`) and uses modules exclusively; no vendoring is committed.
- Keep an eye on `pkg/consolelogger` output when adding evaluators—the logs form part of the artefacts consumed by Semaphore users.

# Semaphore Pipeline Compiler (SPC)

Tooling for compiling and evaluating pipelines on Semaphore 2.0.

# Release Process

Releases are built by Sempahore for every git tag on Github. To initialize the
release process:

1. go to the project root
2. Run `make tag.patch` (or `make tag.minor`, or `make tag.major`) to bump and push a new tag to Github
3. Semaphore will take over, and execute the `.semaphore/release.yml`.

On Sempahore, we use GoReleaser to create releases. Its configuration is stored
in `.goreleaser.yml`.

# TODOs

- [x] Green tests on CI
- [x] Compile and make release on Github

### Change In Evaluation

##### Path types

- [x] Evaluate Change In for absolute paths
- [x] Evaluate Change In for relative paths
- [x] Evaluate Change In for glob expressions

##### Location

- [x] Evaluate Change In blocks/when/skip
- [x] Evaluate Change In blocks/when/run
- [x] Evaluate Change In promotions/auto_promote/when
- [x] Evaluate Change In fail_fast/cancel/when/skip
- [x] Evaluate Change In fail_fast/stop/when/skip
- [x] Evaluate Change In auto_cancel/queued/when/skip
- [x] Evaluate Change In auto_cancel/running/when/skip
- [x] Evaluate Change In queue/when
- [x] Evaluate Change In priority/when

##### Parameters

- [x] Default branch for change_in
- [x] Exclude patterns from the match
- [x] Toggle pipeline file tracking for change_in
- [x] Change branch range for change_in
- [x] Change default range for change_in
- [x] Change on_tags value for change_in

##### Errors

- [x] Default branch does not exists

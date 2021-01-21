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

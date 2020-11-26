# Semaphore Pipeline Compiler (SPC)

Tooling for compiling and evaluating pipelines on Semaphore 2.0.

# TODOs

- [x] Green tests on CI
- [ ] Compile and make release on Github

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
- [ ] Toggle pipeline file tracking for change_in
- [ ] Change branch range for change_in
- [ ] Change default range for change_in
- [ ] Toggle on_tags toggle for change_in

##### Errors

- [ ] Default branch does not exists

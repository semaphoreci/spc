# rubocop:disable all

#
# A pull request ignores branch_range and diffs the target branch against the
# PR branch. A branch_range naming a base commit, such as
# $SEMAPHORE_GIT_COMMIT_RANGE, must not stop spc from fetching both of them:
# the PR branch is absent from the local clone and git merge-base fails.

require_relative "../e2e"
require 'yaml'

pipeline = %{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: Test
    run:
      when: "change_in('/app', {branch_range: '$SEMAPHORE_GIT_COMMIT_RANGE'})"

  - name: Test2
    run:
      when: "change_in('/lib', {branch_range: '$SEMAPHORE_GIT_COMMIT_RANGE'})"
}

origin = TestRepoForChangeIn.setup()

origin.add_file('.semaphore/semaphore.yml', pipeline)
origin.commit!("Bootstrap")

origin.create_branch("dev")
origin.add_file("app/a.yml", "hello")
origin.commit!("Bootstrap app")

origin.switch_branch("master")
origin.add_file("lib/b.yml", "world")
origin.commit!("Bootstrap lib")

origin.merge_branch("dev")

repo = origin.clone_local_copy(branch: "master")

origin.run(%{
  git reset --hard HEAD~1
})

repo.run(%{
  export SEMAPHORE_GIT_SHA=$(git rev-parse HEAD)

  git reset --hard HEAD~1

  git checkout $SEMAPHORE_GIT_SHA

  export SEMAPHORE_GIT_REF_TYPE=pull-request
  export SEMAPHORE_GIT_BRANCH=master
  export SEMAPHORE_GIT_PR_BRANCH=dev

  #{spc} compile \
     --input .semaphore/semaphore.yml \
     --output /tmp/output.yml \
     --logs /tmp/logs.yml
})

assert_eq(YAML.load_file('/tmp/output.yml'), YAML.load(%{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: Test
    run:
      when: "true"

  - name: Test2
    run:
      when: "false"
}))

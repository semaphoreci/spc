# rubocop:disable all

#
# In pipelines triggered by PRs Semaphore checkouts the merge commit as a
# detached head. In cases where PR is targeting non-master branch, there are
# several tricky parts to test:
# - actual cloned repo should only have master branch and a detached merged commit
# - merge commit has changes on both the target and the PR branch,
#   and we want only the PR ones
# - we need to fetch both the target and the PR branch
#
# This test simulates this state by creating repo with two non-master branches,
# merging them and reseting HEAD of target branch back by one so merge commit
# becomes a detached head when it is checked out and then deleting the target
# branch.
#

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
      when: "branch = 'master' and change_in('/app')"

  - name: Test2
    run:
      when: "branch = 'master' and change_in('/test')"

  - name: Test3
    run:
      when: "branch = 'master' and change_in('/lib')"
}

origin = TestRepoForChangeIn.setup()

origin.add_file('.semaphore/semaphore.yml', pipeline)
origin.commit!("Bootstrap")

origin.create_branch("target-branch")
origin.add_file("app/a.yml", "foo")
origin.commit!("Bootstrap app")

origin.create_branch("dev")
origin.add_file("test/a.test", "bar")
origin.commit!("Bootstrap tests")

origin.switch_branch("target-branch")
origin.add_file("lib/b.yml", "baz")
origin.commit!("Bootstrap lib")

origin.merge_branch("dev")

# create a temp branch that will be fetched to get merge commit in local repo
origin.create_branch("temp-branch-for-fetching-merge-commit")

# delete merge commit from the target-branch
origin.switch_branch("target-branch")
origin.run(%{
  git reset --hard HEAD~1
})

repo = origin.clone_local_copy(branch: "master")

repo.run(%{
  git fetch origin temp-branch-for-fetching-merge-commit

  git branch temp-branch-for-fetching-merge-commit FETCH_HEAD

  git checkout temp-branch-for-fetching-merge-commit

  export SEMAPHORE_GIT_SHA=$(git rev-parse HEAD)

  git reset --hard HEAD~1

  git checkout $SEMAPHORE_GIT_SHA

  git branch -D temp-branch-for-fetching-merge-commit

  export SEMAPHORE_GIT_REF_TYPE=pull-request
  export SEMAPHORE_GIT_BRANCH=target-branch
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
      when: "(branch = 'master') and false"

  - name: Test2
    run:
      when: "(branch = 'master') and true"

  - name: Test3
    run:
      when: "(branch = 'master') and false"
}))

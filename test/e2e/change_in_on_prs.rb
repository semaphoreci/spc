# rubocop:disable all

#
# In pipelines triggered by PRs Semaphore checkouts the merge commit as a
# detached head. That makes evaluating change_in tricky because merge commit
# includes changes made to target branch  after the branch that is the source of
# the PR diverged from targeted branch.
#
# In order to get proper changeset we need to fetch both source and target
# branches, since target might not be master.
# After that we can use "<target_branch>....<pr_source_branch>" as a commit
# range for change_in since it does not include changes made to target branch
# after the source branch diverged.
#
# This test simulates this state by creating repo with two branches, merging
# them and reseting HEAD of target branch back by one so merge commit becomes
# a detached head when it is checked out.
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
      when: "branch = 'master' and change_in('/lib')"
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

  #{spc} evaluate change-in \
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
      when: "(branch = 'master') and true"

  - name: Test2
    run:
      when: "(branch = 'master') and false"
}))

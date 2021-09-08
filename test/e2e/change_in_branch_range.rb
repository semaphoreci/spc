# rubocop:disable all

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
      when: "branch = 'master' and change_in('/lib', {branch_range: '$SEMAPHORE_MERGE_BASE...$SEMAPHORE_GIT_SHA'})"

  - name: Test3
    run:
      when: "branch = 'master' and change_in('/app', {branch_range: 'dev...$SEMAPHORE_GIT_SHA'})"

  - name: Test4
    run:
      when: "branch = 'master' and change_in(['/lib', 'log.txt'], {branch_range: 'dev...$SEMAPHORE_GIT_SHA'})"

  - name: Test5
    run:
      when: "branch = 'master' and change_in(['/lib'], {branch_range: '$SEMAPHORE_GIT_SHA^...$SEMAPHORE_GIT_SHA'})"
}

origin = TestRepoForChangeIn.setup()

origin.add_file('.semaphore/semaphore.yml', pipeline)
origin.commit!("Bootstrap")

origin.create_branch("dev")
origin.add_file("app/a.yml", "hello")
origin.commit!("Bootstrap app")

origin.create_branch("feature-1")
origin.add_file("lib/a.yml", "hello")
origin.commit!("Bootstrap lib")

repo = origin.clone_local_copy(branch: "feature-1")
repo.run(%{
  export SEMAPHORE_GIT_SHA=$(git rev-parse HEAD)
  export SEMAPHORE_GIT_BRANCH=feature-1

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
      when: "(branch = 'master') and true"

  - name: Test2
    run:
      when: "(branch = 'master') and true"

  - name: Test3
    run:
      when: "(branch = 'master') and false"

  - name: Test4
    run:
      when: "(branch = 'master') and true"

  - name: Test5
    run:
      when: "(branch = 'master') and true"
}))

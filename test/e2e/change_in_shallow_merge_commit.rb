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
  - name: ChangeIn
    run:
      when: "change_in('/lib')"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"
}

origin = TestRepoForChangeIn.setup()

origin.add_file('.semaphore/semaphore.yml', pipeline)
origin.commit!("Bootstrap")

origin.add_file("lib/base.txt", "base")
origin.commit!("Base change on master")

origin.create_branch("enterprise")

origin.switch_branch("master")
origin.add_file("lib/upstream.txt", "upstream change")
origin.commit!("Upstream change on master")

origin.switch_branch("enterprise")
origin.add_file("lib/upstream.txt", "upstream change")
origin.commit!("Cherry-picked upstream change on enterprise")

origin.run("git merge master --strategy ours --no-edit")

repo = origin.clone_local_copy(branch: "enterprise", depth: 1, single_branch: true)

repo.run(%{
  export SEMAPHORE_GIT_SHA=$(git rev-parse HEAD)
  export SEMAPHORE_GIT_REF_TYPE=pull-request
  export SEMAPHORE_GIT_BRANCH=master
  export SEMAPHORE_GIT_PR_BRANCH=enterprise

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
  - name: ChangeIn
    run:
      when: "false"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"
}))

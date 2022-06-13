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

  #{spc} list-diff > /tmp/output.txt
})

output = File.readlines('/tmp/output.txt')
  .map { |line| line.strip }
  .reject { |line| line.empty? }

assert_eq($?.exitstatus, 0)
assert_eq(["app/a.yml"], output)

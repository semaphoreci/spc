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
      when: "change_in('/app')"

  - name: Test2
    run:
      when: "change_in('/lib')"

  - name: Test3
    run:
      when: "change_in(['/lib'], {branch_range: '$SEMAPHORE_GIT_COMMIT_RANGE'})"
}

origin = TestRepoForChangeIn.setup()

origin.add_file('.semaphore/semaphore.yml', pipeline)
origin.commit!("Bootstrap")

origin.add_file("lib/a.yml", "hello")
origin.commit!("Bootstrap lib")

origin.create_branch("dev")
origin.add_file("app/a.yml", "hello")
origin.commit!("Bootstrap app")

repo = origin.clone_local_copy(branch: "dev")
repo.run(%{
  export SEMAPHORE_GIT_SHA=$(git rev-parse HEAD)
  export SEMAPHORE_GIT_COMMIT_RANGE=$(git rev-parse HEAD~2)...$(git rev-parse HEAD~1)
  echo "SEMAPHORE_GIT_COMMIT_RANGE: $SEMAPHORE_GIT_COMMIT_RANGE"

  #{spc} list-diff --branch-range '$SEMAPHORE_GIT_COMMIT_RANGE' > /tmp/output.txt
})

output = File.readlines('/tmp/output.txt')
  .map { |line| line.strip }
  .reject { |line| line.empty? }
  
assert_eq($?.success?, true)
assert_eq(["lib/a.yml"], output)

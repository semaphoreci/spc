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
      when: "branch = 'master' and change_in(['/app', 'log.txt'], {on_tags: true})"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"

  - name: Test2
    run:
      when: "branch = 'master' and change_in('/app', {on_tags: false})"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"

  - name: Test3
    run:
      when: "branch = 'master' and change_in('/app')"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"
}

origin = TestRepoForChangeIn.setup()

origin.add_file('.semaphore/semaphore.yml', pipeline)
origin.commit!("Bootstrap")

origin.create_branch("dev")
origin.add_file("lib/B.txt", "hello")
origin.commit!("Changes in dev")

repo = origin.clone_local_copy(branch: "dev")

repo.run(%{
  export SEMAPHORE_GIT_REF_TYPE=tag
  #{spc} list-diff 1>/tmp/output.txt 2>/tmp/error.txt
})

output = File.readlines('/tmp/output.txt')
  .map { |line| line.strip }
  .reject { |line| line.empty? }

error = File.readlines('/tmp/error.txt')
  .map { |line| line.strip }
  .reject { |line| line.empty? }

assert_eq($?.exitstatus, 0)
assert_eq([], output)
assert_eq(["Running on a tag, skipping evaluation."], error)

repo.run(%{
  export SEMAPHORE_GIT_REF_TYPE=branch
  #{spc} list-diff > /tmp/output.txt
})

output = File.readlines('/tmp/output.txt')
  .map { |line| line.strip }
  .reject { |line| line.empty? }

assert_eq($?.exitstatus, 0)
assert_eq(["lib/B.txt"], output)

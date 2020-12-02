# rubocop:disable all

require_relative "../e2e"

system "rm -f /tmp/output.yml"
system "rm -f /tmp/logs.jsonl"

pipeline = %{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: Test
    skip:
      when: "branch = 'master' and change_in('/lib', {default_branch: 'random'})"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"
}

origin = TestRepoForChangeIn.setup()

origin.add_file('.semaphore/semaphore.yml', pipeline)
origin.commit!("Bootstrap")

origin.add_file("lib/A.txt", "hello")
origin.commit!("Changes on master")

origin.create_branch("dev")
origin.add_file("lib/B.txt", "hello")
origin.commit!("Changes in dev")

repo = origin.clone_local_copy(branch: "dev")
repo.run("#{spc} evaluate change-in --input .semaphore/semaphore.yml --output /tmp/output.yml --logs /tmp/logs.jsonl", fail: false)

assert_eq($?.exitstatus, 1)

errors = File.read('/tmp/logs.jsonl').lines.map { |l| JSON.parse(l) }
assert_eq(errors.size, 1)


assert_eq(errors[0], {
  "type" => "ErrorChangeInMissingBranch",
  "message" => "Unknown git reference 'random'.",
  "location" => {
    "file" => ".semaphore/semaphore.yml",
    "path" => ["blocks", "0", "skip", "when"]
  }
})

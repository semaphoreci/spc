# rubocop:disable all

require_relative "../e2e"
require 'yaml'
require 'json'

#
# Prepare a repository with two branches, master and dev.
#
system %{
  rm -f /tmp/output.yml
  rm -f /tmp/logs.jsonl
  rm -rf /tmp/test-repo
  mkdir /tmp/test-repo && cd /tmp/test-repo && git init

  # master branch
  mkdir lib .semaphore
  echo A > lib/A.txt
  git add . && git commit -m 'Bootstrap'

  # dev branch
  git checkout -b dev
  echo B > lib/B.txt
  git add . && git commit -m 'Changes in dev'
}

#
# Create a .semaphore/semaphore.yml file.
#

File.write('/tmp/test-repo/.semaphore/semaphore.yml', %{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: Test
    skip:
      when: "branch = 'master' and change_in('lib', {default_branch: 'random'})"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"
})

#
# Evaluate the change-ins
#
system(%{
  cd /tmp/test-repo

  #{spc} evaluate change-in \
     --input .semaphore/semaphore.yml \
     --output /tmp/output.yml \
     --logs /tmp/logs.jsonl
})

assert_eq($?.exitstatus, 1)

#
# Verify that the results are OK.
#
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


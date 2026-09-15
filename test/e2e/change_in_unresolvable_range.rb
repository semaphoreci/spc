# rubocop:disable all

require_relative "../e2e"
require 'yaml'
require 'json'

#
# A commit range with no merge base is not a git failure: every git command
# succeeds, there is simply no common ancestor. Deepening the clone cannot
# help, so change_in resolves to false, as it always has, and says why.
#
# This is deliberately not a hard failure. Failing the compilation would take
# down every pipeline in a repository that is in this state, including blocks
# that use no change_in at all, which is a much worse outcome than one block
# being skipped.
#

pipeline = %{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: GatedOne
    run:
      when: "change_in('/lib')"
    task:
      jobs:
        - name: GatedOne
          commands:
            - echo "gated"
  - name: GatedTwo
    run:
      when: "change_in('/lib')"
    task:
      jobs:
        - name: GatedTwo
          commands:
            - echo "gated"
  - name: Ungated
    task:
      jobs:
        - name: Ungated
          commands:
            - echo "ungated"
}

origin = TestRepoForChangeIn.setup()

origin.add_file('.semaphore/semaphore.yml', pipeline)
origin.commit!("Bootstrap")

origin.add_file("lib/base.txt", "base")
origin.commit!("Base change on master")

#
# An orphan branch: it shares no history at all with the default branch.
#
origin.run("git checkout --orphan feature")
origin.run("git rm -rf . > /dev/null 2>&1 || true")

origin.add_file('.semaphore/semaphore.yml', pipeline)
origin.add_file("lib/feature.txt", "changed")
origin.commit!("Orphan root with a change in lib")

origin.switch_branch("master")

repo = origin.clone_local_copy(branch: "feature", depth: 1, single_branch: true)
repo.run("git checkout --detach")

system "rm -f /tmp/output.yml /tmp/logs.jsonl /tmp/stdout.txt"

repo.run(%{
  export SEMAPHORE_GIT_SHA=$(git rev-parse HEAD)
  export SEMAPHORE_GIT_REF_TYPE=pull-request
  export SEMAPHORE_GIT_BRANCH=master
  export SEMAPHORE_GIT_PR_BRANCH=feature

  #{spc} compile \
    --input .semaphore/semaphore.yml \
    --output /tmp/output.yml \
    --logs /tmp/logs.jsonl > /tmp/stdout.txt 2>&1
}, fail: false)

# The compilation succeeds and every block survives.
assert_eq($?.exitstatus, 0)

assert_eq(YAML.load_file('/tmp/output.yml'), YAML.load(%{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: GatedOne
    run:
      when: "false"
    task:
      jobs:
        - name: GatedOne
          commands:
            - echo "gated"
  - name: GatedTwo
    run:
      when: "false"
    task:
      jobs:
        - name: GatedTwo
          commands:
            - echo "gated"
  - name: Ungated
    task:
      jobs:
        - name: Ungated
          commands:
            - echo "ungated"
}))

# Nothing is reported as a structured error, so the workflow page stays clean.
assert_eq(File.exist?('/tmp/logs.jsonl') && File.read('/tmp/logs.jsonl').strip, "")

# The reason is visible in the initialization log rather than silently swallowed.
output = File.read('/tmp/stdout.txt')

assert(output.include?("WARNING: no merge base found for commit range 'master...feature'"))
assert(output.include?("WARNING: resolving this change_in to false."))

#
# The verdict is reached once and reused. Both gated blocks evaluate the same
# range, but only the first one pays for the round of deepen fetches.
#
deepens = output.scan(/Running git fetch origin --deepen/).size

assert_eq(deepens, 10)

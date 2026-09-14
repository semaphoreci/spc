# rubocop:disable all

require_relative "../e2e"
require 'yaml'

#
# Resolving a commit range in a shallow clone requires deepening it first.
# When that deepen fails, the failure must not be absorbed into a 'false'
# result: a transient failure is retried, and a persistent one is reported.
#

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

ATTEMPTS_FILE = "/tmp/deepen-attempts"
SHIM_DIR = "/tmp/git-shim"

#
# Installs a git wrapper that fails the first `fail_attempts` invocations of
# `git fetch --deepen`, the way git does when it no longer recognises the
# shallow file. When apply_deepen is set the deepen still takes effect before
# the failure is reported, which is what a transient failure looks like.
#
def install_git_shim(fail_attempts:, apply_deepen:)
  real_git = `which git`.strip
  deepen = apply_deepen ? %{  #{real_git} "$@" >/dev/null 2>&1\n} : ""

  system "rm -rf #{SHIM_DIR} #{ATTEMPTS_FILE}"
  system "mkdir -p #{SHIM_DIR}"

  File.write("#{SHIM_DIR}/git", %{#!/bin/bash
if [[ "$*" == *"--deepen"* ]]; then
  attempts=$(cat #{ATTEMPTS_FILE} 2>/dev/null || echo 0)
  attempts=$((attempts + 1))
  echo "$attempts" > #{ATTEMPTS_FILE}

  if [ "$attempts" -le #{fail_attempts} ]; then
#{deepen}    echo "fatal: shallow file has changed since we read it" >&2
    exit 128
  fi
fi

exec #{real_git} "$@"
})

  system "chmod +x #{SHIM_DIR}/git"
end

#
# An origin whose default branch has moved ahead of the feature branch, so the
# merge base sits outside the depth of a shallow clone.
#
origin = TestRepoForChangeIn.setup()

origin.add_file('.semaphore/semaphore.yml', pipeline)
origin.commit!("Bootstrap")

origin.add_file("lib/base.txt", "base")
origin.commit!("Base change on master")

origin.run("git branch feature")
origin.run("for i in $(seq 1 5); do git commit -q --allow-empty -m \"Master history $i\"; done")

origin.switch_branch("feature")
origin.add_file("lib/feature.txt", "changed")
origin.commit!("Change in lib")

origin.switch_branch("master")

def shallow_clone(origin)
  repo = origin.clone_local_copy(branch: "feature", depth: 1, single_branch: true)
  repo.run("git checkout --detach")
  repo
end

def semaphore_env
  %{
    export SEMAPHORE_GIT_SHA=$(git rev-parse HEAD)
    export SEMAPHORE_GIT_REF_TYPE=pull-request
    export SEMAPHORE_GIT_BRANCH=master
    export SEMAPHORE_GIT_PR_BRANCH=feature
    export PATH=#{SHIM_DIR}:$PATH
  }
end

#
# A deepen that fails once is retried, and the condition resolves correctly.
#
repo = shallow_clone(origin)
install_git_shim(fail_attempts: 1, apply_deepen: true)

system "rm -f /tmp/output.yml /tmp/logs.jsonl"

repo.run(%{
  #{semaphore_env}

  #{spc} compile \
    --input .semaphore/semaphore.yml \
    --output /tmp/output.yml \
    --logs /tmp/logs.jsonl
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
      when: "true"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"
}))

#
# A deepen that keeps failing is reported, instead of resolving to an empty diff.
#
repo = shallow_clone(origin)
install_git_shim(fail_attempts: 100, apply_deepen: false)

system "rm -f /tmp/output.txt"

repo.run(%{
  #{semaphore_env}

  #{spc} list-diff --default-branch master > /tmp/output.txt 2>&1
}, fail: false)

assert_eq($?.exitstatus, 1)

output = File.read('/tmp/output.txt')

assert(output.include?("Failed to resolve the git diff"))
assert(output.include?("master...feature"))

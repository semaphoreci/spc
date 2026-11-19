# rubocop:disable all

require_relative "../e2e"
require 'yaml'

#
# The push range of a branch is compared end to end. When a force push moves
# the branch to a commit that does not descend from its previous one, the
# previous commit is fetched by its id, and the result covers both what the
# branch gained and what it lost.
#

pipeline = %{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: Core
    run:
      when: "change_in('/core')"
  - name: Create
    run:
      when: "change_in('/create')"
  - name: Docs
    run:
      when: "change_in('/docs')"
}

def expected(core, create, docs)
  YAML.load(%{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: Core
    run:
      when: "#{core}"
  - name: Create
    run:
      when: "#{create}"
  - name: Docs
    run:
      when: "#{docs}"
})
end

#
# master carries a commit of its own (hotfix), develop has moved on by two.
#
origin = TestRepoForChangeIn.setup()

origin.add_file('.semaphore/semaphore.yml', pipeline)
origin.add_file('core/lib.txt', 'v0')
origin.add_file('create/app.txt', 'v0')
origin.add_file('docs/readme.txt', 'v0')
origin.commit!("Base")

origin.create_branch("develop")
origin.add_file('create/app.txt', 'a')
origin.commit!("A")
origin.add_file('create/app.txt', 'b')
origin.commit!("B")

origin.switch_branch("master")
origin.add_file('core/lib.txt', 'hotfix')
origin.commit!("Hotfix")

def rev(ref)
  `cd /tmp/test-repo-origin && git rev-parse #{ref}`.strip
end

hotfix = rev("master")
a = rev("develop~1")
b = rev("develop")

#
# Points master at a commit the way a push does, clones it the way the
# compilation job does, and compiles with the push's commit range.
#
def compile_after_push(origin, new_master, commit_range)
  origin.run("git update-ref refs/heads/master #{new_master}")

  repo = origin.clone_local_copy(branch: "master", depth: 50, single_branch: true)

  system "rm -f /tmp/output.yml /tmp/logs.jsonl /tmp/stdout.txt"

  repo.run(%{
    export SEMAPHORE_GIT_SHA=$(git rev-parse HEAD)
    export SEMAPHORE_GIT_COMMIT_RANGE=#{commit_range}

    #{spc} compile \
      --input .semaphore/semaphore.yml \
      --output /tmp/output.yml \
      --logs /tmp/logs.jsonl > /tmp/stdout.txt 2>&1
  }, fail: false)

  status = $?.exitstatus
  puts File.read('/tmp/stdout.txt')
  status
end

def no_structured_errors
  !File.exist?('/tmp/logs.jsonl') || File.read('/tmp/logs.jsonl').strip.empty?
end

#
# master moved onto develop: the hotfix is undone and develop's change is in.
#
status = compile_after_push(origin, b, "#{hotfix}...#{b}")

assert_eq(status, 0)
assert_eq(YAML.load_file('/tmp/output.yml'), expected(true, true, false))
assert(no_structured_errors)
assert(File.read('/tmp/stdout.txt').include?("Running git fetch --no-tags --depth=1 origin #{hotfix}"))

#
# master moved back one commit: B's change is reverted.
#
status = compile_after_push(origin, a, "#{b}...#{a}")

assert_eq(status, 0)
assert_eq(YAML.load_file('/tmp/output.yml'), expected(false, true, false))
assert(no_structured_errors)

#
# A normal fast-forward needs nothing fetched and gives the same answer as
# before.
#
status = compile_after_push(origin, b, "#{a}...#{b}")

assert_eq(status, 0)
assert_eq(YAML.load_file('/tmp/output.yml'), expected(false, true, false))
assert(!File.read('/tmp/stdout.txt').include?("--depth=1"))

#
# A previous commit that cannot be obtained leaves nothing to compare with:
# every file of the new commit counts as changed, and the log says why.
#
gone = "1111111111111111111111111111111111111111"

status = compile_after_push(origin, b, "#{gone}...#{b}")

assert_eq(status, 0)
assert_eq(YAML.load_file('/tmp/output.yml'), expected(true, true, true))
assert(no_structured_errors)
assert(File.read('/tmp/stdout.txt').include?("WARNING: treating every file in the new commit as changed."))

# rubocop:disable all

require_relative "../e2e"

#
# list-diff resolves the push range the same way change_in does: end to end,
# fetching the previous commit by its id when a force push left it behind.
#

origin = TestRepoForChangeIn.setup()

origin.add_file('.semaphore/semaphore.yml', "version: v1.0\n")
origin.add_file('core/lib.txt', 'v0')
origin.add_file('create/app.txt', 'v0')
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
gone = "1111111111111111111111111111111111111111"

def list_after_push(origin, new_master, commit_range)
  origin.run("git update-ref refs/heads/master #{new_master}")

  repo = origin.clone_local_copy(branch: "master", depth: 50, single_branch: true)

  repo.run(%{
    export SEMAPHORE_GIT_SHA=$(git rev-parse HEAD)
    export SEMAPHORE_GIT_COMMIT_RANGE=#{commit_range}

    #{spc} list-diff > /tmp/output.txt 2> /tmp/stderr.txt
  })

  File.readlines('/tmp/output.txt').map(&:strip).reject(&:empty?)
end

assert_eq(list_after_push(origin, b, "#{hotfix}...#{b}"), ["core/lib.txt", "create/app.txt"])
assert_eq(list_after_push(origin, a, "#{b}...#{a}"), ["create/app.txt"])
assert_eq(list_after_push(origin, b, "#{a}...#{b}"), ["create/app.txt"])

assert_eq(
  list_after_push(origin, b, "#{gone}...#{b}"),
  [".semaphore/semaphore.yml", "core/lib.txt", "create/app.txt"]
)
assert(File.read('/tmp/stderr.txt').include?("listing every file"))

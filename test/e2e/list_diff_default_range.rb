# rubocop:disable all

require_relative "../e2e"
require 'yaml'

#
# If the change_in is evaluated on the default branch, usually master branch,
# the commit range is the one provided by the git post commit hook.
#
# To configure this range, a developer can pass a default_range parameter to
# the function.
#
# The default value of this parameter is $SEMAPHORE_GIT_COMMIT_RANGE.
#

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

origin.add_file("app/a.yml", "hello")
origin.commit!("Changes in app")

origin.add_file("lib/b.yml", "hello")
origin.commit!("Changes in lib")

origin.add_file("test/c.yml", "hello")
origin.commit!("Changes in test")

repo = origin.clone_local_copy(branch: "master")
repo.list_branches

fixtures = {
  'HEAD~3..HEAD~1' => ['app/a.yml', 'lib/b.yml'],
  'HEAD~2..HEAD' => ['lib/b.yml', 'test/c.yml'],
  'HEAD~1..HEAD' => ['test/c.yml']
}

fixtures.each do |default_range, expected|
  repo.run(%{  
    export SEMAPHORE_GIT_COMMIT_RANGE="$(git rev-parse HEAD~2)...$(git rev-parse HEAD)"
    echo "Passing $SEMAPHORE_GIT_COMMIT_RANGE to the compiler"
    echo ""
  
    #{spc} list-diff --default-range #{default_range} > /tmp/output.txt
  })

  output = File.readlines('/tmp/output.txt')
    .map { |line| line.strip }
    .reject { |line| line.empty? }
  
  assert_eq($?.success?, true)
  assert_eq(expected, output)
end




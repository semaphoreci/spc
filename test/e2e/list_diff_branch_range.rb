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

origin.create_branch("feature-1")
origin.add_file("lib/a.yml", "hello")
origin.commit!("Bootstrap lib")

repo = origin.clone_local_copy(branch: "feature-1")

fixtures = {
  '$SEMAPHORE_MERGE_BASE...$SEMAPHORE_GIT_SHA' => ['app/a.yml', 'lib/a.yml'],
  '$SEMAPHORE_GIT_SHA^...$SEMAPHORE_GIT_SHA' => ['lib/a.yml'],
  'dev...$SEMAPHORE_GIT_SHA' => ['lib/a.yml']
}

fixtures.each do |branch_range, expected|
  output = repo.run(%{
    export SEMAPHORE_GIT_SHA=$(git rev-parse HEAD)
    export SEMAPHORE_GIT_BRANCH=feature-1
  
    #{spc} list-diff --branch-range '#{branch_range}' > /tmp/output.txt
  })

  output = File.readlines('/tmp/output.txt')
    .map { |line| line.strip }
    .reject { |line| line.empty? }
  
  assert_eq($?.success?, true)
  assert_eq(expected, output)
end

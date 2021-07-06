# rubocop:disable all

#
# This file verifies that we have reasonable high performance for evaluating
# multiple change_in expressions on the same commit range.
#
# The implementation is optimized to run git fetch only when necessary, and to
# avoid unnecessary git pulls.
#

require_relative "../e2e"
require 'yaml'

#
# To test the perf, we are creating a pipeline with 100 blocks.
# Every block runs only if one of the folders in the repository changes.
#
# For example, "Block1" runs only if "dir1" is changed.
#

pipeline = %{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
}

100.times.each do |index|
  pipeline += %{  - { name: Block#{index}, skip: { when: "branch = 'master' and change_in('/dir#{index}')" }} \n}
end

origin = TestRepoForChangeIn.setup()

origin.add_file('.semaphore/semaphore.yml', pipeline)
origin.commit!("Bootstrap")

origin.create_branch("dev")
100.times.each { |index| origin.add_file("dir#{index}/a.txt", "hello") }
origin.commit!("Changes in dev")

repo = origin.clone_local_copy(branch: "dev")
repo.list_branches

start = Time.now.to_i
repo.run("#{spc} compile --input .semaphore/semaphore.yml --output /tmp/output.yml --logs /tmp/logs.yml")
finish = Time.now.to_i

duration = finish - start

# Assert that this calculation can be done under 5 seconds
if duration > 5
  abort("Processing the pipeline took more than our goal of 5 seconds. Current #{duration} seconds.")
end

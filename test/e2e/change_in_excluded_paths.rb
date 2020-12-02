# rubocop:disable all

require_relative "../e2e"
require 'yaml'

#
# The exclude option lets us to define paths that are not included in the
# change_in set. An situation where this is useful is repository that has:
#
# - A main project in the root of the repository, for example a Rails project
# - A sub-project called /client
#
# In this scenario, we want to have two blocks in the pipeline, one for the
# main application, and one for the client.
#
# These blocks need to satisfy the following:
#
# - If there are changes in the root and in the /client directory => Run both blocks.
# - If there are changes only outside of the /client directory => Run only the backend tests.
# - If there are changes in the /client directory => Run the frontend tests.
#
# For the Client block, it is simple to construct such a rule.
#
#   change_in('/client')
#
# But for the backend, it is trickier. We want the block to run if there are any
# changes in the repository, except if those changes are comming from the client
# directory. To codify this exlusion, the exclude list can be used:
#
#   change_in('/', {exclude: ['/client']})
#

#
# Prepare a repository with the following branches:
#
#  - master
#  - client-changes             (has changes only in the client dir)
#  - backend-change             (has changes only in the backend)
#  - changes-in-both-places     (has changes in two places)
#
# The dev branch will have change only
#
#
pipeline = %{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: Client
    run:
      when: "change_in('/client')"

  - name: Backend
    run:
      when: "change_in('/', {exclude: ['/client']})"
}

origin = TestRepoForChangeIn.setup()

origin.add_file('.semaphore/semaphore.yml', pipeline)
origin.commit!("Bootstrap")

origin.add_file("client/app.js", "hello")
origin.add_file("config.txt", "hello")
origin.commit!("Changes on master")

repo = origin.clone_local_copy(branch: "master")

#
# Testing out the scenario where only the client changed.
#

repo.create_branch("client-changes")
repo.add_file("client/app.js", "hello hello")
repo.commit!("Change things in the client")

repo.run("#{spc} evaluate change-in --input .semaphore/semaphore.yml --output /tmp/output.yml --logs /tmp/logs.yml")

assert_eq(YAML.load_file('/tmp/output.yml'), YAML.load(%{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: Client
    run:
      when: "true"

  - name: Backend
    run:
      when: "false"
}))

#
# Testing out the scenario where only the backend changed.
#
repo.switch_branch("master")
repo.create_branch("backend-changes")

repo.add_file("config.txt", "hello hello")
repo.commit!("Change things in the backend")

repo.run("#{spc} evaluate change-in --input .semaphore/semaphore.yml --output /tmp/output.yml --logs /tmp/logs.yml")

assert_eq(YAML.load_file('/tmp/output.yml'), YAML.load(%{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: Client
    run:
      when: "false"

  - name: Backend
    run:
      when: "true"
}))

#
# Testing out the scenario where both and client have changes.
#
repo.switch_branch("master")
repo.create_branch("changes-in-both-places")

repo.add_file("client/app.txt", "hello hello")
repo.add_file("config.txt", "hello hello")
repo.commit!("Change things in both places")

repo.run("#{spc} evaluate change-in --input .semaphore/semaphore.yml --output /tmp/output.yml --logs /tmp/logs.yml")

assert_eq(YAML.load_file('/tmp/output.yml'), YAML.load(%{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: Client
    run:
      when: "true"

  - name: Backend
    run:
      when: "true"
}))

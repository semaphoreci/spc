# rubocop:disable all

require_relative "../e2e"
require 'yaml'

#
# The include option lets us to define paths that are explicitly included in the
# change_in set. This is useful in a monorepo with multiple services where you
# want to run a pipeline block only when specific directories have changes.
#
# For example, in a repository with multiple services:
# - /frontend   - contains frontend code
# - /backend    - contains backend code
# - /shared     - contains shared code used by both
#
# In this scenario, we want to have specialized blocks that run only when specific
# parts of the codebase change:
#
# - Frontend block should run when:
#   - Changes in /frontend OR
#   - Changes in /shared (since backend uses shared code)
# - Backend block should run when:
#   - Changes in /backend OR
#   - Changes in /shared (since backend uses shared code)
#
# For these blocks, we can use the include option:
#
#   change_in('/', {include: ['/frontend', '/shared']})  # Frontend block
#   change_in('/', {include: ['/backend', '/shared']})   # Backend block
#
# This ensures that the pipeline block only runs when there are changes in the specified
# directories, and not when changes happen in other directories.
#

#
# Prepare a repository with the following branches:
#
# - master
# - frontend-changes        (has changes only in the frontend dir)
# - backend-changes         (has changes only in the backend dir)
# - shared-changes          (has changes only in the shared dir)
# - unrelated-changes       (has changes in other directories not specified in includes)
#

pipeline = %{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: Frontend
    run:
      when: "change_in('/', {include: ['/frontend', '/shared']})"

  - name: Backend
    run:
      when: "change_in('/', {include: ['/backend', '/shared']})"
}

origin = TestRepoForChangeIn.setup()

origin.add_file('.semaphore/semaphore.yml', pipeline)
origin.commit!("Bootstrap")

origin.add_file("frontend/app.js", "hello")
origin.add_file("backend/server.js", "hello")
origin.add_file("shared/utils.js", "hello")
origin.add_file("docs/readme.md", "hello")
origin.commit!("Initial setup on master")

repo = origin.clone_local_copy(branch: "master")

#
# Testing out the scenario where only the frontend changed.
#

repo.create_branch("frontend-changes")
repo.add_file("frontend/app.js", "hello hello")
repo.commit!("Change things in the frontend")

repo.run("#{spc} compile --input .semaphore/semaphore.yml --output /tmp/output.yml --logs /tmp/logs.yml")

assert_eq(YAML.load_file('/tmp/output.yml'), YAML.load(%{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: Frontend
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

repo.add_file("backend/server.js", "hello hello")
repo.commit!("Change things in the backend")

repo.run("#{spc} compile --input .semaphore/semaphore.yml --output /tmp/output.yml --logs /tmp/logs.yml")

assert_eq(YAML.load_file('/tmp/output.yml'), YAML.load(%{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: Frontend
    run:
      when: "false"

  - name: Backend
    run:
      when: "true"
}))

#
# Testing out the scenario where the shared directory changed.
#
repo.switch_branch("master")
repo.create_branch("shared-changes")

repo.add_file("shared/utils.js", "hello hello")
repo.commit!("Change things in shared")

repo.run("#{spc} compile --input .semaphore/semaphore.yml --output /tmp/output.yml --logs /tmp/logs.yml")

assert_eq(YAML.load_file('/tmp/output.yml'), YAML.load(%{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: Frontend
    run:
      when: "true"

  - name: Backend
    run:
      when: "true"
}))

#
# Testing out the scenario where changes are in unrelated directories.
#
repo.switch_branch("master")
repo.create_branch("unrelated-changes")

repo.add_file("docs/readme.md", "hello hello")
repo.commit!("Change things in unrelated directory")

repo.run("#{spc} compile --input .semaphore/semaphore.yml --output /tmp/output.yml --logs /tmp/logs.yml")

assert_eq(YAML.load_file('/tmp/output.yml'), YAML.load(%{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: Frontend
    run:
      when: "false"

  - name: Backend
    run:
      when: "false"
}))

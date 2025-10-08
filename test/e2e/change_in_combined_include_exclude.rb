# rubocop:disable all

require_relative "../e2e"
require 'yaml'

#
# This test demonstrates how include and exclude options work together.
# The logic order is:
# 1. First, check if a path is excluded (if so, skip it regardless of include patterns)
# 2. Then, check if a path matches include patterns (if include is specified)
# 3. Finally, check if it matches the main path pattern
#
# Consider a repository structure:
# - /frontend/web      - contains web frontend
# - /frontend/mobile   - contains mobile frontend 
# - /backend           - contains backend code
# - /shared            - contains shared code used by all components
#
# In this scenario:
# - Web Frontend team wants a block that runs when only web frontend or shared code changes, but NOT when mobile frontend changes
# - Mobile Frontend team wants a block that runs when only mobile frontend or shared code changes, but NOT when web frontend changes
#

pipeline = %{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: WebFrontend
    run:
      when: "change_in('/', {include: ['/frontend', '/shared'], exclude: ['/frontend/mobile']})"

  - name: MobileFrontend
    run:
      when: "change_in('/', {include: ['/frontend', '/shared'], exclude: ['/frontend/web']})"
}

origin = TestRepoForChangeIn.setup()

origin.add_file('.semaphore/semaphore.yml', pipeline)
origin.commit!("Bootstrap")

origin.add_file("frontend/web/app.js", "hello")
origin.add_file("frontend/mobile/app.js", "hello")
origin.add_file("shared/utils.js", "hello")
origin.commit!("Initial setup on master")

repo = origin.clone_local_copy(branch: "master")

#
# Testing scenario where only web frontend changed.
#

repo.create_branch("web-frontend-changes")
repo.add_file("frontend/web/app.js", "hello hello")
repo.commit!("Change things in web frontend")

repo.run("#{spc} compile --input .semaphore/semaphore.yml --output /tmp/output.yml --logs /tmp/logs.yml")

assert_eq(YAML.load_file('/tmp/output.yml'), YAML.load(%{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: WebFrontend
    run:
      when: "true"

  - name: MobileFrontend
    run:
      when: "false"
}))

#
# Testing scenario where only mobile frontend changed.
#
repo.switch_branch("master")
repo.create_branch("mobile-frontend-changes")

repo.add_file("frontend/mobile/app.js", "hello hello")
repo.commit!("Change things in mobile frontend")

repo.run("#{spc} compile --input .semaphore/semaphore.yml --output /tmp/output.yml --logs /tmp/logs.yml")

assert_eq(YAML.load_file('/tmp/output.yml'), YAML.load(%{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: WebFrontend
    run:
      when: "false"

  - name: MobileFrontend
    run:
      when: "true"
}))

#
# Testing scenario where shared code changed (should run both)
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
  - name: WebFrontend
    run:
      when: "true"

  - name: MobileFrontend
    run:
      when: "true"
}))

#
# Testing scenario where both web and mobile changed
# Only the parts not excluded should run their respective pipelines
#
repo.switch_branch("master")
repo.create_branch("both-frontend-changes")

repo.add_file("frontend/web/app.js", "hello hello")
repo.add_file("frontend/mobile/app.js", "hello hello")
repo.commit!("Change things in both frontends")

repo.run("#{spc} compile --input .semaphore/semaphore.yml --output /tmp/output.yml --logs /tmp/logs.yml")

assert_eq(YAML.load_file('/tmp/output.yml'), YAML.load(%{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: WebFrontend
    run:
      when: "true"

  - name: MobileFrontend
    run:
      when: "true"
}))

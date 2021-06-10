# rubocop:disable all

#
# This test verifies if the compiler is able to recognize and
# process all locations where a change_in expression can appear.
#

require_relative "../e2e"
require 'yaml'

pipeline = %{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

fail_fast:
  cancel:
    when: "branch = 'master' and change_in('/lib')"
  stop:
    when: "branch = 'master' and change_in('/app')"

auto_cancel:
  queued:
    when: "branch = 'master' and change_in('/lib')"
  running:
    when: "branch = 'master' and change_in('/app')"

queue:
  - name: test
    when: "branch = 'master' and change_in('/lib')"

  - name: test2
    when: "branch = 'master' and change_in('/app')"

global_job_config:
  priority:
    - value: 1
      when: "branch = 'master' and change_in('/lib')"

    - value: 10
      when: "branch = 'master' and change_in('/lib')"

blocks:
  - name: Test
    skip:
      when: "branch = 'master' and change_in('/lib')"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"
          priority:
            - value: 1
              when: "branch = 'master' and change_in('/lib')"

            - value: 10
              when: "branch = 'master' and change_in('/lib')"

  - name: Test2
    run :
      when: "branch = 'master' and change_in('/app')"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"

promotions:
  - name: Staging
    auto_promote:
      when: "branch = 'master' and change_in('/lib')"

  - name: Staging2
    auto_promote:
      when: "branch = 'master' and change_in('/lib')"

  - name: Staging3
    auto_promote:
      when: "branch = 'master' and change_in('/lib')"
}

origin = TestRepoForChangeIn.setup()

origin.add_file('.semaphore/semaphore.yml', pipeline)
origin.commit!("Bootstrap")

origin.add_file("lib/A.txt", "hello")
origin.commit!("Changes on master")

origin.create_branch("dev")
origin.add_file("lib/B.txt", "hello")
origin.commit!("Changes in dev")

repo = origin.clone_local_copy(branch: "dev")
repo.list_branches
repo.run("#{spc} compile --input .semaphore/semaphore.yml --output /tmp/output.yml --logs /tmp/logs.yml")

assert_eq(YAML.load_file('/tmp/output.yml'), YAML.load(%{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

fail_fast:
  cancel:
    when: "(branch = 'master') and true"
  stop:
    when: "(branch = 'master') and false"

auto_cancel:
  queued:
    when: "(branch = 'master') and true"
  running:
    when: "(branch = 'master') and false"

queue:
  - name: test
    when: "(branch = 'master') and true"

  - name: test2
    when: "(branch = 'master') and false"

global_job_config:
  priority:
    - value: 1
      when: "(branch = 'master') and true"

    - value: 10
      when: "(branch = 'master') and true"

blocks:
  - name: Test
    skip:
      when: "(branch = 'master') and true"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"
          priority:
            - value: 1
              when: "(branch = 'master') and true"
            - value: 10
              when: "(branch = 'master') and true"

  - name: Test2
    run :
      when: "(branch = 'master') and false"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"

promotions:
  - name: Staging
    auto_promote:
      when: "(branch = 'master') and true"

  - name: Staging2
    auto_promote:
      when: "(branch = 'master') and true"

  - name: Staging3
    auto_promote:
      when: "(branch = 'master') and true"
}))

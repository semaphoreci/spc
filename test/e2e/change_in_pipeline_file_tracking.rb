# rubocop:disable all

require_relative "../e2e"
require 'yaml'

#
# By default, when a Pipeline YAML file is changed, every block is executed.
# The reasoning is that if you have changed the YAML file, conditions have
# changed and it is better to execute every block.
#
# The value of the pipeline_file is by default 'track' for blocks, queues,
# auto_cancel, and fast_fail.
#
# However, for promotions, the default value is 'ignore'.
#

pipeline = %{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: Implicit track
    run:
      when: "branch = 'master' and change_in('/lib')"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"

  - name: Explicit ignore
    run:
      when: "branch = 'master' and change_in('/lib', {pipeline_file: 'ignore'})"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"

  - name: Explicit track
    run:
      when: "branch = 'master' and change_in('/lib', {pipeline_file: 'track'})"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"

  - name: Explicit track + list of paths
    run:
      when: "branch = 'master' and change_in(['/lib', 'log.txt'], {pipeline_file: 'track'})"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"

promotions:
  - name: P1
    auto_promote:
      when: "change_in('/lib')"

  - name: P2
    auto_promote:
      when: "change_in('/lib', {pipeline_file: 'ignore'})"

  - name: P3
    auto_promote:
      when: "change_in('/lib', {pipeline_file: 'track'})"
}

origin = TestRepoForChangeIn.setup()

origin.add_file('.semaphore/semaphore.yml', pipeline)
origin.commit!("Bootstrap")

origin.create_branch("dev")
origin.run(%{echo "\n" >> .semaphore/semaphore.yml})
origin.commit!("Changes in dev")

repo = origin.clone_local_copy(branch: "dev")
repo.run("#{spc} compile --input .semaphore/semaphore.yml --output /tmp/output.yml --logs /tmp/logs.yml")

assert_eq(YAML.load_file('/tmp/output.yml'), YAML.load(%{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: Implicit track
    run:
      when: "(branch = 'master') and true"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"

  - name: Explicit ignore
    run:
      when: "(branch = 'master') and false"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"

  - name: Explicit track
    run:
      when: "(branch = 'master') and true"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"

  - name: Explicit track + list of paths
    run:
      when: "(branch = 'master') and true"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"

promotions:
  - name: P1
    auto_promote:
      when: "false"

  - name: P2
    auto_promote:
      when: "false"

  - name: P3
    auto_promote:
      when: "true"
}))

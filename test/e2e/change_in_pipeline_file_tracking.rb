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

#
# Prepare a repository with two branches, master and dev.
#
system %{
  rm -f /tmp/output.yml
  rm -rf /tmp/test-repo
  mkdir /tmp/test-repo
  cd /tmp/test-repo
  git init

  # master branch
  mkdir lib .semaphore
  echo A > lib/A.txt
}

File.write('/tmp/test-repo/.semaphore/semaphore.yml', %{
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
})

system %{
  cd /tmp/test-repo

  git add . && git commit -m 'Bootstrap'

  git checkout -b dev

  echo "\n" >> .semaphore/semaphore.yml

  git add . && git commit -m 'Changes in dev'
}

#
# Evaluate the change-ins
#
system(%{
  cd /tmp/test-repo

  #{spc} evaluate change-in \
     --input .semaphore/semaphore.yml \
     --output /tmp/output.yml \
     --logs /tmp/logs.yml
})

#
# Verify that the results are OK.
#
output = YAML.load_file('/tmp/output.yml')

assert_eq(output, YAML.load(%{
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

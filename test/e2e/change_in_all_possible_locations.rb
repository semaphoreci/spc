# rubocop:disable all

#
# This test verifies if the compiler is able to recognize and
# process all locations where a change_in expression can appear.
#

require_relative "../e2e"
require 'yaml'

#
# Prepare a repository with two branches, master and dev.
#
system %{
  rm -f /tmp/output.yml
  rm -rf /tmp/test-repo
  mkdir /tmp/test-repo && cd /tmp/test-repo && git init

  # master branch
  mkdir lib .semaphore
  echo A > lib/A.txt
  git add . && git commit -m 'Bootstrap'

  # dev branch
  git checkout -b dev
  echo B > lib/B.txt
  git add . && git commit -m 'Changes in dev'
}

#
# Create a .semaphore/semaphore.yml file.
#

File.write('/tmp/test-repo/.semaphore/semaphore.yml', %{
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
})

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
}))

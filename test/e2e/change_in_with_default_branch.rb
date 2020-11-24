# rubocop:disable all

require_relative "../e2e"
require 'yaml'

#
# Prepare a repository with three branches, master, main, and dev.
#
system %{
  rm -f /tmp/output.yml
  rm -rf /tmp/test-repo
  mkdir /tmp/test-repo && cd /tmp/test-repo && git init

  # master branch
  mkdir app lib .semaphore
  echo A > app/A.txt
  git add . && git commit -m 'Bootstrap'

  # main branch
  git checkout -b main
  echo B > lib/B.txt
  git add . && git commit -m 'Changes in main branch'

  # dev branch
  git checkout -b dev
  echo C > app/C.txt
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

blocks:
  - name: Test
    skip:
      when: "branch = 'master' and change_in('/lib', {default_branch: 'master'})"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"

  - name: Test2
    skip:
      when: "branch = 'master' and change_in('/lib', {default_branch: 'main'})"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"

  - name: Test3
    skip:
      when: "branch = 'master' and change_in('/lib', {default_branch: 'dev'})"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"
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
    skip:
      when: "(branch = 'master') and false"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"

  - name: Test3
    skip:
      when: "(branch = 'master') and false"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"
}))

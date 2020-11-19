# rubocop:disable all

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

blocks:
  - name: Test
    skip:
      when: "branch = 'master' and change_in('lib')"
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
}))

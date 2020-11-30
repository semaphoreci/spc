# rubocop:disable all

require_relative "../e2e"
require 'yaml'

#
# Prepare a repository with two branches, master and dev.
#
system %{
  rm -f /tmp/output.yml
  rm -rf /tmp/test-repo
  mkdir -p /tmp/test-repo/.semaphore
  cd /tmp/test-repo
  git init
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
    run:
      when: "branch = 'master' and change_in('/app')"

  - name: Test2
    run:
      when: "branch = 'master' and change_in('/lib', {branch_range: '$SEMAPHORE_MERGE_BASE...$SEMAPHORE_GIT_SHA'})"

  - name: Test3
    run:
      when: "branch = 'master' and change_in('/app', {branch_range: 'dev...$SEMAPHORE_GIT_SHA'})"

  - name: Test4
    run:
      when: "branch = 'master' and change_in('/lib', {branch_range: 'dev...$SEMAPHORE_GIT_SHA'})"
})

system %{
  cd /tmp/test-repo

  mkdir lib app test

  git add . && git commit -m "Bootstrap YAML"

  git checkout -b dev

  echo "hello" > app/a.yml
  git add . && git commit -m "Bootstrap app"

  git checkout -b feature-1

  echo "hello" > lib/b.yml
  git add . && git commit -m "Bootstrap lib"
}

#
# Evaluate the change-ins
#
system(%{
  cd /tmp/test-repo

  export SEMAPHORE_GIT_SHA=$(git rev-parse HEAD)
  export SEMAPHORE_MERGE_BASE=master

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
    run:
      when: "(branch = 'master') and true"

  - name: Test2
    run:
      when: "(branch = 'master') and true"

  - name: Test3
    run:
      when: "(branch = 'master') and false"

  - name: Test4
    run:
      when: "(branch = 'master') and true"
}))

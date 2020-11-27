# rubocop:disable all

require_relative "../e2e"
require 'yaml'

#
# If the change_in is evaluated on the default branch, usually master branch,
# the commit range is the one provided by the git post commit hook.
#
# To configure this range, a developer can pass a default_range parameter to
# the function.
#
# The default value of this parameter is $SEMAPHORE_GIT_COMMIT_RANGE.
#

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
      when: "branch = 'master' and change_in('/lib')"

  - name: Test2
    run:
      when: "branch = 'master' and change_in('/app', {default_range: 'HEAD~3..HEAD'})"

  - name: Test3
    run:
      when: "branch = 'master' and change_in('/lib', {default_range: 'HEAD~2..HEAD'})"

  - name: Test4
    run:
      when: "branch = 'master' and change_in('/app', {default_range: 'HEAD~1..HEAD'})"
})

system %{
  cd /tmp/test-repo

  mkdir lib app test

  git add . && git commit -m "Bootstrap YAML"

  echo "hello" > app/a.yml
  git add . && git commit -m "Bootstrap app"

  echo "hello" > lib/b.yml
  git add . && git commit -m "Bootstrap lib"
}

#
# Evaluate the change-ins
#
system(%{
  cd /tmp/test-repo

  echo "Displaying git log til now"
  git log

  export SEMAPHORE_GIT_COMMIT_RANGE="$(git rev-parse HEAD~2)...$(git rev-parse HEAD)"
  echo "Passing $SEMAPHORE_GIT_COMMIT_RANGE to the compiler"
  echo ""

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
      when: "(branch = 'master') and false"

  - name: Test3
    run:
      when: "(branch = 'master') and true"

  - name: Test4
    run:
      when: "(branch = 'master') and false"
}))

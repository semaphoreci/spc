# rubocop:disable all

require_relative "../e2e"
require 'yaml'

#
# The exclude option lets us to define paths that are not included in the
# change_in set. An situation where this is useful is repository that has:
#
# - A main project in the root of the repository, for example a Rails project
# - A sub-project called /client
#
# In this scenario, we want to have two blocks in the pipeline, one for the
# main application, and one for the client.
#
# These blocks need to satisfy the following:
#
# - If there are changes in the root and in the /client directory => Run both blocks.
# - If there are changes only outside of the /client directory => Run only the backend tests.
# - If there are changes in the /client directory => Run the frontend tests.
#
# For the Client block, it is simple to construct such a rule.
#
#   change_in('/client')
#
# But for the backend, it is trickier. We want the block to run if there are any
# changes in the repository, except if those changes are comming from the client
# directory. To codify this exlusion, the exclude list can be used:
#
#   change_in('/', {exclude: ['/client']})
#

#
# Prepare a repository with the following branches:
#
#  - master
#  - client-changes             (has changes only in the client dir)
#  - backend-change             (has changes only in the backend)
#  - changes-in-both-places     (has changes in two places)
#
# The dev branch will have change only
#
system %{
  rm -f /tmp/output.yml
  rm -rf /tmp/test-repo
  mkdir -p /tmp/test-repo
  cd /tmp/test-repo
  git init

  # master branch
  mkdir client .semaphore

  echo "hello" > client/app.js
  echo "hello" > config.txt
}

pipeline = %{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: Client
    run:
      when: "change_in('/client')"

  - name: Backend
    run:
      when: "change_in('/', {exclude: ['/client']})"
}

File.write('/tmp/test-repo/.semaphore/semaphore.yml', pipeline)

system %{
  cd /tmp/test-repo
  git add .
  git commit -m "Add semaphore pipeline"
}

#
# Testing out the scenario where only the client changed.
#
system %{
  cd /tmp/test-repo
  git checkout master
  git checkout -b client-changes

  echo "hello hello" > client/app.js

  git add . && git commit -m "Change things in the client"
}


system(%{
  rm -f /tmp/output.yml
  cd /tmp/test-repo

  #{spc} evaluate change-in --input .semaphore/semaphore.yml --output /tmp/output.yml --logs /tmp/logs.yml
})

output = YAML.load_file('/tmp/output.yml')

assert_eq(output, YAML.load(%{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: Client
    run:
      when: "true"

  - name: Backend
    run:
      when: "false"
}))

#
# Testing out the scenario where only the backend changed.
#
system %{
  cd /tmp/test-repo
  git checkout master
  git checkout -b backend-changes

  echo "hello hello" > config.txt

  git add . && git commit -m "Change things in the backend"
}

system(%{
  rm -f /tmp/output.yml
  cd /tmp/test-repo

  #{spc} evaluate change-in --input .semaphore/semaphore.yml --output /tmp/output.yml --logs /tmp/logs.yml
})

output = YAML.load_file('/tmp/output.yml')

assert_eq(output, YAML.load(%{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: Client
    run:
      when: "false"

  - name: Backend
    run:
      when: "true"
}))

#
# Testing out the scenario where both and client have changes.
#
system %{
  cd /tmp/test-repo
  git checkout master
  git checkout -b changes-in-both-places

  echo "hello hello" > client/app.js
  echo "hello hello" > config.txt

  git add . && git commit -m "Change things in the backend"
}

system(%{
  rm -f /tmp/output.yml
  cd /tmp/test-repo

  #{spc} evaluate change-in --input .semaphore/semaphore.yml --output /tmp/output.yml --logs /tmp/logs.yml
})

output = YAML.load_file('/tmp/output.yml')

assert_eq(output, YAML.load(%{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: Client
    run:
      when: "true"

  - name: Backend
    run:
      when: "true"
}))

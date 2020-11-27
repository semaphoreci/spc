# rubocop:disable all

require_relative "../e2e"
require 'yaml'

#
# Change In on tags has no defined value.
#
# When the CI job is building a tag, there is no clear reference what is the
# base commit to which to compare the tag.
#
# However, the YAML file is probably the same for tags and regular branches.
# For this reason, the 'on_tags' parameter allows the developer to define
# what we are going to use a replacement when a tag is run.
#
# For example:
#
#   change_in("/lib", {on_tags: true})
#
# Will always be true on tags. The changes won't be calculated.
#
# The 'true' value is also the default value of this parameter. You can also
# set this value to 'false' with:
#
#   change_in("/lib", {on_tags: false})
#
# In which case the value will always be 'false' on tags.
#

#
# Prepare a repository with two branches, master and dev.
#
system %{
  rm -f /tmp/output.yml
  rm -rf /tmp/test-repo
  mkdir -p /tmp/test-repo
  cd /tmp/test-repo
  git init

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
    run:
      when: "branch = 'master' and change_in('/app', {on_tags: true})"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"

  - name: Test2
    run:
      when: "branch = 'master' and change_in('/app', {on_tags: false})"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"

  - name: Test3
    run:
      when: "branch = 'master' and change_in('/app')"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"
})

#
# Case 1: When the current git reference is a tag:
#

system(%{
  cd /tmp/test-repo

  export SEMAPHORE_GIT_REF_TYPE=tag

  #{spc} evaluate change-in \
     --input .semaphore/semaphore.yml \
     --output /tmp/output.yml \
     --logs /tmp/logs.yml
})

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
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"

  - name: Test2
    run:
      when: "(branch = 'master') and false"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"

  - name: Test3
    run:
      when: "(branch = 'master') and true"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"
}))

#
# Case 2: When the current git reference is not a tag:
#

system(%{
  cd /tmp/test-repo

  export SEMAPHORE_GIT_REF_TYPE=branch

  #{spc} evaluate change-in \
     --input .semaphore/semaphore.yml \
     --output /tmp/output.yml \
     --logs /tmp/logs.yml
})

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
      when: "(branch = 'master') and false"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"

  - name: Test2
    run:
      when: "(branch = 'master') and false"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"

  - name: Test3
    run:
      when: "(branch = 'master') and false"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"
}))

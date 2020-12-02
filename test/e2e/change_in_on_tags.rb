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

pipeline = %{
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
}

origin = TestRepoForChangeIn.setup()

origin.add_file('.semaphore/semaphore.yml', pipeline)
origin.commit!("Bootstrap")

origin.switch_branch("dev")
origin.add_file("lib/B.txt", "hello")
origin.commit!("Changes in dev")

repo = origin.clone_local_copy(branch: "dev")

repo.run(%{
  export SEMAPHORE_GIT_REF_TYPE=tag

  #{spc} evaluate change-in --input .semaphore/semaphore.yml --output /tmp/output.yml --logs /tmp/logs.yml
})

assert_eq(YAML.load_file('/tmp/output.yml'), YAML.load(%{
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

repo.run(%{
  export SEMAPHORE_GIT_REF_TYPE=branch

  #{spc} evaluate change-in --input .semaphore/semaphore.yml --output /tmp/output.yml --logs /tmp/logs.yml
})

assert_eq(YAML.load_file('/tmp/output.yml'), YAML.load(%{
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

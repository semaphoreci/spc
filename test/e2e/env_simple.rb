# rubocop:disable all

require_relative "../e2e"
require 'yaml'

pipeline = %q(
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: Test1
    run:
      when: "branch = 'master' or env('SEMAPHORE_GIT_PR_NAME') =~ '^docs:'"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"

  - name: Test2
    run:
      when: "branch = 'master' or env('SEMAPHORE_GIT_PR_NAME') =~ '^feature:'"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"

  - name: Test3
    run:
      when: "branch = 'master' or env('NON_EXISTENT_ENV_VAR') =~ '^feature:'"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"
)

origin = TestRepoForChangeIn.setup()

origin.add_file('.semaphore/semaphore.yml', pipeline)
origin.commit!("Bootstrap")

repo = origin.clone_local_copy(branch: "master")
repo.run(%{
  export SEMAPHORE_GIT_PR_NAME="docs: Change in deployment order"

  #{spc} evaluate change-in --input .semaphore/semaphore.yml --output /tmp/output.yml --logs /tmp/logs.yml
})

assert_eq(YAML.load_file('/tmp/output.yml'), YAML.load(%{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: Test1
    run:
      when: "(branch = 'master') or true"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"

  - name: Test2
    run:
      when: "(branch = 'master') or false"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"

  - name: Test3
    run:
      when: "(branch = 'master') or false"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"
}))

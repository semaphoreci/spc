# rubocop:disable all

require_relative "../e2e"
require 'yaml'

pipeline = %{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: Test
    run:
      when: "branch = 'master' and change_in('../java/')"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"

  - name: Test2
    run:
      when: "branch = 'master' and change_in('../javascript/')"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"
}

origin = TestRepoForChangeIn.setup()

origin.add_file('.semaphore/semaphore.yml', pipeline)
origin.add_file("javascript/test.js", "hello")
origin.add_file("java/test.java", "hello")
origin.commit!("Bootstrap")

origin.create_branch("dev")
origin.add_file("javascript/test.js", "hello hello")
origin.commit!("Changes in dev")

repo = origin.clone_local_copy(branch: "dev")
repo.run("#{spc} compile --input .semaphore/semaphore.yml --output /tmp/output.yml --logs /tmp/logs.yml")

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
      when: "(branch = 'master') and true"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"
}))

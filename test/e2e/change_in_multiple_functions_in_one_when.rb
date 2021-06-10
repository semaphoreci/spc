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
      when: "change_in('/lib1') or change_in('/lib2') or change_in('/lib3') or change_in('/lib4')"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"

  - name: Test2
    run:
      when: "change_in('/lib1') and change_in('/lib2') and change_in('/lib3') and change_in('/lib4')"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"
}

origin = TestRepoForChangeIn.setup()

origin.add_file('.semaphore/semaphore.yml', pipeline)
origin.commit!("Bootstrap")

origin.add_file("lib1/A.txt", "hello")
origin.add_file("lib2/A.txt", "hello")
origin.add_file("lib3/A.txt", "hello")
origin.add_file("lib4/A.txt", "hello")
origin.commit!("Changes on master")

origin.create_branch("dev")
origin.add_file("lib2/A.txt", "hello hello")
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
      when: "true"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"

  - name: Test2
    run:
      when: "false"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"
}))

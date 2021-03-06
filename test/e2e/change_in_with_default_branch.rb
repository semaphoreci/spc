# rubocop:disable all

require_relative "../e2e"

pipeline = %{
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
      when: "branch = 'master' and change_in(['/lib', 'log.txt'], {default_branch: 'dev'})"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"
}

origin = TestRepoForChangeIn.setup()

origin.add_file('.semaphore/semaphore.yml', pipeline)
origin.commit!("Bootstrap")

origin.add_file("app/A.txt", "hello")
origin.commit!("Changes on master")

origin.create_branch("main")
origin.add_file("lib/B.txt", "hello")
origin.commit!("Changes in main")

origin.create_branch("dev")
origin.add_file("app/C.txt", "hello")
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

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
      when: "branch = 'master' and change_in('/lib')"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"
}

origin = TestRepoForChangeIn.setup()

origin.add_file('.semaphore/semaphore.yml', pipeline)
origin.commit!("Bootstrap")

origin.add_file("lib/A.txt", "hello")
origin.commit!("Changes on master")

# diverge master and dev by 100 commits

origin.create_branch("dev")
350.times do |index|
  origin.add_file("lib/B#{index}.txt", "hello")
  origin.commit!("Changes in dev number #{index}")
end

origin.switch_branch("master")
350.times do |index|
  origin.add_file("lib/B#{index}.txt", "hello")
  origin.commit!("Changes in master number #{index}")
end

origin.switch_branch("dev")

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
      when: "(branch = 'master') and true"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"
}))

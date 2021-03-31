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

300.times do |index|
  origin.add_file("lib/B#{index}.txt", "hello")
  origin.commit!("Changes in master number #{index}")
end

repo = origin.clone_local_copy(branch: "master")
repo.run(%{
  export SEMAPHORE_GIT_COMMIT_RANGE="$(cd ../test-repo-origin && git rev-parse HEAD~95)...$(cd ../test-repo-origin && git rev-parse HEAD)"

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
}))

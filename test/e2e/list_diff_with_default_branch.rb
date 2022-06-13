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

repo.run("#{spc} list-diff --default-branch master > /tmp/output.txt")
output = File.readlines('/tmp/output.txt')
  .map { |line| line.strip }
  .reject { |line| line.empty? }
  
assert_eq($?.success?, true)
assert_eq(["app/C.txt", "lib/B.txt"], output)

repo.run("#{spc} list-diff --default-branch main > /tmp/output.txt")
output = File.readlines('/tmp/output.txt')
  .map { |line| line.strip }
  .reject { |line| line.empty? }
  
assert_eq($?.success?, true)
assert_eq(["app/C.txt"], output)

repo.run("#{spc} list-diff --default-branch dev > /tmp/output.txt")
output = File.readlines('/tmp/output.txt')
  .map { |line| line.strip }
  .reject { |line| line.empty? }
  
assert_eq($?.success?, true)
assert_eq([], output)

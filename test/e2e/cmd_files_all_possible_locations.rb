# rubocop:disable all

#
# This test verifies if the compiler is able to recognize and
# process all locations where a commands_file expression can appear.
#

require_relative "../e2e"
require 'yaml'

pipeline = %{
version: v1.0
name: "Tests"
agent:
  machine:
    type: "e1-standard-2"
    os_image: "ubuntu2004"
global_job_config:
  prologue:
    commands_file: "valid_commands_file.txt"
  epilogue:
    always:
      commands_file: "valid_commands_file.txt"
    on_pass:
      commands_file: "valid_commands_file.txt"
    on_fail:
      commands_file: "valid_commands_file.txt"
blocks:
  - name: Run tests
    task:
      prologue:
        commands_file: "valid_commands_file.txt"
      epilogue:
        always:
          commands_file: "valid_commands_file.txt"
        on_pass:
          commands_file: "valid_commands_file.txt"
        on_fail:
          commands_file: "valid_commands_file.txt"
      jobs:
        - name: Run tests
          commands_file: "valid_commands_file.txt"
after_pipeline:
  task:
    prologue:
      commands_file: "valid_commands_file.txt"
    epilogue:
      always:
        commands_file: "valid_commands_file.txt"
      on_pass:
        commands_file: "valid_commands_file.txt"
      on_fail:
        commands_file: "valid_commands_file.txt"
    jobs:
      - name: "Notify on Slack"
        commands_file: "valid_commands_file.txt"
}

commands_file = %{
echo 1
echo 12
echo 123
}

origin = TestRepoForChangeIn.setup()

origin.add_file('.semaphore/semaphore.yml', pipeline)
origin.add_file('.semaphore/valid_commands_file.txt', commands_file)
origin.commit!("Bootstrap")

repo = origin.clone_local_copy(branch: "master")
repo.run("#{spc} compile --input .semaphore/semaphore.yml --output /tmp/output.yml --logs /tmp/logs.yml")

assert_eq(YAML.load_file('/tmp/output.yml'), YAML.load(%{
    version: v1.0
    name: "Tests"
    agent:
      machine:
        type: "e1-standard-2"
        os_image: "ubuntu2004"
    global_job_config:
      prologue:
        commands:
          - echo 1
          - echo 12
          - echo 123
      epilogue:
        always:
          commands:
            - echo 1
            - echo 12
            - echo 123
        on_pass:
          commands:
            - echo 1
            - echo 12
            - echo 123
        on_fail:
          commands:
            - echo 1
            - echo 12
            - echo 123
    blocks:
      - name: Run tests
        task:
          prologue:
            commands:
              - echo 1
              - echo 12
              - echo 123
          epilogue:
            always:
              commands:
                - echo 1
                - echo 12
                - echo 123
            on_pass:
              commands:
                - echo 1
                - echo 12
                - echo 123
            on_fail:
              commands:
                - echo 1
                - echo 12
                - echo 123
          jobs:
            - name: Run tests
              commands:
                - echo 1
                - echo 12
                - echo 123
    after_pipeline:
      task:
        prologue:
          commands:
            - echo 1
            - echo 12
            - echo 123
        epilogue:
          always:
            commands:
              - echo 1
              - echo 12
              - echo 123
          on_pass:
            commands:
              - echo 1
              - echo 12
              - echo 123
          on_fail:
            commands:
              - echo 1
              - echo 12
              - echo 123
        jobs:
          - name: "Notify on Slack"
            commands:
              - echo 1
              - echo 12
              - echo 123
}))


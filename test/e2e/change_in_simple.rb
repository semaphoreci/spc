# rubocop:disable all

require_relative "../e2e"
require 'yaml'

system "rm -f /tmp/input.yml"
system "rm -f /tmp/output.yml"

File.write('/tmp/input.yml', %{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: Test
    skip:
      when: "branch = 'master' and change_in('lib')"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"
})

system('build/cli evaluate change-in --input /tmp/input.yml --output /tmp/output.yml --logs /tmp/logs.yml')

input = YAML.load_file('/tmp/input.yml')
output = YAML.load_file('/tmp/output.yml')

input["blocks"][0]["skip"]["when"] = "(branch = 'master') and false"

assert_eq(input, output)

# rubocop:disable all

require 'yaml'

File.write('/tmp/input.yml', %{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  - name: Test
    skip:
      when: "change_in('lib')"
})

system('build/cli evaluate change-in --input /tmp/input.yml --output /tmp/output.yml --logs /tmp/logs.yml')

output = YAML.load_file('/tmp/output.yml')

raise "failure" if output["blocks"][0]["skip"]["when"] != "change_in('lib')"

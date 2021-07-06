# rubocop:disable all

#
# This test verifies that copiler can process all possible locations for both the
# parameters and changs in the same yaml file.
#

require_relative "../e2e"
require 'yaml'

pipeline = %{
version: v1.0
name: "Deploy to ${{parameters.SERVICE}} to ${{parameters.DEPLOY_ENV}}"
agent:
  machine:
    type: e1-standard-2

fail_fast:
  cancel:
    when: "branch = 'master' and change_in('/lib')"
  stop:
    when: "branch = 'master' and change_in('/app')"

auto_cancel:
  queued:
    when: "branch = 'master' and change_in('/lib')"
  running:
    when: "branch = 'master' and change_in('/app')"

global_job_config:
  priority:
    - value: 1
      when: "branch = 'master' and change_in('/lib')"
  secrets:
    - name: "${{parameters.DEPLOY_ENV}}_github"
    - name: "github_key"

queue:
  - name: "${{parameters.DEPLOY_ENV}}_deployment_queue"
    when: "branch = 'master' and change_in('/lib')"

  - name: "${{parameters.MISSING}}_queue"
    when: "branch = 'master' and change_in('/app')"

  - name: "default_queue"
    when: true

blocks:
  - name: Build and push image
    skip:
      when: "branch = 'master' and change_in('/lib')"
    task:
      secrets:
        - name: ${{parameters.DEPLOY_ENV}}_dockerhub
        - name: ${{parameters.DEPLOY_ENV}}_ecr
      jobs:
        - name: Build & Push
          priority:
            - value: 1
              when: "branch = 'master' and change_in('/lib')"
          commands:
            - make build
            - make push

  - name: Deploy image
    run :
      when: "branch = 'master' and change_in('/app')"
    task:
      secrets:
        - name: ${{parameters.DEPLOY_ENV}}_deploy_key
        - name: ${{parameters.DEPLOY_ENV}}_aws_creds
      jobs:
        - name: Deploy
          commands:
            - make deploy

promotions:
  - name: Performance tests
    auto_promote:
      when: "branch = 'master' and change_in('/lib')"
}

origin = TestRepoForChangeIn.setup()

origin.add_file('.semaphore/semaphore.yml', pipeline)
origin.commit!("Bootstrap")

origin.add_file("lib/A.txt", "hello")
origin.commit!("Changes on master")

origin.create_branch("dev")
origin.add_file("lib/B.txt", "hello")
origin.commit!("Changes in dev")

repo = origin.clone_local_copy(branch: "dev")

repo.run(%{
  export SERVICE=web_server
  export DEPLOY_ENV=prod

  #{spc} compile \
     --input .semaphore/semaphore.yml \
     --output /tmp/output.yml \
     --logs /tmp/logs.yml
})

assert_eq(YAML.load_file('/tmp/output.yml'), YAML.load(%{
version: v1.0
name: "Deploy to web_server to prod"
agent:
  machine:
    type: e1-standard-2

fail_fast:
  cancel:
    when: "(branch = 'master') and true"
  stop:
    when: "(branch = 'master') and false"

auto_cancel:
  queued:
    when: "(branch = 'master') and true"
  running:
    when: "(branch = 'master') and false"

global_job_config:
  priority:
    - value: 1
      when: "(branch = 'master') and true"
  secrets:
    - name: "prod_github"
    - name: "github_key"

queue:
  - name: "prod_deployment_queue"
    when: "(branch = 'master') and true"

  - name: "MISSING_queue"
    when: "(branch = 'master') and false"

  - name: "default_queue"
    when: true

blocks:
  - name: Build and push image
    skip:
      when: "(branch = 'master') and true"
    task:
      secrets:
        - name: prod_dockerhub
        - name: prod_ecr
      jobs:
        - name: Build & Push
          priority:
            - value: 1
              when: "(branch = 'master') and true"
          commands:
            - make build
            - make push

  - name: Deploy image
    run :
      when: "(branch = 'master') and false"
    task:
      secrets:
        - name: prod_deploy_key
        - name: prod_aws_creds
      jobs:
        - name: Deploy
          commands:
            - make deploy

promotions:
  - name: Performance tests
    auto_promote:
      when: "(branch = 'master') and true"
}))

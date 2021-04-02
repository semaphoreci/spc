# rubocop:disable all

require_relative "../e2e"
require 'yaml'

pipeline = %q(
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  # Build docker image in a Linux VM

  - name: Build
    agent:
      machine:
        type: e1-standard-2
    task:
      jobs:
        - name: Build
          commands:
            - make docker.build TAG=dev-env-$SEMAPHORE_GIT_COMMIT_SHA
            - make docker.push

  # Run tests in a Docker image

  - name: Test
    agent:
      machine:
        type: e1-standard-2
      containers:
        - name: main
          image: "dev-env-$SEMAPHORE_GIT_COMMIT_SHA"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"
)

origin = TestRepoForChangeIn.setup()

origin.add_file('.semaphore/semaphore.yml', pipeline)
origin.commit!("Bootstrap")

repo = origin.clone_local_copy(branch: "master")
repo.run(%{
  export SEMAPHORE_GIT_COMMIT_SHA="055556799236d27ab754b503f2acad3e9a29350f"

  #{spc} evaluate change-in --input .semaphore/semaphore.yml --output /tmp/output.yml --logs /tmp/logs.yml
})

assert_eq(YAML.load_file('/tmp/output.yml'), YAML.load(%{
version: v1.0
name: Test
agent:
  machine:
    type: e1-standard-2

blocks:
  # Build docker image in a Linux VM

  - name: Build
    agent:
      machine:
        type: e1-standard-2
    task:
      jobs:
        - name: Build
          commands:
            - make docker.build TAG=dev-env-$SEMAPHORE_GIT_COMMIT_SHA
            - make docker.push

  # Run tests in a Docker image

  - name: Test
    agent:
      machine:
        type: e1-standard-2
      containers:
        - name: main
          image: "dev-env-055556799236d27ab754b503f2acad3e9a29350f"
    task:
      jobs:
        - name: Hello
          commands:
            - echo "Hello World"
}))

# rubocop:disable all

#
# The TestRepoForChangeIn initializes a repository on your local
# machine. It creates the equivalent of a GitHub hosted "origin"
# repository. You will need to "clone" before using it for testing.
#
# Usage:
#
# 1. Create an 'origin':
#
#   origin = TestRepoForChangeIn.setup()
#
# 2. Add a file:
#
#   origin.add_file("README.md", "hello")
#
# 3. Commit:
#
#   origin.commit!("Bootstrap")
#
# 4. Clone the repository:
#
#   local_copy = origin.clone_local_copy(branch: "master")
#
#   local_copy.list_branches()
#

class TestRepoForChangeIn
  def self.setup
    path = "/tmp/test-repo-origin"

    system "rm -rf #{path}"
    system "mkdir -p #{path}"

    repo = new(path)

    repo.run(%{
      git init
      mkdir .semaphore
    })

    repo
  end

  def initialize(path)
    @path = path
  end

  # Run a command in the context or the repository
  def run(commands, options = {})
    system %{
      cd #{@path}

      #{commands}
    }

    if options[:fail] != false
      raise "Failed to execute command" if $?.exitstatus != 0
    end
  end

  def list_branches
    run "git branch -a"
  end

  def switch_branch(name)
    run("git checkout #{name}")
  end

  def create_branch(name)
    run("git checkout -b #{name}")
  end

  def merge_branch(name)
    run("git merge --no-edit #{name}")
  end

  def add_file(file_path, content)
    full_path = File.join(@path, file_path)
    system "mkdir -p #{File.dirname(full_path)}"

    File.write(full_path, content)
  end

  def commit!(message)
    run("git add . && git commit -m '#{message}'")
  end

  def clone_local_copy(options = {})
    clone_path = "/tmp/test-repo"

    system "rm -rf #{clone_path}"
    system "git clone #{@path} --branch #{options[:branch]} #{clone_path}"

    repo = TestRepoForChangeIn.new(clone_path)

    repo
  end
end

# rubocop:disable all

class TestRepoForChangeIn
  def self.setup(path)
    path = "/tmp/test-repo-#{origin}"

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
  def run(commands)
    system %{
      cd #{@path}

      #{commands}
    }
  end

  def switch_branch(name)
    run("git checkout -b #{name}")
  end

  def add_file(file_path, content)
    File.write(Path.join(@path, file_path), content)
  end

  def commit!(message)
    run("git add . && git commit -m '#{message}'")
  end

  def clone_local_copy(options = {})
    clone_path = "/tmp/test-repo"

    system "rm -rf #{clone_path}"
    system "git clone #{@path} --branch #{options[:branch]} #{clone_path}"

    repo = new(clone_path)

    repo
  end
end

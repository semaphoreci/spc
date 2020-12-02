# rubocop:disable all

require_relative "./e2e_utils/test_repo_for_change_in"
require 'yaml'
require 'json'

def spc
  `echo "$(pwd)/build/cli"`.strip
end

def assert(bool)
  raise "failed" unless bool
end

def blue(msg)
  "\e[34m#{msg}\e[0m"
end

def red(msg)
  "\e[31m#{msg}\e[0m"
end

def green(msg)
  "\e[32m#{msg}\e[0m"
end

puts blue ""
puts blue "Running: #{$PROGRAM_NAME}"
puts blue "="*120
puts blue ""

def assert_eq(a, b)
  puts ""
  puts blue("Assert Equal in #{$PROGRAM_NAME}")
  puts ""
  puts blue "  Left: #{a.inspect[0..100]}"
  puts blue " Right: #{b.inspect[0..100]}"
  puts ""

  if Diff.compare(a, b)
    puts green "PASSED"
  else
    puts red "FAILURE"
    puts Diff.diff(a, b)
    exit(1)
  end
end

class Diff

  def self.compare(a, b)
    if a.is_a?(Hash) && b.is_a?(Hash)
      (a.keys + b.keys).each do |k|
        return false unless compare(a[k], b[k])
      end
    end

    if a.is_a?(Array) && b.is_a?(Array)
      if a.size != b.size
        return false
      end

      a.size.times.each do |i|
        return false unless compare(a[i], b[i])
      end
    end

    a == b
  end

  def self.diff(a, b, path = [])
    if a.is_a?(Hash) && b.is_a?(Hash)
      return (a.keys + b.keys).uniq.map do |k|
        diff(a[k], b[k], path + [k])
      end
    end

    if a.is_a?(Array) && b.is_a?(Array)
      if a.size != b.size
        return ["array sizes do not match"]
      end

      return a.size.times.map do |i|
        diff(a[i], b[i], path + [i])
      end
    end

    if a != b
      return ["#{path}:\n#{a.inspect}\n#{b.inspect}"]
    end

    []
  end

end

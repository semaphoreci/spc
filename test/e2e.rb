# rubocop:disable all

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

  if a == b
    puts green "PASSED"
  else
    puts HashDiff.new(a, b).report
    puts red "FAILURE"
    exit(1)
  end
end

class HashDiff

  attr_reader :left, :right

  def initialize(left, right, config = {}, path = nil)
    @left  = left
    @right = right
    @config = config
    @path = path
    @conformity = 0
  end

  def conformity
    find_differences
    @conformity
  end

  def report
    @config[:report] = true
    find_differences
  end

  def find_differences
    if hash?(left) && hash?(right)
      compare_hashes_keys
    elsif left.is_a?(Array) && right.is_a?(Array)
      compare_arrays
    else
      report_diff
    end
  end

  def compare_hashes_keys
    combined_keys.each do |key|
      l = value_with_default(left, key)
      r = value_with_default(right, key)
      if l == r
        @conformity += 100
      else
        compare_sub_items l, r, key
      end
    end
  end

  private

  def compare_sub_items(l, r, key)
    diff = self.class.new(l, r, @config, path(key))
    @conformity += diff.conformity
  end

  def report_diff
    return unless @config[:report]

    puts "#{@path}:"
    puts "- #{left}" unless left == NO_VALUE
    puts "+ #{right}" unless right == NO_VALUE
  end

  def combined_keys
    (left.keys + right.keys).uniq
  end

  def hash?(value)
    value.is_a?(Hash)
  end

  def compare_arrays
    l, r = left.clone, right.clone
    l.each_with_index do |l_item, l_index|
      max_item_index = nil
      max_conformity = 0
      r.each_with_index do |r_item, i|
        if l_item == r_item
          @conformity += 1
          r[i] = TAKEN
          break
        end

        diff = self.class.new(l_item, r_item, {})
        c = diff.conformity
        if c > max_conformity
          max_conformity = c
          max_item_index = i
        end
      end or next

      if max_item_index
        key = l_index == max_item_index ? l_index : "#{l_index}/#{max_item_index}"
        compare_sub_items l_item, r[max_item_index], key
        r[max_item_index] = TAKEN
      else
        compare_sub_items l_item, NO_VALUE, l_index
      end
    end

    r.each_with_index do |item, index|
      compare_sub_items NO_VALUE, item, index unless item == TAKEN
    end
  end

  def path(key)
    p = "#{@path} > " if @path
    "#{p}#{key}"
  end

  def value_with_default(obj, key)
    obj.fetch(key, NO_VALUE)
  end

  module NO_VALUE; end
  module TAKEN; end

end

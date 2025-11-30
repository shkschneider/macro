99.downto(1) do |n|
  puts "#{n} bottle#{'s' unless n==1} of beer on the wall, #{n} bottle#{'s' unless n==1} of beer."
  if n-1 > 0
    puts "Take one down and pass it around, #{n-1} bottle#{'s' unless n-1==1} of beer on the wall.\n"
  else
    puts "Take one down and pass it around, no more bottles of beer on the wall.\n"
  end
end
puts "No more bottles of beer on the wall, no more bottles of beer."
puts "Go to the store and buy some more, 99 bottles of beer on the wall."

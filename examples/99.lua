for n = 99, 1, -1 do
  print(n .. " bottle" .. (n==1 and "" or "s") .. " of beer on the wall, " .. n .. " bottle" .. (n==1 and "" or "s") .. " of beer.")
  if n-1 > 0 then
    print("Take one down and pass it around, " .. (n-1) .. " bottle" .. ((n-1)==1 and "" or "s") .. " of beer on the wall.\n")
  else
    print("Take one down and pass it around, no more bottles of beer on the wall.\n")
  end
end
print("No more bottles of beer on the wall, no more bottles of beer.")
print("Go to the store and buy some more, 99 bottles of beer on the wall.")

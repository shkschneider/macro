for (n in 99:1) {
  cat(n, "bottle", ifelse(n==1, "", "s"), "of beer on the wall,", n, "bottle", ifelse(n==1, "", "s"), "of beer.\n")
  if (n-1 > 0) cat("Take one down and pass it around,", n-1, "bottle", ifelse(n-1==1, "", "s"), "of beer on the wall.\n\n")
  else cat("Take one down and pass it around, no more bottles of beer on the wall.\n\n")
}
cat("No more bottles of beer on the wall, no more bottles of beer.\n")
cat("Go to the store and buy some more, 99 bottles of beer on the wall.\n")

for n in (1...99).reversed() {
  print("\(n) bottle\(n==1 ? "" : "s") of beer on the wall, \(n) bottle\(n==1 ? "" : "s") of beer.")
  if n-1 > 0 {
    print("Take one down and pass it around, \(n-1) bottle\(n-1==1 ? "" : "s") of beer on the wall.\n")
  } else {
    print("Take one down and pass it around, no more bottles of beer on the wall.\n")
  }
}
print("No more bottles of beer on the wall, no more bottles of beer.")
print("Go to the store and buy some more, 99 bottles of beer on the wall.")

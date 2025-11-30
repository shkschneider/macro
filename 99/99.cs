using System;
class B {
  static void Main() {
    for (int n=99;n>0;n--) {
      Console.WriteLine("{0} bottle{1} of beer on the wall, {0} bottle{1} of beer.", n, n==1?"":"s");
      if (n-1>0)
        Console.WriteLine("Take one down and pass it around, {0} bottle{1} of beer on the wall.\n", n-1, n-1==1?"":"s");
      else
        Console.WriteLine("Take one down and pass it around, no more bottles of beer on the wall.\n");
    }
    Console.WriteLine("No more bottles of beer on the wall, no more bottles of beer.");
    Console.WriteLine("Go to the store and buy some more, 99 bottles of beer on the wall.");
  }
}

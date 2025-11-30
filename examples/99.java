public class Bottles {
  public static void main(String[] args) {
    for (int n = 99; n > 0; n--) {
      System.out.printf("%d bottle%s of beer on the wall, %d bottle%s of beer.%n", n, n==1?"":"s", n, n==1?"":"s");
      if (n-1 > 0)
        System.out.printf("Take one down and pass it around, %d bottle%s of beer on the wall.%n%n", n-1, n-1==1?"":"s");
      else
        System.out.println("Take one down and pass it around, no more bottles of beer on the wall.\n");
    }
    System.out.println("No more bottles of beer on the wall, no more bottles of beer.");
    System.out.println("Go to the store and buy some more, 99 bottles of beer on the wall.");
  }
}

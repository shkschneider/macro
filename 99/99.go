package main
import "fmt"
func main() {
  for n := 99; n > 0; n-- {
    fmt.Printf("%d bottle%s of beer on the wall, %d bottle%s of beer.\n", n, plural(n), n, plural(n))
    if n-1 > 0 {
      fmt.Printf("Take one down and pass it around, %d bottle%s of beer on the wall.\n\n", n-1, plural(n-1))
    } else {
      fmt.Println("Take one down and pass it around, no more bottles of beer on the wall.\n")
    }
  }
  fmt.Println("No more bottles of beer on the wall, no more bottles of beer.")
  fmt.Println("Go to the store and buy some more, 99 bottles of beer on the wall.")
}
func plural(n int) string { if n==1 { return "" }; return "s" }

fn main() {
    for n in (1..=99).rev() {
        println!("{} bottle{} of beer on the wall, {} bottle{} of beer.", n, if n==1 {""} else {"s"}, n, if n==1 {""} else {"s"});
        if n-1 > 0 {
            println!("Take one down and pass it around, {} bottle{} of beer on the wall.\n", n-1, if n-1==1 {""} else {"s"});
        } else {
            println!("Take one down and pass it around, no more bottles of beer on the wall.\n");
        }
    }
    println!("No more bottles of beer on the wall, no more bottles of beer.");
    println!("Go to the store and buy some more, 99 bottles of beer on the wall.");
}

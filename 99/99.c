#include <stdio.h>

int main(void) {
    for (int n = 99; n >= 1; --n) {
        printf("%d bottle%s of beer on the wall, %d bottle%s of beer.\n", n, n==1?"":"s", n, n==1?"":"s");
        if (n-1 > 0)
            printf("Take one down and pass it around, %d bottle%s of beer on the wall.\n\n", n-1, n-1==1?"":"s");
        else
            printf("Take one down and pass it around, no more bottles of beer on the wall.\n\n");
    }
    puts("No more bottles of beer on the wall, no more bottles of beer.");
    puts("Go to the store and buy some more, 99 bottles of beer on the wall.");
    return 0;
}

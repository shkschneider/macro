def sing():
    for n in range(99, 0, -1):
        print(f"{n} bottle{'s' if n!=1 else ''} of beer on the wall, {n} bottle{'s' if n!=1 else ''} of beer.")
        nxt = n-1
        if nxt > 0:
            print(f"Take one down and pass it around, {nxt} bottle{'s' if nxt!=1 else ''} of beer on the wall.\n")
        else:
            print("Take one down and pass it around, no more bottles of beer on the wall.\n")
    print("No more bottles of beer on the wall, no more bottles of beer.")
    print("Go to the store and buy some more, 99 bottles of beer on the wall.")
if __name__ == '__main__':
    sing()

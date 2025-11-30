#include <iostream>
using namespace std;
int main(){
  for(int n=99;n>0;--n){
    cout<<n<<" bottle"<<(n==1?"":"s")<<" of beer on the wall, "<<n<<" bottle"<<(n==1?"":"s")<<" of beer.\n";
    if(n-1>0) cout<<"Take one down and pass it around, "<<n-1<<" bottle"<<(n-1==1?"":"s")<<" of beer on the wall.\n\n";
    else cout<<"Take one down and pass it around, no more bottles of beer on the wall.\n\n";
  }
  cout<<"No more bottles of beer on the wall, no more bottles of beer.\n";
  cout<<"Go to the store and buy some more, 99 bottles of beer on the wall.\n";
}

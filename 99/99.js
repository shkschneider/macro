for (let n = 99; n > 0; n--) {
  console.log(`${n} bottle${n!==1 ? 's' : ''} of beer on the wall, ${n} bottle${n!==1 ? 's' : ''} of beer.`);
  let nxt = n - 1;
  if (nxt > 0)
    console.log(`Take one down and pass it around, ${nxt} bottle${nxt!==1 ? 's' : ''} of beer on the wall.\n`);
  else
    console.log('Take one down and pass it around, no more bottles of beer on the wall.\n');
}
console.log('No more bottles of beer on the wall, no more bottles of beer.');
console.log('Go to the store and buy some more, 99 bottles of beer on the wall.');

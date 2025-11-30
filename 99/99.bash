#!/usr/bin/env bash
for ((n=99;n>0;n--)); do
  s="s"; [[ $n -eq 1 ]] && s=""
  echo "$n bottle$s of beer on the wall, $n bottle$s of beer."
  if (( n-1 > 0 )); then
    m=$((n-1)); ms="s"; [[ $m -eq 1 ]] && ms=""
    echo "Take one down and pass it around, $m bottle$ms of beer on the wall."
  else
    echo "Take one down and pass it around, no more bottles of beer on the wall."
  fi
  echo
done
echo "No more bottles of beer on the wall, no more bottles of beer."
echo "Go to the store and buy some more, 99 bottles of beer on the wall."

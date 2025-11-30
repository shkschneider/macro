for n <- 99..1 |> Enum.reverse do
  IO.puts("#{n} bottle#{if n==1, do: "", else: "s"} of beer on the wall, #{n} bottle#{if n==1, do: "", else: "s"} of beer.")
  if n-1 > 0 do
    IO.puts("Take one down and pass it around, #{n-1} bottle#{if n-1==1, do: "", else: "s"} of beer on the wall.\n")
  else
    IO.puts("Take one down and pass it around, no more bottles of beer on the wall.\n")
  end
end
IO.puts("No more bottles of beer on the wall, no more bottles of beer.")
IO.puts("Go to the store and buy some more, 99 bottles of beer on the wall.")

main :: IO ()
main = mapM_ putStrLn verses >> putStrLn ending
  where
    verses = concatMap verse [99,98..1]
    verse n = [line1 n, line2 n, ""]
    line1 n = show n ++ " bottle" ++ plural n ++ " of beer on the wall, " ++ show n ++ " bottle" ++ plural n ++ " of beer."
    line2 n = if n-1 > 0 then "Take one down and pass it around, " ++ show (n-1) ++ " bottle" ++ plural (n-1) ++ " of beer on the wall." else "Take one down and pass it around, no more bottles of beer on the wall."
    ending = "No more bottles of beer on the wall, no more bottles of beer.\nGo to the store and buy some more, 99 bottles of beer on the wall."
    plural n = if n==1 then "" else "s"

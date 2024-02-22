fn update-bin-index {|bin-dir|
  tmp pwd = $bin-dir
  for f [*/*] {
    echo 'https://dl.elv.sh/'$f
  } > INDEX
}

provider packernix {
  flake = "./testdata/eval"
}

data "packernix_eval" "installable_flake" {
  installable = "#fib10"
  clear_env = true
}

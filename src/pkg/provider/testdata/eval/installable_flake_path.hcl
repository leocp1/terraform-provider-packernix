provider packernix {
  flake_path = "testdata/eval"
}

data "packernix_eval" "installable_flake_path" {
  installable = "#fib10"
  clear_env = true
  working_dir = "./testdata"
}

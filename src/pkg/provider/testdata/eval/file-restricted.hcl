provider packernix {}

data "packernix_eval" "file" {
  file = "./testdata/eval/fib.nix"
  arg = {
    "x" = "10"
  }
  attr = "out"
  clear_env = true
}

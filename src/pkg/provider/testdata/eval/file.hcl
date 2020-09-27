provider packernix {}

data "packernix_eval" "file" {
  file = "./testdata/eval/fib.nix"
  arg = {
    "x" = "10"
  }
  attr = "out"
  # Necessary since we are in restrict-eval by default
  env = {
    "NIX_PATH" = "."
  }
  clear_env = true
}

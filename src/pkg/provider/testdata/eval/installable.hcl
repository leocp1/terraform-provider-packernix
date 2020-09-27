provider packernix {}

data "packernix_eval" "installable" {
  installable = "./testdata/eval#fib10"
  clear_env = true
}

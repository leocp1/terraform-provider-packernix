provider packernix {}

data "packernix_external" "stderr" {
  path = "./testdata/external/stderr"
}

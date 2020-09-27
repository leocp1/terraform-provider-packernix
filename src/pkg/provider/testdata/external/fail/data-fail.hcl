provider packernix {}

data "packernix_external" "fail" {
  path = "./testdata/external/fail"
}

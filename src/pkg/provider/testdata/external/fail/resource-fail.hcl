provider packernix {}

resource "packernix_external" "fail" {
  path = "./testdata/external/fail"
}

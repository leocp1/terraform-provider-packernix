provider packernix {}

data "packernix_external" "noop" {
  path = "./testdata/external/noop"
}

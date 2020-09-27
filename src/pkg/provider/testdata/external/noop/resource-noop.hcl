provider packernix {}

resource "packernix_external" "noop" {
  path = "./testdata/external/noop"
}

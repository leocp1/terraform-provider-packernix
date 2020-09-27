provider packernix {}

data "packernix_external" "pwd" {
  path = "./testdata/external/pwd"
}

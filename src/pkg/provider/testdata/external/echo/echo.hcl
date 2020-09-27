provider packernix {}

data "packernix_external" "echo" {
  path = "./testdata/external/echo"
  options = ["test"]
}

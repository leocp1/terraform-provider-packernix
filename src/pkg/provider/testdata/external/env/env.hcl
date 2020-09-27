provider packernix {
  env = {
    "TEST" = "fail"
  }
}

data "packernix_external" "env" {
  path = "./testdata/external/env"
  env = {
    "TEST" = "testid"
  }
}

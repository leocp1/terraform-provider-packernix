provider packernix {
  env = {
    "TEST" = "testid"
  }
}

data "packernix_external" "env" {
  path = "./testdata/external/env"
}

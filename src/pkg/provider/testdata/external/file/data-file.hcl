provider packernix {}

data "packernix_external" "file" {
  path = "./testdata/external/file"
  options = ["./testdata/external/file/data.txt", "testdata"]
}

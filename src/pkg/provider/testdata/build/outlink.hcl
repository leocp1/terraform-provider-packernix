provider packernix {}

data "packernix_eval" "nixpkgs" {
  inline = file("./testdata/build/static-haskell-nixpkgs.nix")
  nix_options = {
    "allowed-uris" = "https://github.com"
    "restrict-eval" = "true"
  }
}

data "packernix_build" "outlink" {
  file = "./testdata/build/default.nix"
  clear_env = true
  env = {
    "HOME" = "/homeless-shelter"
    "NIX_PATH" = "."
  }
  nixpkgs = jsondecode(data.packernix_eval.nixpkgs.out)
  attr = "hello"
  out_link = "{{.TempDir}}/result"
}

provider packernix {}

data "packernix_eval" "nixpkgs" {
  inline = file("./testdata/build/static-haskell-nixpkgs.nix")
  nix_options = {
    "allowed-uris" = "https://github.com"
    "restrict-eval" = "true"
  }
}

data "packernix_build" "file" {
  file = "./testdata/build/default.nix"
  clear_env = true
  env = {
    "HOME" = "/homeless-shelter"
    "NIX_PATH" = "."
  }
  attr = "hello"
  nixpkgs = jsondecode(data.packernix_eval.nixpkgs.out)
}

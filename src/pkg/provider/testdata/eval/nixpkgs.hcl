provider packernix {}

data "packernix_eval" "nixpkgs" {
  inline = file("./testdata/build/static-haskell-nixpkgs.nix")
  nix_options = {
    "allowed-uris" = "https://github.com"
    "restrict-eval" = "true"
  }
  out_link = "{{.TempDir}}/nixpkgs-eval"
}

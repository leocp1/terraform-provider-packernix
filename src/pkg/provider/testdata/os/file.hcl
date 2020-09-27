provider packernix {}

data "packernix_eval" "nixpkgs" {
  inline = file("./testdata/os/nixpkgs-20.03.nix")
  nix_options = {
    "allowed-uris" = "https://github.com"
    "restrict-eval" = "true"
  }
}

data "packernix_os" "file" {
  file = "./testdata/os/configuration.nix"
  clear_env = true
  env = {
    "HOME" = "/homeless-shelter"
    "NIX_PATH" = "."
  }
  config = jsonencode({
	"hostName" = "tfpnhost"
  })
  nixpkgs = jsondecode(data.packernix_eval.nixpkgs.out)
  out_link = "{{.TempDir}}/result-file"
}

provider packernix {}

data "packernix_eval" "nixpkgs" {
  inline = file("./testdata/os/nixpkgs-20.03.nix")
  nix_options = {
    "allowed-uris" = "https://github.com"
    "restrict-eval" = "true"
  }
}

data "packernix_os" "installable" {
  installable = "#nixosConfiguration"
  flake_path = "./testdata/os"
  clear_env = true
  env = {
    "NIX_PATH" = "."
  }
  config = jsonencode({
	"hostName" = "tfpnhost"
  })
  nixpkgs = jsondecode(data.packernix_eval.nixpkgs.out)
  out_link = "{{.TempDir}}/result-installable"
}

provider packernix{}

data "packernix_eval" "nixpkgs" {
  file = "./testdata/os/nixpkgs-20.03.nix"
  env = {
    "HOME" = "/homeless-shelter"
    "NIX_PATH" = "."
  }
  nix_options = {
    "allowed-uris" = "https://github.com"
    "restrict-eval" = "true"
  }
}

data "packernix_os" "nixos" {
  file = "./testdata/os/{{ .P }}.nix"
  clear_env = true
  env = {
    "HOME" = "/homeless-shelter"
    "NIX_PATH" = "."
  }
  nixpkgs = jsondecode(data.packernix_eval.nixpkgs.out)
}

data "packernix_const" "out" {}

data "packernix_eval" "template" {
  file = "${data.packernix_const.out.packer}/{{ .P }}.nix"
  env = {
    "HOME" = "/homeless-shelter"
    "NIX_PATH" = "${data.packernix_const.out.packer}"
  }
  argstr = {
    nixos = data.packernix_os.nixos.out_path
  }
}

resource "packernix_image" "image" {
  template = data.packernix_eval.template.out
}

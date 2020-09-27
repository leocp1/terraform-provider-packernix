{ lib
, callPackage
, packer
, packerPlugins
, nix
, terraform
}:
let
  packerFakessh = packer.withPlugins [ packerPlugins.provisioner-fakessh ];
  provider = callPackage ./src { inherit nix terraform; };
  packer = with import ./packer/lib.nix; path;
  nixos = ./nixos;
in
provider.overrideAttrs (
  oldAttrs: rec {
    postInstall = ''
      mkdir -p $out/share
      ln -sf ${packer} $out/share/packer
      ln -sf ${nixos} $out/share/nixos
      ln -sf $out/bin/terraform-provider-packernix{,_v${oldAttrs.version}}
    '';
    passthru = oldAttrs.passthru // {
      modulesPath = "${nixos}/modules";
      provider-source-address = "registry.terraform.io/leocp1/packernix";
    };
  }
)

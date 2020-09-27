final: prev: rec {
  # DELETE ME
  inherit (prev.callPackage ./srcPkgSupport.nix {})
    gitDescribe nixFilter
    ;

  gitignoreSource = let
    gisrc = prev.fetchFromGitHub {
      owner = "hercules-ci";
      repo = "gitignore";
      rev = "647d0821b590ee96056f4593640534542d8700e5";
      sha256 = "0ks37vclz2jww9q0fvkk9jyhscw0ial8yx2fpakra994dm12yy1d";
    };
  in
    (prev.callPackage gisrc {}).gitignoreSource;

  packer = let
    packerExtender = prev.callPackage
      ./pkgs/development/tools/packer/packer-extender.nix {};
  in
    packerExtender prev.packer;
  packerPlugins =
    prev.callPackage ./pkgs/development/tools/packer-plugins {};
  inherit (prev.callPackage ./pkgs/applications/networking/cluster/terraform {})
    terraform_0_11
    terraform_0_11-full
    terraform_0_12
    terraform_0_13
    terraform_plugins_test
    ;

  terraform = terraform_0_12;

  terraform-providers = (
    prev.recurseIntoAttrs (
      prev.callPackage
        pkgs/applications/networking/cluster/terraform-providers {}
    )
  ) // {
    packernix = prev.callPackage ./default.nix {
      terraform = final.terraform_0_13;
      nix = final.nixFlakes;
    };
  };
}

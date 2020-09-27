# Generated by terraform-provider-packernix
# Compare to lib.nixosSystem

{
  inputs = {
    base.url = "{{.Flake}}";

    nixpkgs.url = "{{.Nixpkgs}}";

    nixpkgsRaw = {
      url = "{{.Nixpkgs}}";
      flake = false;
    };

    tfpnModules = {
      url = "{{.TfpnModPath}}";
      flake = false;
    };
  };

  outputs = { self, base, nixpkgs, nixpkgsRaw, tfpnModules, ... }:
    let
      tfpnModulesPath = tfpnModules.outPath;
      modulesPath = "${nixpkgsRaw.outPath}/nixos/modules";
      baseModules = import "${modulesPath}/module-list.nix";
    in
      rec {
        nixosModules = {
          tfpn = { config, lib, ... }: {
            options.tfpn = lib.mkOption {
              description = "Option passed from terraform-provider-packernix";
            };
            config.tfpn = builtins.fromJSON (builtins.readFile ./tfpn-config.json);
          };
          pkgs = { config, lib, ... }: rec {
            config = {
              nixpkgs = {
                system = lib.mkDefault builtins.currentSystem;
                initialSystem = builtins.currentSystem;
              };
              system = {
                nixos = {
                  versionSuffix = ".${lib.substring 0 8 (self.lastModifiedDate or
                    self.lastModified or "19700101")}.${self.shortRev or "dirty"}";
                revision = lib.mkIf (self ? rev) self.rev;
                };
              };
            };
          };
        };

        nixosConfiguration = nixpkgs.lib.evalModules {
          modules = [
            nixosModules.tfpn
            nixosModules.pkgs
            base.{{.Attr}}
          ];
          specialArgs = {
            inherit modulesPath baseModules tfpnModulesPath;
          };
        };
      };
}
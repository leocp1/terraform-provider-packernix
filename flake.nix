{
  description = "Terraform Provider for Nix with Packer";

  outputs = { self, nixpkgs, ... }:
    let
      supportedSystems = [ "x86_64-linux" ];
      forAllSystems = f: nixpkgs.lib.genAttrs supportedSystems (system: f system);
    in
      rec {
        overlay = import ./overlay.nix;
        nixosModules = import ./nixos/modules/modules.nix;
        legacyPackages = forAllSystems (
          system: (
            import nixpkgs {
              inherit system;
              overlays = [ self.overlay ];
            }
          )
        );
        devShell = forAllSystems (
          system: let
            pkgs = self.legacyPackages.${system};
          in
            import ./shell.nix { inherit pkgs; }
        );
      };
}

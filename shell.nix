let
  pkgs' = (import <nixpkgs> { overlays = [ (import ./overlay.nix) ]; });
in
{ pkgs ? pkgs' }:
  with pkgs; let
    packer' = packer.withPlugins (
      p: with p; [
        builder-vultr
        builder-delete-vultr
        # builder-delete-amazon-ebs
        provisioner-fakessh
      ]
    );
    terraform' = terraform_0_13.withPlugins (
      p: with p; [
        vultr
        packernix
      ]
    );
  in
    mkShell {
      buildInputs = [
        packer'
        terraform'
      ];
    }

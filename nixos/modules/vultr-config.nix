{ config, pkgs, lib, modulesPath, ... }:
{
  # Set system
  nixpkgs.localSystem.system = "x86_64-linux";

  # From hardware-configuration.nix output of nixos-generate-config
  imports =
    [
      "${modulesPath}/profiles/qemu-guest.nix"
    ];
  boot.initrd.availableKernelModules = lib.mkDefault [ "ata_piix" "uhci_hcd" "virtio_pci" "sr_mod" "virtio_blk" ];
  boot.initrd.kernelModules = lib.mkDefault [ ];
  boot.kernelModules = lib.mkDefault [ ];
  boot.extraModulePackages = lib.mkDefault [ ];
  swapDevices = lib.mkDefault [ ];
  nix.maxJobs = lib.mkDefault 1;
  fileSystems."/" = lib.mkDefault {
    device = "/dev/disk/by-label/nixos";
    fsType = "ext4";
  };

  # Recommended by vultr
  boot.loader.grub.device = lib.mkDefault "/dev/vda";

  users.mutableUsers = lib.mkDefault true;

  # ssh settings
  services.openssh = lib.mkDefault {
    enable = true;
    challengeResponseAuthentication = false;
    extraConfig = ''
      AllowUsers root
    '';
    passwordAuthentication = false;
  };
}

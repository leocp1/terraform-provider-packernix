{ config, tfpnModulesPath, baseModules, ... }:
{
  imports = baseModules ++ [
    "${tfpnModulesPath}/vultr-config.nix"
  ];
  networking.hostName = config.tfpn.hostName;
}

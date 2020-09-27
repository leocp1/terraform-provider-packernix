{ config, tfpnModulesPath, baseModules, ... }:
{
  imports = baseModules ++ [
    "${tfpnModulesPath}/vultr-config.nix"
  ];
}

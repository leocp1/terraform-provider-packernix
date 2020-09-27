---
layout: "packernix"
page_title: "Packer Nix: `packernix_os`"
sidebar_current: "docs-packernix-datasource-os"
description: |-
  OS data source
---

# OS data source

Build a NixOS configuration.

## Example usage

```hcl
data "packernix_eval" "nixpkgs" {
  installable = "nixpkgs#path"
}

data "packernix_os" "example" {
  file = "./configuration.nix"
  config = jsonencode({
    "hostName" = "tfpnhost"
  })
  env = {
    "HOME" = "/homeless-shelter"
    "NIX_PATH" = "."
  }
  nixpkgs = jsondecode(data.packernix_eval.nixpkgs.out)
}
```

In `./configuration.nix`:

```nix
{ config, tfpnModulesPath, baseModules, ... }:
{
  imports = baseModules ++ [
    "${tfpnModulesPath}/amazon-ebs-config.nix"
  ];
  # Arbitrary configuration here
  networking.hostName = config.tfpn.hostName;
}
```

## Argument reference

The following arguments are supported: (Please see the general
[notes on paths](../index.html#notes-on-paths))

### Expression (exactly one of the following must be set)

- `file` - Path to Nix expression containing the main NixOS module.

- `installable` - A Nix flake style installable containing the main NixOS
  module. Note that this string will be passed verbatim to the generated
  `flake.nix` file, so consider also using the [`flake_path`](#flake_path)
  option if the flake is a local directory.

Unlike
[nixos-rebuild](https://nixos.org/manual/nixos/stable/index.html#sec-changing-config)
and other commands based off of the
[`eval-config.nix`](https://github.com/NixOS/nixpkgs/blob/master/nixos/lib/eval-config.nix)
function, the
[`baseModules`](https://github.com/NixOS/nixpkgs/blob/master/nixos/modules/module-list.nix)
will not be imported by default, and must be explicitly added to the
[imports list](https://nixos.org/manual/nixos/stable/index.html#module-syntax-2)
if desired.

The module will also be passed the `tfpnModulesPath` special argument, which
refers to a
[directory](https://github.com/leocp1/terraform-provider-packernix/tree/master/nixos/modules)
containing cloud provider specific configuration modules.

### Other options

- `nixpkgs` - (Required, unless the expression comes from a [`file`](#file) and
  `<nixpkgs>` follows the mentioned restrictions) A filesystem path to set
  [`<nixpkgs>`](https://nixos.org/manual/nix/stable/#env-NIX_PATH) to. For the
  OS data source specifically, the path must refer to a base Nixpkgs that
  contains the `nixos/modules` directory. If this is not the case, consider
  using the [Eval data source](./eval.html) to evaluate the `pkgs.path`
  attribute, then pass that value instead.

- `arg` - (Optional) A map of Nix expression valued arguments to pass to the Nix
  expression. Prefer `argstr` for passing HCL values to Nix. Note that at the
  time of writing (November 2020), this argument
  [is ignored](https://github.com/NixOS/nix/issues/3949) when using flakes.

- `argstr` - (Optional) A map of string valued arguments to pass to the Nix
  expression. To pass a HCL expression to Nix, encode the expression as JSON
  with
  [`jsonencode`](https://www.terraform.io/docs/configuration/functions/jsonencode.html)
  from the Terraform configuration, then decode the expression in Nix with
  [`builtins.fromJSON`](https://nixos.org/manual/nix/stable/#builtin-fromJSON).
  Note that at the time of writing (November 2020), this argument
  [is ignored](https://github.com/NixOS/nix/issues/3949) when using flakes.

- `build_path` - (Optional) A directory where the generated `default.nix`,
  `flake.nix`, and `tfpn-config.json` files will be written. If unset, a
  temporary directory will be created and deleted instead.

- `config` - (Optional) A JSON encoded object that will be available in the
  NixOS module under the `tfpn` module option.

- `clear_env` - (Optional) If this or the
  [provider `clear_env`](../index.html#clear_env) argument are set to true,
  start with an empty environment. Defaults to false.

- `env` - (Optional) A map of environment variables to set. Defaults to the
  empty map, but consider using settings similar to the
  [nixpkgs builder](https://github.com/nixos/ofborg#how-does-ofborg-call-nix-build):
  For example:

  ```hcl
  env = {
    "HOME" = "/homeless-shelter"
    "NIX_PATH" = "."
  }
  ```

- `flake_path` - (Optional) A filesystem path to a Nix flake that is prepended
  to `installable`. Respects the `working_dir`.

- `nix_options` - (Optional) A map of
  [Nix options](https://nixos.org/manual/nix/stable/#sec-conf-file) to set.

- `out_link` - (Optional) A filesystem path that will contain a symlink to the
  output path. If unset, no symlink will be created. Note that store paths
  without symlinks may be deleted by `nix-store --gc`.

- `working_dir` - (Optional) Working directory.

## Attributes reference

The following attributes are exported:

- `out_path` - The output Nix store path.

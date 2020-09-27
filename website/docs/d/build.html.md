---
layout: "packernix"
page_title: "Packer Nix: `packernix_build`"
sidebar_current: "docs-packernix-datasource-build"
description: |-
  Build data source
---

# Build data source

Build a Nix derivation.

## Example usage

```hcl
data "packernix_build" "example" {
  file = "default.nix"
  clear_env = true
  env = {
    "HOME" = "/homeless-shelter"
    "NIX_PATH" = "."
  }
}
```

## Argument reference

The following arguments are supported: (Please see the general
[notes on paths](../index.html#notes-on-paths))

### Expression (exactly one of the following must be set)

- `file` - Path to Nix expression containing derivation.

- `installable` - A Nix flake style installable.

### Other options

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

- `attr` - (Optional) Attribute to select from expression. Not valid if
  `installable` is set.

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

- `nix_options` - (Optional) A map of
  [Nix options](https://nixos.org/manual/nix/stable/#sec-conf-file) to set.

- `nixpkgs` - (Optional) A filesystem path to set
  [`<nixpkgs>`](https://nixos.org/manual/nix/stable/#env-NIX_PATH) to.

- `out_link` - (Optional) A filesystem path that will contain a symlink to the
  output path. If unset, no symlink will be created. Note that store paths
  without symlinks may be deleted by `nix-store --gc`.

- `working_dir` - (Optional) Working directory.

## Attributes reference

The following attributes are exported:

- `out_path` - The output Nix store path.

---
layout: "packernix"
page_title: "Packer Nix: `packernix_eval`"
sidebar_current: "docs-packernix-datasource-eval"
description: |-
  Eval data source
---

# Eval data source

Evaluate a Nix expression.

## Example usage

```hcl
data "packernix_eval" "example" {
  inline = <<EOT
builtins.fetchTarball {
  url = "https://github.com/NixOS/nixpkgs-channels/archive/nixos-20.03.tar.gz";
  sha256 = "the0hash0of0the0nixpkgs0channel000000000000000000000000000000000";
}
EOT
  nix_options = {
    "restrict-eval" = "true"
    "allowed-uris" = "https://github.com"
  }
  out_link = "nixpkgs-20.03"
}
```

## Argument reference

The following arguments are supported: (Please see the general
[notes on paths](../index.html#notes-on-paths))

### Expression (exactly one of the following must be set)

- `file` - Path to Nix expression.

- `inline` - String containing Nix expression.

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
  [nixpkgs builder](https://github.com/nixos/ofborg#how-does-ofborg-call-nix-build).
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

- `out_link` - (Optional) A filesystem path to use as a link name. If this
  option is set, `packernix_eval` will attempt to convert the evaluation output
  into a path, then create a Nix garbage collector root to it with
  `nix-store --realise`. If unset, no symlink will be created. Some example
  uses:

  - Workaround `builtins.fetchTarball`, which returns a Nix store path to the
    downloaded directory, but is not a derivation.
  - Pass a Nix store path [`inline`](#inline) as a string then download it from
    a [substituter](https://nixos.org/manual/nix/stable/#conf-substituters) if
    it is not already in the store.

- `working_dir` - (Optional) Working directory.

## Attributes reference

The following attributes are exported:

- `out` - The json encoded output of the Nix expression.

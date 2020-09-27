---
layout: "packernix"
page_title: "Provider: packernix"
sidebar_current: "docs-packernix-index"
description: |-
  The Packer/Nix provider is used to build Nix provisioned Packer images.
---

# Packer/Nix provider

A provider for deploying [NixOS](https://nixos.org/) configurations with
[Packer](https://www.packer.io/). A good entry point is the
[image](./r/image.html) resource.

## Argument reference

The arguments described below are intended as default values for the
corresponding arguments in the resources. In general, they are not saved to the
state.

### General

- `working_dir` - (Optional) The default working directory for resources.
  (Please see the [notes on `path`](#provider-working_dir))

- `env` - (Optional) Map of environment variables to set for all resources.
  Defaults to the empty map. Note:

  - This map is not saved to the state, so consider setting variables for Packer
    authentication here.
  - The values are passed unmodified, so all relative paths in the variables are
    resolved relative to resource `working_dir`, not provider `working_dir`.
  - If a variable is set both here and in a resource `env` argument, the
    variable will take the resource `env` argument's value.

- `clear_env` - (Optional) If set to true, force all resources to start with an
  empty environment. Defaults to false.

### Nix

- `flake` - (Optional) A Nix flake that is prepended to `installable`s by
  default. Conflicts with `flake_path`.

- `flake_path` - (Optional) A filesystem path to a Nix flake that is prepended
  to `installable`s by default. Conflicts with `flake`. Respects the
  `working_dir`.

- `nix_options` - (Optional) A map of
  [Nix options](https://nixos.org/manual/nix/stable/#sec-conf-file) to set for
  all commands by default. If an option is set in both the resource and the
  provider, the resource's value takes precedence. Defaults to

  ```hcl
  {
    "restrict_eval" = "true"
  }
  ```

- `nixpkgs` - (Optional) A filesystem path to set
  [`<nixpkgs>`](https://nixos.org/manual/nix/stable/#env-NIX_PATH) to by
  default.

## Notes on paths

### Resource `working_dir`

All the resources in this provider that accept a file path argument also accept
an optional `working_dir` argument. This argument controls

- The root of relative paths passed into the resource
- The working directory of any external commands called

If the resource `working_dir` is a relative path, its root is the
[provider `working_dir`](#provider-working_dir)

### Provider `working_dir`

The provider also has a required `working_dir` argument. This argument controls

- The root of relative paths passed into the provider configuration
- The default value for resources whose `working_dir` is unset
- The root of relative resource `working_dir`s

It defaults to the current working directory.

If the provider `working_dir` is a relative path, its root is the current
working directory.

### Internal attributes

Every resource has a `path_uses_provider_wd` attribute that stores which
`working_dir` should be used to resolve relative paths.

Some resources have a `path_hashes` attribute that stores cryptographic hashes
of paths referred to in the state.

Both are intended as internal implementation details.

### Spurious diffs and normalization

Referring directly to filesystem paths in resource arguments may cause spurious
diffs if the same configuration is applied from multiple systems or on different
host operating systems.

To attempt to reduce these spurious diffs:

- Paths and `working_dir` have their slash (`/`) characters become the OS
  specific separator when being read from the configuration or the state
- Absolute paths are cleaned when being written to the state
- Relative paths have OS specific separators become the slash (`/`) character
  when being written to the state
- The provider's `working_dir` is not saved to the state and assumed to be
  system/OS specific
- A resource `working_dir` argument is only saved to the state if set explicitly
  in the configuration

### Recommended practices

- The `working_dir` provider argument should be set to `dirname(path.module)`
- Path arguments should be
  - A Nix store path
  - A relative path
- A `working_dir` resource argument should be either
  - Unset
  - Set to a relative path
  - Set to a Nix store path

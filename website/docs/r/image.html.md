---
layout: "packernix"
page_title: "Packer Nix: `packernix_image`"
sidebar_current: "docs-packernix-resource-image"
description: |-
  Image resource
---

# Image resource

Build a machine image from a Packer template.

## Dependencies

This resource depends on having a `packer` executable in the path with access to
any plugins referred to in the passed [template](#template), including the
special `delete-$builderName` builder, that reads and deletes images built by a
`$builderName` builder plugin.

## Warning

This resource assumes all Packer templates with the same
[`builders`](https://www.packer.io/docs/templates/builders.html) section are
equivalent. Thus:

- The `builders` section of the Packer template must indicate what the template
  provisions on the image. With the Packer templates generated from the
  [bundled Nix expressions](#using-the-bundled-templates), the condition is met
  by setting the "resource name" or "description" of the image to the hash of
  the provisioned NixOS configuration.
- When destroying the resource, all images that could have been created by a
  Packer template with the same `builders` section will be deleted, not just the
  image in the state.
- When creating the resource, if a preexisting image is found that matches the
  Packer `builders` section, it will be reused, and no new image will be
  created.

## Example usage

```hcl
provider packernix {
  working_dir = dirname(path.module)
}

# Download nixpkgs
data "packernix_eval" "nixpkgs" {
  installable = "nixpkgs#path"
  out_link = "build/nixpkgs"
}

# Build an OS configuration
data "packernix_os" "nixos" {
  file = "./configuration.nix"
  env = {
    "HOME" = "/homeless-shelter"
    "NIX_PATH" = "."
  }
  nixpkgs = jsondecode(data.packernix_eval.nixpkgs.out)
  build_path = "build"
  out_link = "build/result-nixos"
}

# Obtain the Packer template generator directory
data "packernix_const" "out" {}

# Generate a Packer template matching the OS configuration
data "packernix_eval" "template" {
  file = "${data.packernix_const.out.packer}/amazon-ebs.nix"
  env = {
    "HOME" = "/homeless-shelter"
    "NIX_PATH" = data.packernix_const.out.packer
  }
  argstr = {
    nixos = data.packernix_os.nixos.out_path
  }
}

# Build a machine image with the Packer template
resource "packernix_image" "image" {
  template = data.packernix_eval.template.out
  build_path = "build"
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
}
```

## Argument reference

The following arguments are supported: (Please see the general
[notes on paths](../index.html#notes-on-paths))

- `template` - (Required) A Packer template passed as a JSON string. This is
  required to be a valid Packer template with only one builder.

- `build_path` - (Optional) A directory where the generated `create.json`,
  `read.json`, and `delete.json` files will be written. If unset, a temporary
  directory will be created and deleted instead. Using the same `build_path` for
  two different `packernix_image` resources is not allowed, since both resources
  will try to write to the same `*.json` paths. Similarly, deposed objects will
  try to write to the same `delete.json` path as the current resource.

- `clear_env` - (Optional) If this or the
  [provider `clear_env`](../index.html#clear_env) argument are set to true,
  start with an empty environment. Defaults to false.

- `env` - (Optional) A map of environment variables to set. Defaults to the
  empty map.

- `working_dir` - (Optional) Working directory.

## Attributes reference

The following attributes are exported:

- `builder_id` - The ID of the builder.
- `image` - The output machine image ID.

## Using the bundled templates

The
[Packer template generator](https://github.com/leocp1/terraform-provider-packernix/tree/master/packer)
directory (exported as `data.packernix_const.out.packer` in the example above)
contains Nix functions that evaluate to Packer templates for a chosen builder.
Each template generator function is allowed to define builder specific input
arguments, but some common ones include:

- `nixos` : (Required) A string containing the Nix store path of the NixOS
  configuration to copy.
- `builder` : (Optional) An attribute set of values to override in the final
  [builder](https://www.packer.io/docs/templates/builders). If a string is
  passed, it will be automatically decoded as JSON.
- `variables` : (Optional) An attribute set of values to override in the final
  [variables](https://www.packer.io/docs/templates/user-variables) section. If a
  string is passed, it will be automatically decoded as JSON.
- `nix_conf` : (Optional) A string containing the contents of a
  [`nix.conf`](https://nixos.org/manual/nix/stable/#sec-conf-file) to write to
  the remote before `nix-copy-closure`. This is intended to allow control of
  [substituters](https://nixos.org/manual/nix/stable/#conf-substituters).
- `fakessh` : (Optional) If set to `true`, use the
  [`fakessh`](https://github.com/leocp1/packer-provisioner-fakessh) provisioner
  instead of system `ssh`.
- `compress` : (Optional) If set to `false`, do not compress data sent over
  `ssh`. `fakessh` does not support compression.

The bundled template generators import from paths in the `packer` directory, so
the `packer` directory should be added to the `NIX_PATH`.

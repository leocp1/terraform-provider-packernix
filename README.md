# Terraform Provider for Nix with Packer

A [Terraform](https://www.terraform.io) provider for deploying
[NixOS](https://nixos.org/) configurations with
[Packer](https://www.packer.io/).

## Approach

- Build a NixOS configuration with a Terraform wrapped version of `nix-build`.
- Build a machine image from this configuration with a Terraform wrapped version
  of Packer.
- Deploy this image with Terraform.

Compared to [other approaches](Alternatives.md) this has a few benefits:

- If the NixOS or Packer configuration changes, `terraform apply` will detect it
  and redeploy.
- Machine images make redeploying instances of the same configuration fast.
- Adding support for a [new cloud provider](NewProviders.md) only requires
  writing a Packer builder plugin, a NixOS module, and a Packer template.

## Building

A `nixpkgs` style derivation is in `default.nix`. The derivation has an unlisted
dependency on having a `packer` in your `PATH` with access to the
[packer-provisioner-fakessh](https://github.com/leocp1/packer-provisioner-fakessh)
plugin and any other provider specific plugins used.

To try out the plugin with AWS, run

```sh
nix-shell shell.nix
```

## Documentation

The [website](./website) directory contains
[terraform-website](https://github.com/hashicorp/terraform-website) compatible
markdown files.

## Cloud Support

| Provider | Packer template | `packer-builder-delete-` | NixOS module |
| ........ | :.............: | :......................: | :..........: |
| `vultr` | [✔️](./packer/vultr.nix) | [✔️](https://github.com/leocp1/packer-builder-delete-vultr) | [✔️](./nixos/modules/vultr-config.nix) |

## Testing

The provider uses the usual `go test` command for testing:

- Creating NixOS images can be slow, so consider disabling the timeout with
  [`-timeout 0`](https://golang.org/cmd/go/#hdr-Testing_flags).
- Most of the tests require
  [`TF_ACC=true`](https://www.terraform.io/docs/extend/testing/acceptance-tests/index.html#running-acceptance-tests)
  to run.
- Set [`TF_LOG=INFO`](https://www.terraform.io/docs/internals/debugging.html) or
  higher to view command output.

The location of the share directory can be overridden with the
`TERRAFORM_PACKERNIX_SHARE` environment variable. This directory is expected to
contain the following directories from the repository root.

- `nixos/modules`: Cloud provider specific configuration.
- `nixos/template`: Template files for the OS data source.
- `packer`: Packer template generator Nix functions.

The `packer` directory is also expected to be added to the Nix store with name
`packer`, so running

```bash
nix-store --recursive --add-fixed sha256 "$TERRAFORM_PACKERNIX_SHARE/packer"
```

may be necessary.

## Licenses

All Go source files (files with extension `.go`) in this repository are licensed
under the Mozilla Public License 2.0.
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

All Nix source files (files with extension `.nix`) in this repository, all shell
scripts (files with extension `.sh`) in this repository, and all files under the
`pkgs` directory are licensed under the MIT License.
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

All test data files (files under a `testdata` directory) in this repository not
covered by the previous two paragraphs and the provider documentation (files
under the root `website` directory and files with extension `.md` in the top
level directory) are licensed under the CC0 License 1.0 Universal.
[![License: CC0-1.0](https://img.shields.io/badge/License-CC0%201.0-lightgrey.svg)](http://creativecommons.org/publicdomain/zero/1.0/)

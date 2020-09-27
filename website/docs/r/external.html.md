---
layout: "packernix"
page_title: "Packer Nix: `packernix_external`"
sidebar_current: "docs-packernix-resource-external"
description: |-
  External resource
---

# External resource

When using the
[`terraform-plugin-sdk`](https://github.com/hashicorp/terraform-plugin-sdk), a
Terraform resource is usually implemented by defining a set of CRUD operations.

The `packernix_external` resource allows the user to manage a custom resource by
using a set of executables as the implementation of the create, read, and delete
operations.

~> **Warning** This mechanism is provided as an "escape hatch" for exceptional
situations where a first-class Terraform provider is not more appropriate. If
possible, consider using the [external data source](../d/external.html) to avoid
the complexities of managing state with external programs.

## Example usage: Install a NixOS configuration to the local machine

```hcl
# Download nixpkgs
data "packernix_eval" "nixpkgs" {
  installable = "nixpkgs#path"
  out_link = "result-nixpkgs"
}

# Build an OS configuration
data "packernix_os" "nixos" {
  file = "./configuration.nix"
  env = {
    "HOME" = "/homeless-shelter"
    "NIX_PATH" = "."
  }
  nixpkgs = jsondecode(data.packernix_eval.nixpkgs.out)
  out_link = "result-nixos"
}

# Build askpass
data "packernix_build" "ap" {
  installable = "nixpkgs#x11_ssh_askpass"
  out_link = "result-askpass"
}

# Install the NixOS configuration
resource "packernix_external" "example" {
  # Ideally this path would be a Nix store path or some other immutable
  # directory, but an ordinary directory is used here for demonstration purposes
  path = "./nixos_install"
  env = {
    "SUDO_ASKPASS" = "${data.packernix_build.ap.out_path}/libexec/ssh-askpass"
  }
  options = [
    data.packernix_os.nixos.out_path
  ]
}
```

In `./nixos_install/bin/create`

```sh
#!/usr/bin/env bash

# Install NixOS to the local machine

set -euo pipefail
cfg="$1"
# Use askpass so password entry does not happen over stdin
sudo -A "$cfg/bin/nix-env" \
  --profile /nix/var/nix/profiles/system \
  --set "$cfg"
sudo -A "$cfg/bin/switch-to-configuration" switch
printf "$cfg"
```

In `./nixos_install/bin/read`

```sh
#!/usr/bin/env bash

# Read the current system.

set -euo pipefail
desired="$1"
prev=$(tr -d '[:space:]' </dev/stdin)
current="$(realpath /nix/var/nix/profiles/system | tr -d '[:space:]')"

if [ "$prev" == ""  ]; then
  # Find a state value
  if [ "$#" -eq 0 ] || [ "$desired" == "$current" ]; then
    printf "$current"
  fi
else
  # Refresh the current value of the state
  printf "$prev"
fi
```

In `./nixos_install/bin/delete`

```sh
#!/usr/bin/env bash

# "Delete" the current system.
# Should technically uninstall NixOS, but since that may leave the system
# inoperable, this script is a noop.
```

## Argument reference

The following arguments are supported: (Please see the general
[notes on paths](../index.html#notes-on-paths))

- `path` - (Required) A path containing executables at:

  - `$path/bin/read`: implementation of "refresh" and "find" operations
  - `$path/bin/create`: implementation of "create" operation
  - `$path/bin/delete`: implementation of "delete" operation

- `options` - (Optional) A list of command line options to pass to the programs.
  Defaults to the empty list.

- `clear_env` - (Optional) If this or the
  [provider `clear_env`](../index.html#clear_env) argument are set to true,
  start with an empty environment. Defaults to false.

- `env` - (Optional) A map of environment variables to set. Defaults to the
  empty map.

- `working_dir` - (Optional) Working directory of the programs.

## Attributes reference

The following attributes are exported:

- `state` - A string uniquely identifying the resource. The output of `read` or
  `create`.

## External program protocol

A Terraform-managed resource is stored in the state file as a string that
uniquely identifies it (the resource's ID). The external set of programs
communicate with Terraform by passing this ID over standard input/output:

- The `create` program must create a resource then print its ID to standard
  output.
- The `delete` program will be passed an ID over standard input. It must then
  delete the associated resource.
- The `read` program implements two operations:
  - When passed no data over standard input, the `read` program must attempt to
    "find" a resource consistent with the passed command-line options and
    environment variables. If at least one such resource is found, the program
    must pick one and print its ID to standard output. Otherwise, the program
    must print the empty string to standard output.
  - When passed an ID over standard input, the `read` program must "refresh" the
    current state of the corresponding resource and output its ID to standard
    output. The input ID and the output ID are allowed to differ, as long as
    `delete`-ing the output ID is the same as `delete`-ing the input ID.
- Update operations are not supported. When the configuration changes, Terraform
  will call the "find" operation on the configuration. If the returned ID is the
  same as the ID stored in the state, the configuration is considered up to
  date, and no further operations happen. Otherwise, the resource recorded in
  the state is destroyed and a new resource is created from the configuration.

If any program encounters an error, it must exit with a non-zero status. Any
data on standard output is ignored in this case.

All programs will receive the same [options](#options) and [environment](#env).

Having two resources with the same arguments is not allowed: the arguments
should uniquely identify a resource.

## Maintaining access to `path` between runs

If a [path](#path) is set in the Terraform state, it must exist on every machine
running Terraform and must not be modified until it is unset, or refreshes and
deletes will fail.

For example, if using a Nix store path for this argument, the path should be:

- put in a store available to all machines running Terraform
- registered as a store root so it is not garbage collected between runs.

## State string representation

When encoding a more complex state, consider outputting JSON and using the
[`jsondecode`](https://www.terraform.io/docs/configuration/functions/jsondecode.html)
function to access the state within Terraform.

## Import

The external resources can be imported:

```sh
terraform import packernix_external.example uuid
```

This is an ID-only import: no validation takes place. Make sure the identifier
matches the expected format of the programs set in the state.

## See also

- [External Data Source](https://registry.terraform.io/providers/hashicorp/external/latest/docs/data-sources/data_source)

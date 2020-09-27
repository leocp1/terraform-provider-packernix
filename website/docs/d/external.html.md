---
layout: "packernix"
page_title: "Packer Nix: `packernix_external`"
sidebar_current: "docs-packernix-datasource-external"
description: |-
  External data source
---

# External data source

Allow an external program to act as a data source.

~> **Warning** This mechanism is provided as an "escape hatch" for exceptional
situations where a first-class Terraform provider is not more appropriate.

## Example usage

```hcl
data "packernix_external" "example" {
  path = "/nix/store/0external0data0source0hash0here0-data-source-name"
}
```

## Argument reference

The following arguments are supported: (Please see the general
[notes on paths](../index.html#notes-on-paths))

- `path` - (Required) A path containing an executable at `$path/bin/read` to
  run.

- `options` - (Optional) A list of command line options to pass to the `read`
  program. Defaults to the empty list.

- `clear_env` - (Optional) If this or the
  [provider `clear_env`](../index.html#clear_env) argument are set to true,
  start with an empty environment. Defaults to false.

- `env` - (Optional) A map of environment variables to set. Defaults to the
  empty map.

- `working_dir` - (Optional) Working directory of the program.

## Attributes reference

The following attributes are exported:

- `state` - A string uniquely identifying the data source. The output of `read`.

## External program protocol

The `read` program must print a string uniquely identifying the data source to
standard output or the empty string if a data source can not be found.

If the program encounters an error, it must exit with a non-zero status. Any
data on standard output is ignored in this case.

Terraform expects a data source to have _no observable side-effects_, and will
re-run the program each time the state is refreshed.

## State string representation

When encoding a more complex state, consider outputting JSON and using the
[`jsondecode`](https://www.terraform.io/docs/configuration/functions/jsondecode.html)
function to access the state within Terraform.

## See also

- [External Data Source](https://registry.terraform.io/providers/hashicorp/external/latest/docs/data-sources/data_source)

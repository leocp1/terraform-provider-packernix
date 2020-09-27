---
layout: "packernix"
page_title: "Packer Nix: `packernix_const`"
sidebar_current: "docs-packernix-datasource-const"
description: |-
  Const data source
---

# Const data source

Read provider constants. This data source takes no arguments.

## Example usage

```hcl
data "packernix_const" "out" {}
```

## Attributes reference

The following attributes are exported:

- `modules` - The file path used as the
  [NixOS module](https://github.com/leocp1/terraform-provider-packernix/tree/master/nixos/modules)
  directory.

- `nix` - The Nix executable.

- `packer` - The file path used as the
  [Packer template generator](https://github.com/leocp1/terraform-provider-packernix/tree/master/packer)
  directory.

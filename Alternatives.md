# Alternatives

## [krops](https://github.com/krebs/krops) ([plops](https://github.com/mrVanDalo/plops))

Nix library that creates an executable that

- accesses a remote NixOS server via ssh
- runs a variant of nixos-rebuild on the server

## [nix-deploy](https://github.com/awakesecurity/nix-deploy)

- Provision NixOS machines with any provisioner
- Run reimplementation of `nixos-rebuild` that accepts a nix store path instead
  of a `configuration.nix`

See
[nix-deploy](https://awakesecurity.com/blog/deploy-software-easily-securely-using-nix-deploy/)
and
[NixOS in Production](http://www.haskellforall.com/2018/08/nixos-in-production.html)
for more details.

## [nixform](https://github.com/brainrape/nixform)

Evaluate a nix attribute set strictly to convert it into a Terraform JSON
configuration.

## [nixiform](https://github.com/icetan/nixiform)

- Provision Ubuntu nodes with some provisioner
- Pass a name, IP, ssh key, and provider from the provisioner
- Install NixOS by ssh-ing into node and manually editing paths in bash similar
  to [`nixos-infect`](https://github.com/elitak/nixos-infect)
- Add
  [provider specific settings](https://github.com/icetan/nixiform/tree/master/lib/configurators)
  to NixOS configuration and build
- `nix-copy-closure` the NixOS closure
- Run NixOS switch

## [nixops](https://nixos.org/nixops/)

- Parse Nix file to produce `nixops` deploy/destroy commands that will produce a
  NixOS install on a new node (the base NixOS machine image is usually prebuilt)
- Add provider specific hardware configuration to NixOS configuration and build
- Run `nix-copy-closure` the NixOS closure
- Run NixOS switch

## [nixops-terraform](https://github.com/adisbladis/nixops-terraform)

NixOps Terraform plugin.

## [nixos-rebuild](https://www.mankier.com/8/nixos-rebuild)

- built-in way to generate and swap to a configuration.nix:

```
  nixos-rebuild switch \
    --target-host $target \
    --build-host $target \
    -I nixos-config=$NIXOS_CONFIG
```

## [nixus](https://github.com/Infinisil/nixus)

Uses Nix modules to generate a bash script that

- builds a NixOS configuration
- copies the configuration with `nix-copy-closure` to a remote host
- calls `$nixos_store_path/bin/switch-to-configuration`

## [terraform-nixos](https://github.com/tweag/terraform-nixos)

- (AWS) Map version, region, and AMI type to an official NixOS AMI.
- (Google) Map version to official NixOS tarball, then use tarball to create
  NixOS image.
- (Google) Create a Google Compute Image with
  `<nixpkgs/nixos/modules/virtualisation/google-compute-image.nix>`. Upload
  using the Google Terraform provider.
- Update running NixOS machine with Terraform wrapped equivalent to
  `nixos-rebuild`

## [terraform-provider-nix](https://github.com/andrewchambers/terraform-provider-nix)

- Build ad hoc Nix expression and return store paths to Terraform (notably, the
  `${providerName}-image.nix` expressions in
  [`<nixpkgs/nixos/modules/virtualisation/>`](https://github.com/NixOS/nixpkgs/tree/master/nixos/modules/virtualisation)
  produce machine images for many providers)
- Update running NixOS machine with Terraform wrapped `nixos-rebuild`

## [terraform-provider-nixos](https://github.com/tweag/terraform-provider-nixos)

- Provision NixOS servers with Terraform, adding the `nixos_node` resource.
- Provider outputs extra configuration nix files to be consumed by `nixops`
- Run `nixops` to configure the provisioned NixOS servers

## [terraform-provider-packer (gsaslis)](https://github.com/gsaslis/terraform-provider-packer)

WIP?

## [terraform-provider-packer](https://github.com/juliosueiras/terraform-provider-packer)

A Terraform Provider to generate Packer JSON

## [terranix](https://github.com/mrVanDalo/terranix) ([jack-williamson](https://github.com/kreisys/jack-williamson)?)

Generate terraform.json from Nix.

## [tfnix](https://github.com/arcnmx/tf-nix)

Use a Nix expression to generate Terraform JSON that calls
`"${config.system}/bin/switch-to-configuration switch"` in the
[provisioner](https://www.terraform.io/docs/provisioners/index.html) block.

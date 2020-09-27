{
  outputs = { self, ... }: {
    nixosConfiguration = import ./configuration.nix;
  };
}

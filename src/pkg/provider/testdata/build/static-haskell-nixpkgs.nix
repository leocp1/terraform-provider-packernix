# pinned nixpkgs version used by https://github.com/nh2/static-haskell-nix
builtins.fetchTarball {
  url = "https://github.com/NixOS/nixpkgs/archive/0c960262d159d3a884dadc3d4e4b131557dad116.tar.gz";
  sha256 = "0d7ms4dxbxvd6f8zrgymr6njvka54fppph1mrjjlcan7y0dhi5rb";
}

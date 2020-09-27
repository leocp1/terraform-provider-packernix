rec {
  inherit (import ./nixpkgslib.nix) filterAttrs recursiveUpdate;

  /*
    A store path containing the Packer template generator directory. Prefer this
    to ./.  when specifying scripts since `nix-instaniate --eval` will convert
    paths to hashed Nix store paths, but not add them to the store.

    Type: path
  */
  path = builtins.path {
    path = ./.;
    name = "packer";
  };

  /*
    If the argument is a string, assume it is JSON and decode. Otherwise just
    return the argument.
  */
  fromMaybeJSON = arg:
    if builtins.isString arg
    then builtins.fromJSON arg
    else arg;

  /*
     Produce an attribute set representing a Packer template. Strictly evaluates
     since `nix-instantiate --eval` will likely not fully force the thunks.

     Adds nixos (the passed store path) and name (a cleaned resource name) to
     the variables

     Type: mkTemplate :: attrs -> attrs
  */
  mkTemplate =
    {
      # The store path to the NixOS config to provision
      nixos
      # The initial builder configuration
    , builder' ? {}
      # possibly JSON encoded builder overrides specified by the user
    , builder ? {}
      # The initial variables configuration
    , variables' ? {}
      # possibly JSON encoded variables overrides specified by the user
    , variables ? {}
    , ...
    }@args:
      let
        capturedArgs = {
          nixos = false;
          builder' = false;
          builder = false;
          variables' = false;
          variables = false;
        };
        isUncaptured = n: v: capturedArgs."${n}" or true;
        uncaptured = filterAttrs isUncaptured args;
        b = fromMaybeJSON builder;
        vs = fromMaybeJSON variables;
        bf = recursiveUpdate builder' b;
        vf = recursiveUpdate
          (
            recursiveUpdate
              {
                inherit nixos;
                name = cleanResourceName nixos;
              }
              variables'
          )
          vs;
        t = uncaptured // {
          builders = [ bf ];
          variables = vf;
        };
      in
        assert builtins.isAttrs b;
        assert builtins.isAttrs builder';
        assert builtins.isAttrs vs;
        assert builtins.isAttrs variables';
        builtins.deepSeq t t;

  /*
    Get Nix store path hash

    Type: pathHash :: string -> string
  */
  pathHash = p: let
    bn = builtins.baseNameOf p;
  in
    builtins.substring 0 32 bn;

  /*
     Convert a string containing a NixOS configuration store path to a clean
     resource name.

     Type: cleanResourceName :: string -> string
  */
  cleanResourceName = p: "nixos-${pathHash p}";

  /*
    Create a provisioner step to write to write nix.conf

    Type: writeNixConf :: string -> provisioner
  */
  writeNixConf = nc:
    let
      qnc = builtins.replaceStrings [ "'" ] [ "' \\' '" ] nc;
      dstDir = ''
        ''${XDG_CONFIG_HOME:-$HOME/.config}/nix'';
    in
      {
        type = "shell";
        inline = [
          "mkdir -p \"${dstDir}\""
          "printf -- '%s' '${qnc}' > \"${dstDir}/nix.conf\""
        ];
      };

  /*
    nix-copy-closure wrapper.

    Compression not supported with fakessh.

    Type: nixCopyClosure :: atrs -> provisioner
  */
  nixCopyClosure =
    {
      # Derivation to copy
      drv ? "{{ user `nixos` }}"
      # Whether to use packer-provisioner-fakessh or system ssh
    , fakessh ? false
      # Whether to compress data
    , compress ? true
    }: let
      remote =
        if fakessh
        then "packer-communicator-ssh"
        else "\"ssh://{{build `User`}}@{{build `Host`}}:{{build `Port`}}\"";
      # Currently Go ssh library does not support gzip compression
      fakestep = {
        type = "fakessh";
        inline = [ "nix-copy-closure --to -s ${remote} ${drv}" ];
      };
      tmpid = "/tmp/packer-session-{{build `PackerRunUUID`}}.pem";
      sshopts = toString (
        [
          "-i ${tmpid}"
          "-F /dev/null"
          "-o StrictHostKeyChecking=no"
          "-o CheckHostIP=no"
          "-o UserKnownHostsFile=/dev/null"
        ]
        ++ (if compress then [ "-C" ] else [])
      );
      step = {
        type = "shell-local";
        inline = [
          "umask 077"
          "echo '{{ build `SSHPrivateKey`}}' > ${tmpid}"
          "nix-copy-closure --to -s ${remote} ${drv}"
        ];
        environment_vars = [
          "NIX_SSHOPTS=${sshopts}"
        ];
      };
    in
      assert compress && fakessh == false;
      if fakessh then fakestep else step;
}

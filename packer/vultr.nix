{ lib ? (import ./lib.nix)
, nixos
, builder ? {}
, variables ? {}
, nix_conf ? ""
, fakessh ? false
, compress ? true
}:
lib.mkTemplate {
  inherit nixos builder variables;
  variables' = {
    vultr_api_key = "{{ env `VULTR_API_KEY` }}";
  };
  builder' = {
    type = "vultr";
    api_key = "{{ user `vultr_api_key` }}";
    os_id = 352; # Debian
    region_id = 1; # NJ
    plan_id = 201; # Least expensive for now
    ssh_username = "root";
    snapshot_description = "{{ user `name`}}";
    instance_label = "packer-builder";
    state_timeout = "24h";
  };
  provisioners = [
    {
      type = "shell";
      script = "${lib.path}/scripts/nix_install_root.sh";
    }
    (lib.writeNixConf nix_conf)
    {
      type = "shell";
      script = "${lib.path}/scripts/auto_relabel.sh";
      environment_vars = [
        "ROOT=/"
        "INPUT_LABEL=nixos"
      ];
    }
    (
      lib.nixCopyClosure {
        inherit fakessh compress;
      }
    )
    {
      type = "shell";
      script = "${lib.path}/scripts/nixos_infect.sh";
      environment_vars = [
        "NIXOS_STORE_PATH={{ user `nixos` }}"
        "NIXOS_ROOT=/"
      ];
    }
    {
      type = "shell";
      inline = "reboot";
      pause_after = "10s";
      expect_disconnect = true;
    }
    {
      type = "shell";
      script = "${lib.path}/scripts/cleanup.sh";
    }
  ];
}

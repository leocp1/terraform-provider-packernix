#! /usr/bin/env sh

#
# nixos_infect - install a NixOS config on a Linux system with Nix installed.
#
# Input variables:
# NIXOS_STORE_PATH - the store path to the system configuration
# NIXOS_ROOT - the path to install NixOS to
#
NIXOS_ROOT="${NIXOS_ROOT:-/}"
if [ -z ${NIXOS_STORE_PATH+x} ]; then
  >&2 echo "NIXOS_STORE_PATH must be set"
fi

# https://github.com/NixOS/nixpkgs/issues/38991
export LOCALE_ARCHIVE="$NIXOS_STORE_PATH/sw/lib/locale/locale-archive"
export NIXOS_INSTALL_BOOTLOADER=1

chown -R 0.0 /nix

if ! [ -e "$NIXOS_ROOT/etc/NIXOS" ] ; then
  touch "$NIXOS_ROOT/etc/NIXOS"
  touch "$NIXOS_ROOT/etc/NIXOS_LUSTRATE"
  echo "$HOME/.ssh/authorized_keys" >> "$NIXOS_ROOT/etc/NIXOS_LUSTRATE"
  rm --one-file-system -rf "$NIXOS_ROOT/boot"
fi

if [ "$NIXOS_ROOT" = "/" ] ; then
  "$NIXOS_STORE_PATH/sw/bin/nix-env" --profile /nix/var/nix/profiles/system \
    --set "$NIXOS_STORE_PATH"
  "$NIXOS_STORE_PATH/bin/switch-to-configuration" boot
else
  mkdir -m 0755 -p "$NIXOS_ROOT/etc"
  ln -sfn /proc/mounts "$NIXOS_ROOT/etc/mtab"
  cp -a --parents /etc/passwd "$NIXOS_ROOT"
  cp -a --parents /etc/group "$NIXOS_ROOT"
  cp -a --parents "$HOME/.ssh/authorized_keys" "$NIXOS_ROOT"
  "$NIXOS_STORE_PATH/sw/bin/nixos-enter" --root "$NIXOS_ROOT" -- \
    "$NIXOS_STORE_PATH/sw/bin/nix-env" --profile /nix/var/nix/profiles/system \
    --set "$NIXOS_STORE_PATH"
  "$NIXOS_STORE_PATH/sw/bin/nixos-enter" --root "$NIXOS_ROOT" -- \
     "$NIXOS_STORE_PATH/bin/switch-to-configuration" boot
fi

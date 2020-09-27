#!/usr/bin/env sh

#
# auto_relabel: relabel a disk
#
# Input variables:
# ROOT : mount point of disk we want to relabel
# INPUT_LABEL : new label name
#
ROOT="${ROOT:-/}"
INPUT_LABEL="${INPUT_LABEL:-nixos}"

ROOT_DEVICE=$(< /proc/mounts grep " $ROOT " | cut -f 1 -d ' ')
ROOT_FSTYPE=$(< /proc/mounts grep " $ROOT " | cut -f 3 -d ' ')
if [ -e "/dev/disk/by-label/$INPUT_LABEL" ] ; then
  if ! [[ $(readlink -f "$ROOT_DEVICE") = $(readlink -f "/dev/disk/by-label/$INPUT_LABEL") ]] ; then
    exit 1
  fi
else
  case "$ROOT_FSTYPE" in
    ext2)
      e2label "$ROOT_DEVICE" "$INPUT_LABEL"
      ;;
    ext3)
      e2label "$ROOT_DEVICE" "$INPUT_LABEL"
      ;;
    ext4)
      e2label "$ROOT_DEVICE" "$INPUT_LABEL"
      ;;
    btrfs)
      btrfs filesystem label "/" "$INPUT_LABEL"
      ;;
    *)
      exit 1
      ;;
  esac
fi

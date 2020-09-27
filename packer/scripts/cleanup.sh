#! /usr/bin/env sh

#
# cleanup: clean files leftover after nixos-infect
#

if [ -e /old-root ] ; then
  rm -rf /old-root
fi
nix-collect-garbage -d

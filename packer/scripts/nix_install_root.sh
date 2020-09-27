#! /usr/bin/env sh

#
# nixos_install_root : Install Nix as root
#

# Create non-root user with sudo
echo 'packer  ALL=(ALL) NOPASSWD: ALL' > /etc/sudoers.d/packer
chmod 0440 /etc/sudoers.d/packer
useradd -m packer

# Single user nix install
curl -L "https://nixos.org/nix/install" | su packer -c sh

# Horrible hack: just copy nix dotfiles to root
echo '. /home/packer/.nix-profile/etc/profile.d/nix.sh' >> $HOME/.profile
echo '. /home/packer/.nix-profile/etc/profile.d/nix.sh' >> $HOME/.bashrc
cp -r /home/packer/.nix-profile $HOME
cp -r /home/packer/.nix-defexpr $HOME

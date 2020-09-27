// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Possibly patched to include Nix store paths.
package patches

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/yookoala/realpath"
)

func Nix() string {
	if "@nix@" == ("@" + "nix@") {
		return "nix"
	} else {
		return filepath.Join("@nix@", "bin", "nix")
	}
}

func NixHash() string {
	if "@nix@" == ("@" + "nix@") {
		return "nix-hash"
	} else {
		return filepath.Join("@nix@", "bin", "nix-hash")
	}
}

func NixBuild() string {
	if "@nix@" == ("@" + "nix@") {
		return "nix-build"
	} else {
		return filepath.Join("@nix@", "bin", "nix-build")
	}
}

func NixInstantiate() string {
	if "@nix@" == ("@" + "nix@") {
		return "nix-instantiate"
	} else {
		return filepath.Join("@nix@", "bin", "nix-instantiate")
	}
}

func NixStore() string {
	if "@nix@" == ("@" + "nix@") {
		return "nix-store"
	} else {
		return filepath.Join("@nix@", "bin", "nix-store")
	}
}

func Packer() string {
	if "@packer@" == ("@" + "packer@") {
		return "packer"
	} else {
		return filepath.Join("@packer@", "bin", "packer")
	}
}

func Share() (share string) {
	var err error
	share = os.Getenv("TERRAFORM_PACKERNIX_SHARE")
	if share != "" {
		share, err = realpath.Realpath(share)
		if err == nil {
			return
		}
	}

	if "@out@" != ("@" + "out@") {
		share = filepath.Join("@out@", "share")
		return
	}

	wd, err := os.Getwd()
	if err == nil {
		share = filepath.Join(wd, "..", "..", "..")
		return
	}

	return "."
}

// Check for nix flake support
func SupportsNixFlake(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, Nix(), "flake", "--help")
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
}

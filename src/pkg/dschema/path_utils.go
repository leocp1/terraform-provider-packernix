// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package dschema

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/leocp1/terraform-provider-packernix/src/pkg/logwriter"
	"github.com/leocp1/terraform-provider-packernix/src/pkg/patches"
)

// Prepend the working dir to a relative path
// and check if path is valid
func readRelPath(ps string, wds string) (p string) {
	op := filepath.FromSlash(ps)
	wd := filepath.FromSlash(wds)
	if filepath.IsAbs(op) {
		p = op
	} else {
		p = filepath.Join(wd, op)
	}
	return
}

// Quick conversion of path separator
func getPath(dg dataGetter, k string) (p string, d diag.Diagnostics) {
	p, d = getString(dg, k)
	p = filepath.FromSlash(p)
	return
}

// Convert OS separator to '/' for relative dschema.
// Intended for state storage.
func normalizePath(p string) string {
	p = filepath.Clean(filepath.FromSlash(p))
	if filepath.IsAbs(p) {
		return p
	} else {
		return strings.Join(filepath.SplitList(p), "/")
	}
}

// Normalize path before storage
func pathStateFunc(pi interface{}) string {
	p, ok := pi.(string)
	if !ok {
		return p
	}
	return normalizePath(p)
}

// Calculate a cryptographic hash of a path.
// Implemented with the `nix-hash` command, since Nix is almost definitely
// installed for users of this provider.
func HashPath(
	ctx context.Context,
	path string,
) (h string, err error) {
	outb := &bytes.Buffer{}
	cmd := exec.CommandContext(
		ctx,
		patches.NixHash(),
		"--type", "sha256",
		"--base32",
		path,
	)
	cmd.Stdout = outb
	cmd.Stderr = logwriter.New(fmt.Sprintf("[INFO] [nix-hash %s] ", path), nil)
	err = cmd.Run()
	if err != nil {
		return
	}

	h = outb.String()
	return
}

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package packerout_test

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	. "github.com/leocp1/terraform-provider-packernix/src/pkg/packerout"
)

func TestPackerOut(t *testing.T) {
	ts := []struct {
		name     string
		bname    string
		inpath   string
		expected *PackerOut
	}{
		{
			name:  "vultr",
			bname: "vultr",
			inpath: filepath.Join(
				"./testdata",
				"packer-builder-vultr-output.txt",
			),
			expected: &PackerOut{
				BuilderID:  "packer.vultr",
				ID:         "DESIREDID",
				String:     "Vultr Snapshot: /nix/store/scrubbedhash-nixos-system-nixos-20.03post-git (scrubbed)",
				FilesCount: 0,
			},
		},
		{
			name:  "vultr-read",
			bname: "read-vultr",
			inpath: filepath.Join(
				"./testdata",
				"packer-builder-vultr-read-output.txt",
			),
			expected: &PackerOut{
				BuilderID:  "packer.vultr",
				ID:         "",
				String:     "0",
				FilesCount: 0,
			},
		},
	}
	for i, tt := range ts {
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			pin, err := os.Open(tt.inpath)
			if err != nil {
				t.Fatalf(err.Error())
			}
			got := &PackerOut{}
			err = got.ParsePackerOut(nil, pin, tt.bname)
			if err != nil {
				t.Fatalf(err.Error())
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf(
					"expected %#v, but got %#v",
					tt.expected,
					got,
				)
			}
		})
	}
}

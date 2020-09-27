// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider_test

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceEval(t *testing.T) {
	ctx := context.Background()
	// Since this function exits before the tests necessarily run, we just leave
	// the temporary directory undeleted
	td, err := ioutil.TempDir("", "resource_external_test")
	if err != nil {
		t.Skip(err.Error())
	}
	tmplS := struct{ TempDir string }{TempDir: td}
	ts := map[string]resource.TestCase{
		"file": {
			ProviderFactories: ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: ReadConfig(
						t,
						filepath.Join("eval", "file.hcl"),
						tmplS,
					),
					Check: resource.TestCheckResourceAttr(
						"data.packernix_eval.file",
						"out",
						"55",
					),
				},
			},
		},
		"file-restricted": {
			ProviderFactories: ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: ReadConfig(
						t,
						filepath.Join("eval", "file-restricted.hcl"),
						tmplS,
					),
					ExpectError: regexp.MustCompile("exit status"),
				},
			},
		},
		"inline": {
			ProviderFactories: ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: ReadConfig(
						t,
						filepath.Join("eval", "inline.hcl"),
						tmplS,
					),
					Check: resource.TestCheckResourceAttr(
						"data.packernix_eval.inline",
						"out",
						"55",
					),
				},
			},
		},
		"installable": {
			ProviderFactories: ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: ReadConfig(
						t,
						filepath.Join("eval", "installable.hcl"),
						tmplS,
					),
					Check: resource.TestCheckResourceAttr(
						"data.packernix_eval.installable",
						"out",
						"55",
					),
					SkipFunc: FlakeSkipFunc(ctx, t),
				},
			},
		},
		"installable_flake": {
			ProviderFactories: ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: ReadConfig(
						t,
						filepath.Join("eval", "installable_flake.hcl"),
						tmplS,
					),
					Check: resource.TestCheckResourceAttr(
						"data.packernix_eval.installable_flake",
						"out",
						"55",
					),
					SkipFunc: FlakeSkipFunc(ctx, t),
				},
			},
		},
		"installable_flake_path": {
			ProviderFactories: ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: ReadConfig(
						t,
						filepath.Join("eval", "installable_flake_path.hcl"),
						tmplS,
					),
					Check: resource.TestCheckResourceAttr(
						"data.packernix_eval.installable_flake_path",
						"out",
						"55",
					),
					SkipFunc: FlakeSkipFunc(ctx, t),
				},
			},
		},
		"nixpkgs": {
			ProviderFactories: ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: ReadConfig(
						t,
						filepath.Join("eval", "nixpkgs.hcl"),
						tmplS,
					),
					Check: CheckSymlink(
						"data.packernix_eval.nixpkgs",
						filepath.Join(td, "nixpkgs-eval"),
						"out",
					),
				},
			},
		},
		"nixpkgs-flake": {
			ProviderFactories: ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: ReadConfig(
						t,
						filepath.Join("eval", "nixpkgs-flake.hcl"),
						tmplS,
					),
					Check: CheckSymlink(
						"data.packernix_eval.nixpkgs",
						filepath.Join(td, "nixpkgs-eval-flake"),
						"out",
					),
				},
			},
		},
	}
	for k, tt := range ts {
		t.Run(k, func(t *testing.T) {
			resource.ParallelTest(t, tt)
		})
	}
}

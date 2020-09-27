// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.
//
// Requires POSIX shell commands
// +build darwin linux

package provider_test

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceExternal(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Skipf(err.Error())
	}
	ts := map[string]resource.TestCase{
		"echo": {
			IsUnitTest:        true,
			ProviderFactories: ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: ReadConfig(
						t,
						filepath.Join("external", "echo", "echo.hcl"),
						struct{}{},
					),
					Check: resource.TestCheckResourceAttr(
						"data.packernix_external.echo",
						"state",
						"test",
					),
				},
			},
		},
		"env": {
			IsUnitTest:        true,
			ProviderFactories: ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: ReadConfig(
						t,
						filepath.Join("external", "env", "env.hcl"),
						struct{}{},
					),
					Check: resource.TestCheckResourceAttr(
						"data.packernix_external.env",
						"state",
						"testid",
					),
				},
				{
					Config: ReadConfig(
						t,
						filepath.Join("external", "env", "env-defaulted.hcl"),
						struct{}{},
					),
					Check: resource.TestCheckResourceAttr(
						"data.packernix_external.env",
						"state",
						"testid",
					),
				},
			},
		},
		"fail": {
			IsUnitTest:        true,
			ProviderFactories: ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: ReadConfig(
						t,
						filepath.Join("external", "fail", "data-fail.hcl"),
						struct{}{},
					),
					ExpectError: regexp.MustCompile("exit status"),
				},
			},
		},
		"file": {
			IsUnitTest:        true,
			ProviderFactories: ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: ReadConfig(
						t,
						filepath.Join("external", "file", "data-file.hcl"),
						struct{}{},
					),
					Check: resource.TestMatchResourceAttr(
						"data.packernix_external.file",
						"state",
						regexp.MustCompile(regexp.QuoteMeta(
							"./testdata/external/file/data.txt",
							)),
					),
				},
			},
		},
		"noop": {
			IsUnitTest:        true,
			ProviderFactories: ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: ReadConfig(
						t,
						filepath.Join("external", "noop", "data-noop.hcl"),
						struct{}{},
					),
					ExpectError: regexp.MustCompile("not found"),
				},
			},
		},
		"pwd": {
			IsUnitTest:        true,
			ProviderFactories: ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: ReadConfig(
						t,
						filepath.Join("external", "pwd", "pwd.hcl"),
						struct{}{},
					),
					Check: resource.TestCheckResourceAttr(
						"data.packernix_external.pwd",
						"state",
						wd,
					),
				},
			},
		},
		"stderr": {
			IsUnitTest:        true,
			ProviderFactories: ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: ReadConfig(
						t,
						filepath.Join("external", "stderr", "stderr.hcl"),
						struct{}{},
					),
					Check: resource.TestCheckResourceAttr(
						"data.packernix_external.stderr",
						"state",
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

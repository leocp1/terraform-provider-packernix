// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider_test

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/yookoala/realpath"
)

func CheckSymlink(name string, ol string, arg string) resource.TestCheckFunc {
	var op string
	return resource.ComposeAggregateTestCheckFunc(
		func(*terraform.State) (err error) {
			op, err = realpath.Realpath(ol)
			if err != nil {
				return
			}
			err = os.Remove(ol)
			return
		},
		resource.TestMatchResourceAttr(
			name,
			arg,
			regexp.MustCompile(regexp.QuoteMeta(op)),
		),
	)
}

func TestAccDataSourceBuild(t *testing.T) {
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
						filepath.Join("build", "file.hcl"),
						tmplS,
					),
				},
			},
		},
		"outlink": {
			ProviderFactories: ProviderFactories(),
			Steps: []resource.TestStep{
				{
					Config: ReadConfig(
						t,
						filepath.Join("build", "outlink.hcl"),
						tmplS,
					),
					Check: CheckSymlink(
						"data.packernix_build.outlink",
						filepath.Join(td, "result"),
						"out_path",
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
						filepath.Join("build", "installable.hcl"),
						tmplS,
					),
					SkipFunc: FlakeSkipFunc(ctx, t),
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

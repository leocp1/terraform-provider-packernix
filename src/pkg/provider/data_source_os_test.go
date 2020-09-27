// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider_test

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func CheckHostName(ol string, hn string) resource.TestCheckFunc {
	return func(*terraform.State) (err error) {
		hostname, err := ioutil.ReadFile(filepath.Join(ol, "etc", "hostname"))
		if err != nil {
			return
		}
		if hn != strings.TrimSpace(string(hostname)) {
			return errors.New(fmt.Sprintf(
				"/etc/hostname is %#v but %#v was expected",
				string(hostname),
				hn,
			))
		}
		return
	}
}

func TestAccDataSourceOS(t *testing.T) {
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
						filepath.Join("os", "file.hcl"),
						tmplS,
					),
					Check: resource.ComposeAggregateTestCheckFunc(
						CheckHostName(
							filepath.Join(td, "result-file"),
							"tfpnhost",
						),
						CheckSymlink(
							"data.packernix_os.file",
							filepath.Join(td, "result-file"),
							"out_path",
						),
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
						filepath.Join("os", "installable.hcl"),
						tmplS,
					),
					SkipFunc: FlakeSkipFunc(ctx, t),
					Check: resource.ComposeAggregateTestCheckFunc(
						CheckHostName(
							filepath.Join(td, "result-installable"),
							"tfpnhost",
						),
						CheckSymlink(
							"data.packernix_os.installable",
							filepath.Join(td, "result-installable"),
							"out_path",
						),
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

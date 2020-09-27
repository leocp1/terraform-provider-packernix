// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider_test

import (
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func GenImageTest(t *testing.T, p string) resource.TestCase {
	return resource.TestCase{
		ProviderFactories: ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: ReadConfig(
					t,
					filepath.Join("image", "image.hcl"),
					struct{ P string }{P: p},
				),
			},
		},
	}
}

func TestAccResourceImage(t *testing.T) {
	ps := []string{
		"vultr",
	}
	for _, p := range ps {
		t.Run(p, func(t *testing.T) {
			resource.ParallelTest(t, GenImageTest(t, p))
		})
	}
}

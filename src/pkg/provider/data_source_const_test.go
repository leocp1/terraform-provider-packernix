// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider_test

import (
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceConst(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		IsUnitTest:        true,
		ProviderFactories: ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: ReadConfig(
					t,
					filepath.Join("const.hcl"),
					struct{}{},
				),
			},
		},
	})
}

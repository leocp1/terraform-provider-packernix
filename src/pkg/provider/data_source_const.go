// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"path/filepath"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/leocp1/terraform-provider-packernix/src/pkg/patches"
)

func DataSourceConst() *schema.Resource {
	return &schema.Resource{
		Schema:      SchemaConst(),
		ReadContext: ReadConst,
		Description: "Read terraform-provider-packernix constants",
	}
}

func SchemaConst() (m map[string]*schema.Schema) {
	return map[string]*schema.Schema{
		"modules": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Directory of NixOS modules",
		},
		"nix": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Nix executable",
		},
		"packer": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Directory of Packer template generators",
		},
	}
}

func ReadConst(
	ctx context.Context,
	rd *schema.ResourceData,
	i interface{},
) diag.Diagnostics {

	err := rd.Set(
		"nix",
		patches.Nix(),
	)
	if err != nil {
		return diag.FromErr(err)
	}

	share := patches.Share()

	err = rd.Set(
		"modules",
		filepath.Join(share, "nixos", "modules"),
	)
	if err != nil {
		return diag.FromErr(err)
	}

	err = rd.Set(
		"packer",
		filepath.Join(share, "packer"),
	)
	if err != nil {
		return diag.FromErr(err)
	}

	rd.SetId(time.Now().UTC().String())

	return nil
}

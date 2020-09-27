// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Packer/Nix provider
package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/leocp1/terraform-provider-packernix/src/pkg/dschema"
	"github.com/leocp1/terraform-provider-packernix/src/pkg/faillock"
)

// Provider
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: ProviderSchema(),
		ResourcesMap: map[string]*schema.Resource{
			"packernix_external": ResourceExternal(),
			"packernix_image":    ResourceImage(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"packernix_build":    DataSourceBuild(),
			"packernix_const":    DataSourceConst(),
			"packernix_eval":     DataSourceEval(),
			"packernix_external": DataSourceExternal(),
			"packernix_os":       DataSourceOS(),
		},
		ConfigureContextFunc: ConfigureContextFunc,
	}
}

func ProviderSchema() (m map[string]*schema.Schema) {
	m = map[string]*schema.Schema{}
	dschema.AddPSchema(BuildDSchema, m)
	dschema.AddPSchema(EvalDSchema, m)
	dschema.AddPSchema(ExternalDSchema, m)
	dschema.AddPSchema(ImageDSchema, m)
	dschema.AddPSchema(OSDSchema, m)
	return
}

// Context passed to resources. Mainly default directories
type ProviderContext struct {
	DMap map[string]interface{}
	FL   *faillock.Faillock
}

func (c *ProviderContext) ProviderDefaults() map[string]interface{} {
	return c.DMap
}

func NewProviderContext() *ProviderContext {
	return &ProviderContext{
		DMap: map[string]interface{}{},
		FL:   faillock.New(),
	}
}

// Configure context func
func ConfigureContextFunc(
	ctx context.Context,
	rd *schema.ResourceData,
) (c interface{}, d diag.Diagnostics) {
	c = NewProviderContext()

	c, d = dschema.Configure(ctx, BuildDSchema, rd, c)
	if d.HasError() {
		return
	}
	c, d0 := dschema.Configure(ctx, EvalDSchema, rd, c)
	d = append(d, d0...)
	if d.HasError() {
		return
	}
	c, d0 = dschema.Configure(ctx, ExternalDSchema, rd, c)
	d = append(d, d0...)
	if d.HasError() {
		return
	}
	c, d0 = dschema.Configure(ctx, ImageDSchema, rd, c)
	d = append(d, d0...)
	if d.HasError() {
		return
	}
	c, d0 = dschema.Configure(ctx, OSDSchema, rd, c)
	d = append(d, d0...)
	return c, d
}

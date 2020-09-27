// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/leocp1/terraform-provider-packernix/src/pkg/dschema"
)

func ResourceExternal() *schema.Resource {
	return &schema.Resource{
		Schema:        SchemaExternal(),
		CreateContext: CreateExternal,
		ReadContext:   ReadExternal,
		UpdateContext: UpdateExternal,
		DeleteContext: DeleteExternal,
		CustomizeDiff: CustomizeDiffExternal,
		Description:   "Use a external programs as a resource",
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func CreateExternal(
	ctx context.Context,
	rd *schema.ResourceData,
	i interface{},
) (d diag.Diagnostics) {
	cg := &dschema.ConfigGetter{
		Ds: ExternalDSchema,
		Rd: rd,
		Pd: i.(*ProviderContext),
	}
	id, d := RunExternal(ctx, cg, "read", "")
	if d.HasError() {
		return
	}
	if id != "" {
		return append(d, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Resource matching configuration already exists.",
		})
	}

	id, d = RunExternal(ctx, cg, "create", "")
	if d.HasError() {
		return
	}
	if id == "" {
		return append(d, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Program output empty string but no error.",
		})
	}
	d = append(d, cg.SetAll(ctx)...)
	if d.HasError() {
		id = ""
	}
	SetState(rd, id)
	return
}

func ReadExternal(
	ctx context.Context,
	rd *schema.ResourceData,
	i interface{},
) (d diag.Diagnostics) {
	// if path is unset, but ID is, assume we are importing
	_, ok := rd.GetOk("path")
	if !ok && rd.Id() != "" {
		return
	}
	sg := &dschema.StateGetter{
		Ds: ExternalDSchema,
		Rd: rd,
		Pd: i.(*ProviderContext),
	}
	id, d := RunExternal(ctx, sg, "read", rd.Id())
	if d.HasError() {
		return
	}
	SetState(rd, id)
	return
}

func CustomizeDiffExternal(
	ctx context.Context,
	rd *schema.ResourceDiff,
	i interface{},
) (err error) {
	cg := &dschema.ConfigGetter{
		Ds: ExternalDSchema,
		Rd: dschema.ResourceDiffAdapter(rd),
		Pd: i.(*ProviderContext),
	}
	d := cg.SetAll(ctx)
	if d.HasError() {
		return dschema.DiagsToErr(d)
	}
	id, d0 := RunExternal(ctx, cg, "read", "")
	d = append(d, d0...)
	if d.HasError() {
		return dschema.DiagsToErr(d)
	}
	if id == rd.Id() {
		err = rd.SetNew("state", id)
		return
	}
	for k := range ExternalDSchema {
		if rd.HasChange(k) {
			err = rd.ForceNew(k)
			if err != nil {
				return
			}
		}
	}
	err = rd.SetNewComputed("state")
	return
}

func UpdateExternal(
	ctx context.Context,
	rd *schema.ResourceData,
	i interface{},
) diag.Diagnostics {
	cg := &dschema.ConfigGetter{
		Ds: ExternalDSchema,
		Rd: rd,
		Pd: i.(*ProviderContext),
	}
	return cg.SetAll(ctx)
}

func DeleteExternal(
	ctx context.Context,
	rd *schema.ResourceData,
	i interface{},
) (d diag.Diagnostics) {
	sg := &dschema.StateGetter{
		Ds: ExternalDSchema,
		Rd: rd,
		Pd: i.(*ProviderContext),
	}
	_, d = RunExternal(ctx, sg, "delete", rd.Id())
	if d.HasError() {
		return
	}
	rd.SetId("")
	return
}

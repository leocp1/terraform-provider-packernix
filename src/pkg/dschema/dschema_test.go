// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package dschema_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	. "github.com/leocp1/terraform-provider-packernix/src/pkg/dschema"
)

type TestPD struct {
	M map[string]interface{}
}

func (tpd *TestPD) ProviderDefaults() map[string]interface{} {
	return tpd.M
}

func NewTestPD(
	ctx context.Context,
	t *testing.T,
	ds DSchemas,
	raw map[string]interface{},
) (pd ProviderDefaulter, d diag.Diagnostics) {
	pd = &TestPD{M: map[string]interface{}{}}
	ps := map[string]*schema.Schema{}
	AddPSchema(ds, ps)
	prd := schema.TestResourceDataRaw(t, ps, raw)
	pdi, d := Configure(ctx, ds, prd, pd)
	pd = pdi.(ProviderDefaulter)
	return
}

func NewTestRD(
	t *testing.T,
	ds DSchemas,
	raw map[string]interface{},
) (rd *schema.ResourceData) {
	rs := map[string]*schema.Schema{}
	AddSchema(ds, rs)
	rd = schema.TestResourceDataRaw(t, rs, raw)
	return
}

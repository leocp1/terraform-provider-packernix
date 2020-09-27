// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"encoding/json"
	"reflect"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func JSONDiffSuppressFunc(k, o, n string, d *schema.ResourceData) bool {
	var oi interface{}
	var ni interface{}
	err := json.Unmarshal(([]byte)(o), &oi)
	if err != nil {
		return false
	}
	err = json.Unmarshal(([]byte)(n), &ni)
	if err != nil {
		return false
	}
	return reflect.DeepEqual(oi, ni)
}

func JSONValidateDiagFunc(
	i interface{},
	p cty.Path,
) (o interface{}, d diag.Diagnostics) {
	s, ok := i.(string)
	if !ok {
		d = append(d, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Not a string",
		})
		return
	}
	err := json.Unmarshal(([]byte)(s), &o)
	if err != nil {
		d = append(d, diag.FromErr(err)...)
		return
	}
	return
}

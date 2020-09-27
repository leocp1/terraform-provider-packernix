// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package dschema

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Implements a schema that simply passes through the underlying schema.Schema
type GenericDSchema struct {
	// Set to true to create an argument in the provider to use as a default
	HasProviderDefault bool
	// Base schema to use
	Base func() *schema.Schema
	// Transformation before returning on Get
	GetFunc func(dataGetter, string) (interface{}, diag.Diagnostics)
	// How a provider value and a resource value should be merged
	MergeFunc func(p interface{}, r interface{}) interface{}
}

//------------------------------------------------------------------------------

func (gds *GenericDSchema) AddSchema(k string, m map[string]*schema.Schema) {
	val := gds.Base()
	if gds.HasProviderDefault {
		val.Default = nil
		val.DefaultFunc = nil
	}
	m[k] = val
}

//------------------------------------------------------------------------------

func (gds *GenericDSchema) AddPSchema(k string, m map[string]*schema.Schema) {
	if gds.HasProviderDefault {
		m[k] = gds.Base()
	}
}

//------------------------------------------------------------------------------

func (gds *GenericDSchema) Configure(
	ctx context.Context,
	rd resource,
	pd ProviderDefaulter,
	k string,
) (d diag.Diagnostics) {
	var err error
	if gds.HasProviderDefault {
		rdg := &rdGetter{D: rd}
		v := rdg.Get(k)
		if v == nil {
			// Manually call DefaultValue, since it sometimes doesn't seem to be
			// called when the schema is a container type.
			// https://github.com/hashicorp/terraform-plugin-sdk/issues/142
			// TODO: Remove this fix if behaviour is changed upstream
			v, err = gds.Base().DefaultValue()
			if err != nil {
				return diag.FromErr(err)
			}
		}
		if v != nil {
			pd.ProviderDefaults()[k] = v
		}
	}
	return nil
}

//------------------------------------------------------------------------------

type tmpGenPD struct {
	m map[string]interface{}
}

func (t *tmpGenPD) ProviderDefaults() map[string]interface{} {
	return t.m
}

func (gds *GenericDSchema) Resolve(
	ctx context.Context,
	rd resource,
	pd ProviderDefaulter,
	k string,
	useConfig bool,
) (rr interface{}, d diag.Diagnostics) {
	var d0 diag.Diagnostics
	var rdg dataGetter
	if useConfig {
		rdg = &rdGetter{D: rd}
	} else {
		rdg = &stateGetter{D: rd}
	}

	rri := rdg.Get(k)
	if rri != nil {
		rr, d = gds.GetFunc(rdg, k)
		if d.HasError() {
			return
		}
	} else {
		// Manually call DefaultValue, since it sometimes doesn't seem to be
		// called when the schema is a container type.
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/142
		// We use a temporary ProviderDefaulter since it's more convenient than
		// trying to do all the type checking in GetFunc manually.
		// TODO: Remove this fix if behaviour is changed upstream
		ds := &tmpGenPD{m: map[string]interface{}{}}
		v, err := gds.Base().DefaultValue()
		if err != nil {
			d = append(d, diag.FromErr(err)...)
			return
		}
		if v != nil {
			ds.ProviderDefaults()[k] = v
		}
		rr, d0 = gds.GetFunc(&defaultGetter{M: ds}, k)
		d = append(d, d0...)
		if d.HasError() {
			return
		}
	}

	if gds.HasProviderDefault {
		prr, d0 := gds.GetFunc(&defaultGetter{M: pd}, k)
		d = append(d, d0...)
		if d.HasError() {
			return
		}
		rr = gds.MergeFunc(prr, rr)
	}

	return
}

//------------------------------------------------------------------------------

func (gds *GenericDSchema) Get(rr interface{}) (interface{}, diag.Diagnostics) {
	return rr, nil
}

//------------------------------------------------------------------------------

func (gds *GenericDSchema) Set(
	rd resource,
	k string,
	rr interface{},
) (d diag.Diagnostics) {
	return nil
}

//------------------------------------------------------------------------------

// A schema for possibly defaultable strings
func StringDSchema(
	hasProviderDefault bool,
	base func() *schema.Schema,
) DSchema {
	return &GenericDSchema{
		HasProviderDefault: hasProviderDefault,
		Base:               base,
		GetFunc: func(dg dataGetter, k string) (interface{}, diag.Diagnostics) {
			return getString(dg, k)
		},
		MergeFunc: func(p interface{}, r interface{}) interface{} {
			if r.(string) == "" {
				return p
			} else {
				return r
			}
		},
	}
}

// A schema for possibly defaultable string slices
func StringSliceDSchema(
	hasProviderDefault bool,
	base func() *schema.Schema,
) DSchema {
	return &GenericDSchema{
		HasProviderDefault: hasProviderDefault,
		Base:               base,
		GetFunc: func(dg dataGetter, k string) (interface{}, diag.Diagnostics) {
			return getStringSlice(dg, k)
		},
		MergeFunc: func(p interface{}, r interface{}) interface{} {
			if len(r.([]string)) == 0 {
				return p
			} else {
				return r
			}
		},
	}
}

// A schema for possibly defaultable maps to strings
func StringMapDSchema(
	hasProviderDefault bool,
	base func() *schema.Schema,
) DSchema {
	return &GenericDSchema{
		HasProviderDefault: hasProviderDefault,
		Base:               base,
		GetFunc: func(dg dataGetter, k string) (interface{}, diag.Diagnostics) {
			return getStringMap(dg, k)
		},
		MergeFunc: func(p interface{}, r interface{}) interface{} {
			m := map[string]string{}
			for k, v := range p.(map[string]string) {
				m[k] = v
			}
			for k, v := range r.(map[string]string) {
				m[k] = v
			}
			return m
		},
	}
}

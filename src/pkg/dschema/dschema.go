// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// A wrapper around schema.ResourceData/schema.ResourceDiff.
//
// This package allows
//	- the default value of an attribute to be based off of other attributes and
//	the provider config
//	- transformation of values obtained from the state/config on Get()
//
// This is mainly necessary because schema.Schema.DefaultFunc is not passed any
// parameters, so we have to do all these operations at apply-time. (Setting
// global variables for defaults during ConfigureContextFunc is also an option,
// but then all providers launched by the same plugin binary would share the
// same defaults.)
//
// WARNING: attributes that have provider defaults with different DSchemas
// must NOT have the same name
package dschema

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Configuration for an attribute
type DSchema interface {
	// Add a schema (and any dependencies) to a resource
	AddSchema(string, map[string]*schema.Schema)
	// Add a schema (and any dependencies) to a provider
	AddPSchema(string, map[string]*schema.Schema)
	// Configure the provider
	Configure(
		context.Context,
		resource,
		ProviderDefaulter,
		string,
	) diag.Diagnostics
	// Return a structure with data to pass to Get and Set.
	// The last parameter controls whether values should be read from the state
	// (false) or the configuration (true).
	Resolve(
		context.Context,
		resource,
		ProviderDefaulter,
		string,
		bool,
	) (interface{}, diag.Diagnostics)
	// Given the result of a Resolve, return a operation friendly result
	Get(interface{}) (interface{}, diag.Diagnostics)
	// Set the result of a Resolve on the configuration to the state
	Set(resource, string, interface{}) diag.Diagnostics
}

type DSchemas map[string]DSchema

//------------------------------------------------------------------------------

// An interface that provider contexts are expected to implement to store
// default arguments
type ProviderDefaulter interface {
	ProviderDefaults() map[string]interface{}
}

//------------------------------------------------------------------------------

// A convenience interface for writing functions that can easily access
// arguments set in the state or the configuration
type DataGetter interface {
	Get(context.Context, string) (interface{}, diag.Diagnostics)
	Id() string
}

//------------------------------------------------------------------------------

// Struct to get values from the state
type StateGetter struct {
	Ds DSchemas
	Rd resource
	Pd ProviderDefaulter
}

func (sg *StateGetter) Get(
	ctx context.Context,
	k string,
) (rv interface{}, d diag.Diagnostics) {
	rr, d := sg.Ds[k].Resolve(ctx, sg.Rd, sg.Pd, k, false)
	if d.HasError() {
		return
	}
	rv, d0 := sg.Ds[k].Get(rr)
	d = append(d, d0...)
	return
}

func (sg *StateGetter) Id() string {
	return sg.Rd.Id()
}

//------------------------------------------------------------------------------

// Struct to get values from the configuration
type ConfigGetter struct {
	Ds DSchemas
	Rd resource
	Pd ProviderDefaulter
}

func (cg *ConfigGetter) Get(
	ctx context.Context,
	k string,
) (rv interface{}, d diag.Diagnostics) {
	rr, d := cg.Ds[k].Resolve(ctx, cg.Rd, cg.Pd, k, true)
	if d.HasError() {
		return
	}
	rv, d0 := cg.Ds[k].Get(rr)
	d = append(d, d0...)
	return
}

func (cg *ConfigGetter) Id() string {
	return cg.Rd.Id()
}

func (cg *ConfigGetter) Set(
	ctx context.Context,
	k string,
) (d diag.Diagnostics) {
	rr, d := cg.Ds[k].Resolve(ctx, cg.Rd, cg.Pd, k, true)
	if d.HasError() {
		return
	}
	d = append(d, cg.Ds[k].Set(cg.Rd, k, rr)...)
	return
}

func (cg *ConfigGetter) SetAll(
	ctx context.Context,
) (d diag.Diagnostics) {
	for k := range cg.Ds {
		d = append(d, cg.Set(ctx, k)...)
		if d.HasError() {
			return
		}
	}
	return
}

//------------------------------------------------------------------------------

// Add all schemas to resource
func AddSchema(ds DSchemas, m map[string]*schema.Schema) {
	for k, v := range ds {
		v.AddSchema(k, m)
	}
}

// Add all schemas to provider
func AddPSchema(ds DSchemas, m map[string]*schema.Schema) {
	for k, v := range ds {
		v.AddPSchema(k, m)
	}
}

// Generate a ConfigureContextFunc from DSchemas
func Configure(
	ctx context.Context, // context
	ds DSchemas, // schema configuration
	rd resource, // passed provider configuration
	ic interface{}, // initial state of context
) (i interface{}, d diag.Diagnostics) {
	i = ic
	c, ok := ic.(ProviderDefaulter)
	if !ok {
		d = append(d, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Provider error: provider config not a ProviderDefaulter",
		})
		return
	}
	for k, v := range ds {
		d = append(d, v.Configure(ctx, rd, c, k)...)
	}
	i = c
	return
}

// Convert diag.Diagnostics to an error
func DiagsToErr(ds diag.Diagnostics) error {
	for i := len(ds) - 1; i >= 0; i-- {
		d := ds[i]
		if d.Severity == diag.Error {
			if d.AttributePath != nil {
				return d.AttributePath.NewErrorf(d.Summary)
			} else {
				return errors.New(d.Summary)
			}
		}
	}
	return nil
}

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package dschema

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Implements a schema that produces paths.
// Sets in the provider:
//	- working_dir
//	- key
// Sets in the resource:
//	- working_dir
//	- key
//	- path_uses_provider_wd
//	- path_hashes
// Get returns a resolved absolute path.
// Relative paths normalized to use '/' as the separator.
type PathDSchema struct {
	// Set to true to create an argument in the provider to use as a default
	HasProviderDefault bool
	// Set to true to skip the path nix-hash check on get
	SkipHashCheck bool
	// Set to true to override Resolve()'s useConfig parameter and perform
	// Get()s directly. Mostly intended for testing, though it should also work
	// for data sources.
	GetFromConfig bool

	// Passed to schema.Schema
	Required      bool
	Optional      bool
	ForceNew      bool
	ConflictsWith []string
	ExactlyOneOf  []string
	AtLeastOneOf  []string
	RequiredWith  []string
	Description   string
	DefaultFunc   schema.SchemaDefaultFunc
}

//------------------------------------------------------------------------------

// Schema to set which paths should resolve with the provider working_dir.
func PUPWDSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeMap,
		Computed:    true,
		Description: "Should the path resolve with the provider working_dir?",
		Elem:        schema.TypeBool,
	}
}

// Schema to store path hashes
func PathHashSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeMap,
		Computed:    true,
		Description: "Hashes of paths",
		Elem:        schema.TypeString,
	}
}

func (pds *PathDSchema) AddSchema(k string, m map[string]*schema.Schema) {
	m[k] = &schema.Schema{
		Type:      schema.TypeString,
		StateFunc: pathStateFunc,

		Required:      pds.Required,
		Optional:      pds.Optional,
		ForceNew:      pds.ForceNew,
		ConflictsWith: pds.ConflictsWith,
		ExactlyOneOf:  pds.ExactlyOneOf,
		AtLeastOneOf:  pds.AtLeastOneOf,
		RequiredWith:  pds.RequiredWith,
		Description:   pds.Description,
	}
	if !pds.HasProviderDefault {
		m[k].DefaultFunc = pds.DefaultFunc
	}
	(&WDDSchema{}).AddSchema(k, m)
	_, ok := m["path_uses_provider_wd"]
	if !ok {
		m["path_uses_provider_wd"] = PUPWDSchema()
	}
	if !pds.SkipHashCheck {
		_, ok = m["path_hashes"]
		if !ok {
			m["path_hashes"] = PathHashSchema()
		}
	}
}

//------------------------------------------------------------------------------

func (pds *PathDSchema) AddPSchema(k string, m map[string]*schema.Schema) {
	if pds.HasProviderDefault {
		m[k] = &schema.Schema{
			Type:      schema.TypeString,
			StateFunc: pathStateFunc,

			Required:    pds.Required,
			Optional:    pds.Optional,
			Description: pds.Description,
			DefaultFunc: pds.DefaultFunc,
		}
	}
	(&WDDSchema{}).AddPSchema(k, m)
}

//------------------------------------------------------------------------------

func (pds *PathDSchema) Configure(
	ctx context.Context,
	rd resource,
	pd ProviderDefaulter,
	k string,
) (d diag.Diagnostics) {
	d = (&WDDSchema{}).Configure(ctx, rd, pd, k)
	if d.HasError() {
		return
	}
	if pds.HasProviderDefault {
		p, d0 := getPath(&rdGetter{D: rd}, k)
		d = append(d, d0...)
		if d.HasError() {
			return
		}
		if p != "" {
			pd.ProviderDefaults()[k] = p
		}
	}
	return
}

//------------------------------------------------------------------------------

func pupwdDiag(bd diag.Diagnostics, s string) diag.Diagnostics {
	return append(bd, diag.Diagnostic{
		Severity:      diag.Error,
		AttributePath: cty.GetAttrPath("path_uses_provider_wd"),
		Summary:       s,
	})
}

func phDiag(bd diag.Diagnostics, s string) diag.Diagnostics {
	return append(bd, diag.Diagnostic{
		Severity:      diag.Error,
		AttributePath: cty.GetAttrPath("path_hashes"),
		Summary:       s,
	})
}

type pathResolveResult struct {
	// resolved absolute path
	Absolute string
	// True if provider working_dir was used
	UsePWD bool
	// nix-hash of path
	Hash string
}

func getPUPWD(
	rd dataGetter,
	pname string,
) (usepwd bool, d diag.Diagnostics) {
	pupwd, d := getMap(rd, "path_uses_provider_wd")
	if d.HasError() {
		return
	}
	usepwdi, ok := pupwd[pname]
	if !ok {
		d = pupwdDiag(d, fmt.Sprintf("Provider error: %s unset", pname))
		return
	}
	usepwd, ok = usepwdi.(bool)
	if !ok {
		d = pupwdDiag(d, fmt.Sprintf(
			"Provider error: %s value not bool",
			pname,
		))
		return
	}
	return
}

func checkHash(
	dg dataGetter,
	pname string,
	hash string,
) (d diag.Diagnostics) {
	oh, d0 := getStringMap(dg, "path_hashes")
	d = append(d, d0...)
	if d.HasError() {
		return
	}
	eh, ok := oh[pname]
	if !ok {
		d = phDiag(d, fmt.Sprintf("no hash stored for %s", pname))
		return
	}
	if eh != hash {
		d = phDiag(d, fmt.Sprintf("hash mismatch for %s", pname))
		return
	}
	return
}

func (pds *PathDSchema) Resolve(
	ctx context.Context,
	rd resource,
	pd ProviderDefaulter,
	k string,
	useConfig bool,
) (interface{}, diag.Diagnostics) {
	rr := &pathResolveResult{}
	var d diag.Diagnostics

	var rdg dataGetter
	if pds.GetFromConfig || useConfig {
		rdg = &rdGetter{D: rd}
	} else {
		rdg = &stateGetter{D: rd}
	}

	pp, d0 := getPath(&defaultGetter{M: pd}, k)
	ppSet := (pp != "")
	d = append(d, d0...)
	if d.HasError() {
		return rr, d
	}

	rp, d0 := getPath(rdg, k)
	rpSet := (rp != "")
	d = append(d, d0...)
	if d.HasError() {
		return rr, d
	}

	var p string
	if !rpSet && !ppSet {
		return rr, d
	} else if rpSet {
		p = rp
	} else {
		p = pp
	}

	wdrr, d0 := (&WDDSchema{}).Resolve(ctx, rd, pd, "working_dir", useConfig)
	d = append(d, d0...)
	if d.HasError() {
		return rr, d
	}

	if useConfig {
		rr.UsePWD = !wdrr.(*wddResolveResult).WdSet || !rpSet
	} else {
		rr.UsePWD, d0 = getPUPWD(rdg, k)
		d = append(d, d0...)
		if d.HasError() {
			return rr, d
		}
	}

	if rr.UsePWD {
		rr.Absolute = readRelPath(p, wdrr.(*wddResolveResult).Pwd)
	} else {
		rr.Absolute = readRelPath(p, wdrr.(*wddResolveResult).Wd)
	}

	if !pds.SkipHashCheck {
		hash, err := HashPath(ctx, rr.Absolute)
		rr.Hash = hash
		if !useConfig {
			if err != nil {
				d = phDiag(d, err.Error())
				return rr, d
			}
			d = append(d, checkHash(rdg, k, hash)...)
		}
	}

	return rr, d
}

//------------------------------------------------------------------------------

func (pds *PathDSchema) Get(
	rr interface{},
) (interface{}, diag.Diagnostics) {
	return rr.(*pathResolveResult).Absolute, nil
}

//------------------------------------------------------------------------------

func setPUPWD(
	rd resource,
	pname string,
	val interface{},
) (d diag.Diagnostics) {
	pupwd, d := getMap(&rdGetter{D: rd}, "path_uses_provider_wd")
	if d.HasError() {
		return
	}
	if val == nil {
		delete(pupwd, pname)
	} else {
		pupwd[pname] = val
	}
	err := rd.Set("path_uses_provider_wd", pupwd)
	if err != nil {
		d = pupwdDiag(d, err.Error())
		return
	}
	return
}

// Set the hash in the schema to be the hash of the system path
func setPathHash(
	rd resource,
	pname string,
	hash string,
) (d diag.Diagnostics) {
	phi, d := getMap(rd, "path_hashes")
	if d.HasError() {
		return
	}
	phi[pname] = hash
	err := rd.Set("path_hashes", phi)
	if err != nil {
		d = phDiag(d, err.Error())
	}
	return
}

func (pds *PathDSchema) Set(
	rd resource,
	k string,
	rr interface{},
) (d diag.Diagnostics) {
	prr := rr.(*pathResolveResult)
	if prr.Absolute == "" {
		d = setPUPWD(rd, k, nil)
		return
	}
	d = setPUPWD(rd, k, prr.UsePWD)
	if d.HasError() {
		return
	}

	if !pds.SkipHashCheck {
		d = append(d, setPathHash(rd, k, prr.Hash)...)
	}

	return
}

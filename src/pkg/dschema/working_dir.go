// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package dschema

import (
	"context"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yookoala/realpath"
)

// Implements a schema for the working directory
// Sets in the provider:
//	- working_dir
// Sets in the resource:
//	- working_dir
// Get returns a resolved absolute path.
// Relative paths normalized to use '/' as the separator
type WDDSchema struct{}

//------------------------------------------------------------------------------

// Schema for resoure working directory
func WDSchema() *schema.Schema {
	return &schema.Schema{
		Type:      schema.TypeString,
		Optional:  true,
		StateFunc: pathStateFunc,
		Description: `The working_dir for any commands executed.
Also sets root for relative dschema`,
	}
}

func (wds *WDDSchema) AddSchema(k string, m map[string]*schema.Schema) {
	_, ok := m["working_dir"]
	if !ok {
		m["working_dir"] = WDSchema()
	}
}

//------------------------------------------------------------------------------

// os.Getwd wrapper
func DefaultWD() (interface{}, error) {
	return os.Getwd()
}

// Validate absolute dschema and working directories
func validateAbsPath(i interface{}, path cty.Path) (d diag.Diagnostics) {
	p, ok := i.(string)
	if !ok {
		return append(d, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Expected a string type",
		})
	}
	_, err := realpath.Realpath(filepath.FromSlash(p))
	if err != nil {
		return diag.FromErr(err)
	}
	return
}

// Schema for default working directory set in provider config.
func DWDSchema() *schema.Schema {
	return &schema.Schema{
		Type:             schema.TypeString,
		Required:         true,
		DefaultFunc:      DefaultWD,
		ValidateDiagFunc: validateAbsPath,
		StateFunc:        pathStateFunc,
		Description: `Default value for working_dir in resources.
Also sets root for relative dschema in the provider`,
	}
}

func (wds *WDDSchema) AddPSchema(k string, m map[string]*schema.Schema) {
	_, ok := m["working_dir"]
	if !ok {
		m["working_dir"] = DWDSchema()
	}
}

//------------------------------------------------------------------------------

func (wds *WDDSchema) Configure(
	ctx context.Context,
	rd resource,
	pd ProviderDefaulter,
	k string,
) diag.Diagnostics {
	var wd string
	var d diag.Diagnostics
	wdDiag := func(s string) diag.Diagnostics {
		return append(d, diag.Diagnostic{
			Severity:      diag.Error,
			AttributePath: cty.GetAttrPath("working_dir"),
			Summary:       s,
		})
	}
	_, ok := pd.ProviderDefaults()["working_dir"]
	if !ok {
		wd, d = getPath(&rdGetter{D: rd}, "working_dir")
		if wd == "" {
			d = wdDiag("unset")
			return d
		}

		oswd, err := os.Getwd()
		if err != nil {
			d = wdDiag(err.Error())
			return d
		}

		wd = readRelPath(wd, oswd)
		_, err = realpath.Realpath(wd)
		if err != nil {
			d = wdDiag(err.Error())
			return d
		}
		pd.ProviderDefaults()["working_dir"] = wd
	}
	return d
}

//------------------------------------------------------------------------------

type wddResolveResult struct {
	// a resolved working directory
	Wd string
	// the provider working directory
	Pwd string
	// true if the working directory was set in the resource
	WdSet bool
}

func (wds *WDDSchema) Resolve(
	ctx context.Context,
	rd resource,
	pd ProviderDefaulter,
	k string,
	useConfig bool,
) (interface{}, diag.Diagnostics) {
	var d = diag.Diagnostics{}
	rr := &wddResolveResult{}
	wdDiag := func(s string) diag.Diagnostics {
		return append(d, diag.Diagnostic{
			Severity:      diag.Error,
			AttributePath: cty.GetAttrPath("working_dir"),
			Summary:       s,
		})
	}

	pwd, d0 := getPath(&defaultGetter{M: pd}, "working_dir")
	d = append(d, d0...)
	if d.HasError() {
		return rr, d
	}
	if pwd == "" {
		d = wdDiag("Provider error: unset")
		return rr, d
	}

	var rdg dataGetter
	if useConfig {
		rdg = &rdGetter{D: rd}
	} else {
		rdg = &stateGetter{D: rd}
	}
	wd, d0 := getPath(rdg, "working_dir")
	d = append(d, d0...)
	if d.HasError() {
		return rr, d
	}

	wdSet := (wd != "")
	if !wdSet {
		wd = pwd
	} else if !filepath.IsAbs(wd) {
		wd = filepath.Join(pwd, wd)
	}

	_, err := realpath.Realpath(wd)
	if err != nil {
		d = wdDiag(err.Error())
		return rr, d
	}

	rr.Wd = wd
	rr.Pwd = pwd
	rr.WdSet = wdSet
	return rr, d
}

//------------------------------------------------------------------------------

func (wds *WDDSchema) Get(rr interface{}) (interface{}, diag.Diagnostics) {
	return rr.(*wddResolveResult).Wd, nil
}

//------------------------------------------------------------------------------

func (wds *WDDSchema) Set(
	rd resource,
	k string,
	rr interface{},
) (d diag.Diagnostics) {
	return nil
}

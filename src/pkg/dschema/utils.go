// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package dschema

import (
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

// An internal equivalent to the DataGetter struct
type dataGetter interface {
	Get(string) interface{}
	Id() string
}

// Get data from the state
type stateGetter struct {
	D resource
}

func (sd *stateGetter) Get(key string) interface{} {
	o, _ := sd.D.GetChange(key)
	return o
}

func (sd *stateGetter) Id() string {
	return sd.D.Id()
}

type configGetter struct {
	D resource
}

func (cd *configGetter) Get(key string) interface{} {
	_, n := cd.D.GetChange(key)
	return n
}

func (cd *configGetter) Id() string {
	return cd.D.Id()
}

// Get data from resource data. Unlike usual Get values, the return value will
// always be nil if unset
type rdGetter struct {
	D resource
}

func (rd *rdGetter) Get(key string) interface{} {
	v, ok := rd.D.GetOk(key)
	if !ok {
		return nil
	}
	return v
}

func (rd *rdGetter) Id() string {
	return rd.D.Id()
}

// Get data from an interface that is expected to contain a ProviderDefaulter
type defaultGetter struct {
	M interface{}
}

// Here we treat nil values as equivalent to unset
func (dg *defaultGetter) Get(key string) interface{} {
	meta, ok := dg.M.(ProviderDefaulter)
	if !ok {
		return nil
	}
	v, ok := meta.ProviderDefaults()[key]
	if !ok {
		return nil
	}
	return v
}

// Id doesn't make sense for defaults
func (dg *defaultGetter) Id() string {
	return ""
}

//------------------------------------------------------------------------------

// Quick cast to bool
func getBool(dg dataGetter, k string) (b bool, d diag.Diagnostics) {
	bi := dg.Get(k)
	if bi == nil {
		return
	}
	b, ok := bi.(bool)
	if !ok {
		d = append(d, diag.Diagnostic{
			Severity:      diag.Error,
			AttributePath: cty.GetAttrPath(k),
			Summary:       "Provider error: not a bool",
		})
	}
	return
}

// Quick cast to string check
func getString(dg dataGetter, k string) (v string, d diag.Diagnostics) {
	vi := dg.Get(k)
	if vi == nil {
		return
	}
	v, ok := vi.(string)
	if !ok {
		d = append(d, diag.Diagnostic{
			Severity:      diag.Error,
			AttributePath: cty.GetAttrPath(k),
			Summary:       "Provider error: not a string",
		})
	}
	return
}

// Quick cast to map[string]interface{}
func getMap(
	dg dataGetter,
	k string,
) (m map[string]interface{}, d diag.Diagnostics) {
	vi := dg.Get(k)
	if vi == nil {
		m = map[string]interface{}{}
		return
	}
	m, ok := vi.(map[string]interface{})
	if !ok {
		m = map[string]interface{}{}
		d = append(d, diag.Diagnostic{
			Severity:      diag.Error,
			AttributePath: cty.GetAttrPath(k),
			Summary:       "Provider error: not a map[string]interface{}",
		})
		return
	}
	return
}

// Quick cast to map[string]string
func getStringMap(
	dg dataGetter,
	key string,
) (m map[string]string, d diag.Diagnostics) {
	m = map[string]string{}
	mi, d := getMap(dg, key)
	if d.HasError() {
		return
	}
	for k, vi := range mi {
		v, ok := vi.(string)
		if !ok {
			d = append(d, diag.Diagnostic{
				Severity:      diag.Error,
				AttributePath: cty.GetAttrPath(key).IndexString(k),
				Summary:       "not a string",
			})
		}
		m[k] = v
	}
	return
}

// Quick cast to []interface{}
func getSlice(
	dg dataGetter,
	k string,
) (sl []interface{}, d diag.Diagnostics) {
	sli := dg.Get(k)
	if sli == nil {
		sl = []interface{}{}
		return
	}
	sl, ok := sli.([]interface{})
	if !ok {
		sl = []interface{}{}
		d = append(d, diag.Diagnostic{
			Severity:      diag.Error,
			AttributePath: cty.GetAttrPath(k),
			Summary:       "Provider error: not a []interface{}",
		})
	}
	return
}

// Quick cast to []string
func getStringSlice(
	rd dataGetter,
	k string,
) (sl []string, d diag.Diagnostics) {
	sli, d := getSlice(rd, k)
	sl = make([]string, 0, len(sli))
	for i := range sli {
		s, ok := sli[i].(string)
		if !ok {
			return sl, append(d, diag.Diagnostic{
				Severity:      diag.Error,
				AttributePath: cty.GetAttrPath(k).IndexInt(i),
				Summary:       "not a string",
			})
		}
		sl = append(sl, s)
	}
	return
}

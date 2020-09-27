// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package dschema

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Implements a schema that produces env lists
// Sets in the provider:
//	- env
//	- clear_env
// Sets in the resource:
//	- env
//	- clear_env
// Get returns an env list to pass into os/exec. If clear_env is set to true,
// then the environment will start empty, otherwise it will start with os.Env.
type EnvDSchema struct{}

//------------------------------------------------------------------------------

func EnvSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Optional: true,
		Elem:     schema.TypeString,
		DefaultFunc: func() (interface{}, error) {
			return map[string]interface{}{}, nil
		},
		Description: "Environment variables",
	}
}

func ClearEnvSchema() *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Description: "Whether to clear the environment before run",
	}
}

func (es *EnvDSchema) AddSchema(k string, m map[string]*schema.Schema) {
	_, ok := m["env"]
	if !ok {
		m["env"] = EnvSchema()
	}
	_, ok = m["clear_env"]
	if !ok {
		m["clear_env"] = ClearEnvSchema()
	}
}

//------------------------------------------------------------------------------

func (es *EnvDSchema) AddPSchema(k string, m map[string]*schema.Schema) {
	es.AddSchema(k, m)
}

//------------------------------------------------------------------------------

func (es *EnvDSchema) Configure(
	ctx context.Context,
	rd resource,
	pd ProviderDefaulter,
	k string,
) (d diag.Diagnostics) {
	rdg := &rdGetter{D: rd}
	em, d := getMap(rdg, "env")
	if d.HasError() {
		return
	}
	ce, d0 := getBool(rdg, "clear_env")
	d = append(d, d0...)
	if d.HasError() {
		return
	}
	pd.ProviderDefaults()["clear_env"] = ce
	pd.ProviderDefaults()["env"] = em
	return
}

//------------------------------------------------------------------------------

type envdResolveResult struct {
	L []string
	C bool
}

func (es *EnvDSchema) Resolve(
	ctx context.Context,
	rd resource,
	pd ProviderDefaulter,
	k string,
	useConfig bool,
) (interface{}, diag.Diagnostics) {

	var ce bool
	var d0 diag.Diagnostics
	var d diag.Diagnostics
	rr := &envdResolveResult{}

	var rdg dataGetter
	if useConfig {
		rdg = &rdGetter{D: rd}
	} else {
		rdg = &stateGetter{D: rd}
	}

	pg := &defaultGetter{M: pd}

	// Take the or of the provider and resource values to determine whether we
	// should clear the environment before adding the variables.
	// Ideally, clear_env would behave more like a default (the provider value
	// is used if the resource value is unset), but checking if a boolean
	// argument is unset or false using the schema API is tricky.
	pce, d0 := getBool(pg, "clear_env")
	d = append(d, d0...)
	if d.HasError() {
		return rr, d
	}
	rce, d0 := getBool(rdg, "clear_env")
	d = append(d, d0...)
	if d.HasError() {
		return rr, d
	}
	ce = pce || rce

	pem, d0 := getStringMap(pg, "env")
	d = append(d, d0...)
	if d.HasError() {
		return rr, d
	}
	rem, d0 := getStringMap(rdg, "env")
	d = append(d, d0...)
	if d.HasError() {
		return rr, d
	}

	var el []string
	if ce {
		el = make([]string, 0, len(rem)+len(pem))
	} else {
		osenv := os.Environ()
		el = make([]string, 0, len(osenv)+len(rem)+len(pem))
		el = append(el, osenv...)
	}

	for k, v := range pem {
		el = append(el, fmt.Sprintf("%s=%s", k, v))
	}
	for k, v := range rem {
		el = append(el, fmt.Sprintf("%s=%s", k, v))
	}

	rr.L = el
	rr.C = ce

	return rr, d
}

//------------------------------------------------------------------------------

func (es *EnvDSchema) Get(rr interface{}) (interface{}, diag.Diagnostics) {
	return rr.(*envdResolveResult).L, nil
}

//------------------------------------------------------------------------------

func (es *EnvDSchema) Set(
	rd resource,
	k string,
	rr interface{},
) (d diag.Diagnostics) {
	return nil
}

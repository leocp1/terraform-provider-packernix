// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/leocp1/terraform-provider-packernix/src/pkg/dschema"
	"github.com/leocp1/terraform-provider-packernix/src/pkg/logwriter"
)

func DataSourceExternal() *schema.Resource {
	return &schema.Resource{
		Schema:      SchemaExternal(),
		ReadContext: ReadDataExternal,
		Description: "Use an external program as a data source",
	}
}

var ExternalDSchema = map[string]dschema.DSchema{
	"path": &dschema.PathDSchema{
		Required:    true,
		Description: `A path with executable(s) in $path/bin`,
	},
	"options": dschema.StringSliceDSchema(
		false,
		func() *schema.Schema {
			return &schema.Schema{
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				DefaultFunc: func() (interface{}, error) {
					return []interface{}{}, nil
				},
				Description: "A list of options to pass to the program(s)",
			}
		},
	),
	"env":         &dschema.EnvDSchema{},
	"working_dir": &dschema.WDDSchema{},
}

func SchemaExternal() (m map[string]*schema.Schema) {
	m = map[string]*schema.Schema{
		"state": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "A string uniquely identifying the data source.",
		},
	}
	dschema.AddSchema(ExternalDSchema, m)
	return
}

func RunExternal(
	ctx context.Context,
	rd dschema.DataGetter,
	op string,
	initId string,
) (id string, d diag.Diagnostics) {
	id = initId

	wd, d := rd.Get(ctx, "working_dir")
	if d.HasError() {
		return
	}

	p, d0 := rd.Get(ctx, "path")
	d = append(d, d0...)
	if d.HasError() {
		return
	}

	opts, d0 := rd.Get(ctx, "options")
	d = append(d, d0...)
	if d.HasError() {
		return
	}

	env, d0 := rd.Get(ctx, "env")
	d = append(d, d0...)
	if d.HasError() {
		return
	}

	exe := filepath.Join(p.(string), "bin", op)
	inb := bytes.NewBufferString(id)
	outb := &bytes.Buffer{}
	log.Printf("[DEBUG] %#v %#v", exe, opts)
	cmd := exec.CommandContext(ctx, exe, opts.([]string)...)
	cmd.Stdin = inb
	cmd.Stdout = outb
	cmd.Stderr = logwriter.New(fmt.Sprintf("[INFO] [%s] ", exe), nil)
	cmd.Dir = wd.(string)
	cmd.Env = env.([]string)
	err := cmd.Run()
	d = exeFail(d, exe, opts.([]string), err)
	if d.HasError() {
		return
	}

	id = outb.String()

	return
}

// Set both the state and id
func SetState(rd *schema.ResourceData, id string) {
	// this should never fail. if it does set id to "" so we're not left with
	// partial state
	err := rd.Set("state", id)
	if err != nil {
		id = ""
	}
	rd.SetId(id)
	return
}

func ReadDataExternal(
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
	if id == "" {
		d = append(d, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "resource not found",
		})
	}
	SetState(rd, id)

	return
}

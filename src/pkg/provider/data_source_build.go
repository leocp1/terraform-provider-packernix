// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"bytes"
	"context"
	"log"
	"os/exec"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/leocp1/terraform-provider-packernix/src/pkg/dschema"
	"github.com/leocp1/terraform-provider-packernix/src/pkg/logwriter"
	"github.com/leocp1/terraform-provider-packernix/src/pkg/patches"
)

func DataSourceBuild() *schema.Resource {
	return &schema.Resource{
		Schema:      SchemaBuild(),
		ReadContext: ReadBuild,
		Description: "Build a nix derivation",
	}
}

var BuildDSchema = map[string]dschema.DSchema{
	// primary arguments
	"file":        NixFileDSchema(false),
	"installable": NixInstallableDSchema(false),
	// other arguments
	"arg":         NixArgDSchema(),
	"argstr":      NixArgstrDSchema(),
	"attr":        NixAttrDSchema(),
	"env":         &dschema.EnvDSchema{},
	"flake":       FlakeDSchema(),
	"flake_path":  FlakePathDSchema(),
	"nix_options": NixOptionsDSchema(),
	"nixpkgs":     NixpkgsDSchema(),
	"out_link":    OutLinkDSchema(),
	"working_dir": &dschema.WDDSchema{},
}

func SchemaBuild() (m map[string]*schema.Schema) {
	m = map[string]*schema.Schema{
		"out_path": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Nix store path of built derivation",
		},
	}
	dschema.AddSchema(BuildDSchema, m)
	return
}

func ReadBuild(
	ctx context.Context,
	rd *schema.ResourceData,
	i interface{},
) (d diag.Diagnostics) {
	cg := &dschema.ConfigGetter{
		Ds: BuildDSchema,
		Rd: rd,
		Pd: i.(*ProviderContext),
	}

	// working dir and env
	wd, d := cg.Get(ctx, "working_dir")
	if d.HasError() {
		return
	}
	env, d0 := cg.Get(ctx, "env")
	d = append(d, d0...)
	if d.HasError() {
		return
	}

	// command
	_, flake := rd.GetOk("installable")
	exe := "nix"
	cmdSlice := []string{}
	if flake {
		if !patches.SupportsNixFlake(ctx) {
			return append(d, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "no flake support",
			})
		}
		exe = patches.Nix()
		cmdSlice = append(cmdSlice, "build")
	} else {
		exe = patches.NixBuild()
	}

	// options
	cmdSlice, d0 = AddNixOptions(ctx, cmdSlice, cg, i, !flake, true, flake)
	d = append(d, d0...)
	if d.HasError() {
		return
	}

	// expression
	var inst string
	cmdSlice, _, inst, d0 = AddNixExpression(
		ctx,
		cmdSlice,
		cg,
		i,
		false,
		flake,
	)
	d = append(d, d0...)
	if d.HasError() {
		return
	}

	log.Printf("[DEBUG] %#v %#v", exe, cmdSlice)
	cmd := exec.CommandContext(ctx, exe, cmdSlice...)
	outb := &bytes.Buffer{}
	cmd.Stdout = outb
	cmd.Stderr = logwriter.New("[INFO] [build]", nil)
	cmd.Dir = wd.(string)
	cmd.Env = env.([]string)
	err := cmd.Run()
	d = exeFail(d, exe, cmdSlice, err)
	if d.HasError() {
		return
	}

	outpath, d0 := GetOutPath(ctx, inst, wd, flake, outb.String())
	d = append(d, d0...)
	if d.HasError() {
		return
	}

	err = rd.Set("out_path", outpath)
	if err != nil {
		d = append(d, diag.FromErr(err)...)
		return d
	}
	rd.SetId(outpath)

	return
}

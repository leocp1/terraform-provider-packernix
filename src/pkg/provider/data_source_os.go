// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"bytes"
	"context"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/leocp1/terraform-provider-packernix/src/pkg/dschema"
	"github.com/leocp1/terraform-provider-packernix/src/pkg/logwriter"
	"github.com/leocp1/terraform-provider-packernix/src/pkg/patches"
)

func DataSourceOS() *schema.Resource {
	return &schema.Resource{
		Schema:      SchemaOS(),
		ReadContext: ReadOS,
		Description: "Build a NixOS configuration",
	}
}

var OSDSchema = map[string]dschema.DSchema{
	// primary arguments
	"file":        NixFileDSchema(false),
	"installable": NixInstallableDSchema(false),
	// other arguments
	"arg":         NixArgDSchema(),
	"argstr":      NixArgstrDSchema(),
	"build_path":  BuildPathDSchema(),
	"config":      NixOSConfigDSchema(),
	"env":         &dschema.EnvDSchema{},
	"flake":       FlakeDSchema(),
	"flake_path":  FlakePathDSchema(),
	"nix_options": NixOptionsDSchema(),
	"nixpkgs":     NixpkgsDSchema(),
	"out_link":    OutLinkDSchema(),
	"working_dir": &dschema.WDDSchema{},
}

func SchemaOS() (m map[string]*schema.Schema) {
	m = map[string]*schema.Schema{
		"out_path": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Nix store path of built NixOS configuration",
		},
	}
	dschema.AddSchema(OSDSchema, m)
	return
}

func ReadOS(
	ctx context.Context,
	rd *schema.ResourceData,
	i interface{},
) (d diag.Diagnostics) {
	var err error
	cg := &dschema.ConfigGetter{
		Ds: OSDSchema,
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
	cmdSlice, d0 = AddNixOptions(ctx, cmdSlice, cg, i, false, true, flake)
	d = append(d, d0...)
	if d.HasError() {
		return
	}

	// build path
	buildPathi, d0 := cg.Get(ctx, "build_path")
	d = append(d, d0...)
	if d.HasError() {
		return
	}
	buildPath := buildPathi.(string)
	if buildPath == "" {
		buildPath, err = ioutil.TempDir(
			"",
			"terraform-provider-packernix-nixos-build",
		)
		if err != nil {
			d = append(d, diag.FromErr(err)...)
			return d
		}
		defer os.RemoveAll(buildPath)
	}
	err = os.MkdirAll(buildPath, 0700)
	if err != nil {
		d = append(d, diag.FromErr(err)...)
		return
	}
	cmdSlice = append(cmdSlice, "-I", buildPath)

	// tfpn config
	d = append(d, GenTFPNConfig(ctx, cg, buildPath)...)
	if d.HasError() {
		return
	}

	// Lock generated file
	var tmplFile string
	if flake {
		tmplFile = filepath.Join(buildPath, "flake.nix")
	} else {
		tmplFile = filepath.Join(buildPath, "default.nix")
	}
	tmplUL, d0 := i.(*ProviderContext).FL.TryLock(tmplFile)
	d = append(d, d0...)
	if d.HasError() {
		return
	}
	defer tmplUL.Unlock()

	// Generated Nix files
	if flake {
		cmdSlice, d0 = GenNixOSFlake(ctx, cg, buildPath, cmdSlice)
	} else {
		cmdSlice, d0 = GenNixOSFile(ctx, cg, buildPath, cmdSlice)
	}
	d = append(d, d0...)
	if d.HasError() {
		return
	}

	// expression
	inst := buildPath + "#nixosConfiguration.config.system.build.toplevel"
	if flake {
		cmdSlice = append(cmdSlice, inst)
	} else {
		cmdSlice = append(cmdSlice, buildPath)
	}

	log.Printf("[DEBUG] %#v %#v", exe, cmdSlice)
	cmd := exec.CommandContext(ctx, exe, cmdSlice...)
	outb := &bytes.Buffer{}
	cmd.Stdout = outb
	cmd.Stderr = logwriter.New("[INFO] [os]", nil)
	cmd.Dir = wd.(string)
	cmd.Env = env.([]string)
	err = cmd.Run()
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

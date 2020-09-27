// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/yookoala/realpath"

	"github.com/leocp1/terraform-provider-packernix/src/pkg/dschema"
	"github.com/leocp1/terraform-provider-packernix/src/pkg/logwriter"
	"github.com/leocp1/terraform-provider-packernix/src/pkg/patches"
)

// Primary arguments

func NixFileDSchema(hasinline bool) dschema.DSchema {
	var eoo []string
	if hasinline {
		eoo = []string{"file", "installable", "inline"}
	} else {
		eoo = []string{"file", "installable"}
	}
	return &dschema.PathDSchema{
		Optional:      true,
		ExactlyOneOf:  eoo,
		SkipHashCheck: true,
		Description:   "The path to the Nix expression to evaluate.",
	}
}

func NixInlineDSchema() dschema.DSchema {
	return dschema.StringDSchema(
		false,
		func() *schema.Schema {
			return &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"file", "inline", "installable"},
				Description:  "String containing Nix expression to evaluate",
			}
		},
	)
}

func NixInstallableDSchema(hasinline bool) dschema.DSchema {
	var eoo []string
	if hasinline {
		eoo = []string{"file", "installable", "inline"}
	} else {
		eoo = []string{"file", "installable"}
	}
	return dschema.StringDSchema(
		false,
		func() *schema.Schema {
			return &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: eoo,
				Description:  "A Nix flake style installable",
			}
		},
	)
}

// Options

func NixArgDSchema() dschema.DSchema {
	return dschema.StringMapDSchema(
		false,
		func() *schema.Schema {
			return &schema.Schema{
				Type:     schema.TypeMap,
				Elem:     schema.TypeString,
				Optional: true,
				DefaultFunc: func() (interface{}, error) {
					return map[string]interface{}{}, nil
				},
				Description: "Arguments to pass to Nix expression",
			}
		},
	)
}

func NixArgstrDSchema() dschema.DSchema {
	return dschema.StringMapDSchema(
		false,
		func() *schema.Schema {
			return &schema.Schema{
				Type:     schema.TypeMap,
				Elem:     schema.TypeString,
				Optional: true,
				DefaultFunc: func() (interface{}, error) {
					return map[string]interface{}{}, nil
				},
				Description: "String arguments to pass to Nix expression",
			}
		},
	)
}

func NixAttrDSchema() dschema.DSchema {
	return dschema.StringDSchema(
		false,
		func() *schema.Schema {
			return &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ConflictsWith: []string{
					"installable",
				},
				Description: "An attribute from the Nix expression",
			}
		},
	)
}

func BuildPathDSchema() dschema.DSchema {
	return &dschema.PathDSchema{
		Optional:      true,
		SkipHashCheck: true,
		Description:   "A directory to place build artifacts",
	}
}

func FlakeDSchema() dschema.DSchema {
	return dschema.StringDSchema(
		true,
		func() *schema.Schema {
			return &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"flake_path"},
				Description:   `A Nix flake. Prepended to "installable"`,
			}
		},
	)
}

func FlakePathDSchema() dschema.DSchema {
	return &dschema.PathDSchema{
		Optional:           true,
		ConflictsWith:      []string{"flake"},
		HasProviderDefault: true,
		SkipHashCheck:      true,
		Description:        `A path to a nix flake. Prepended to "installable"`,
	}
}

func NixpkgsDSchema() dschema.DSchema {
	return &dschema.PathDSchema{
		Optional:           true,
		HasProviderDefault: true,
		SkipHashCheck:      true,
		Description:        "The value of <nixpkgs>. Respects working_dir",
	}
}

func NixOptionsDSchema() dschema.DSchema {
	return dschema.StringMapDSchema(
		true,
		func() *schema.Schema {
			return &schema.Schema{
				Type:     schema.TypeMap,
				Elem:     schema.TypeString,
				Optional: true,
				DefaultFunc: func() (interface{}, error) {
					return map[string]interface{}{
						"restrict-eval": "true",
					}, nil
				},
				Description: "Nix options to set for the command",
			}
		},
	)
}

func NixOSConfigDSchema() dschema.DSchema {
	return dschema.StringDSchema(
		false,
		func() *schema.Schema {
			return &schema.Schema{
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: JSONDiffSuppressFunc,
				ValidateDiagFunc: func(
					i interface{},
					p cty.Path,
				) (d diag.Diagnostics) {
					_, d = JSONValidateDiagFunc(i, p)
					return
				},
				Description: `Value of NixOS tfpn option encoded as JSON`,
			}
		},
	)
}

func OutLinkDSchema() dschema.DSchema {
	return &dschema.PathDSchema{
		Optional:      true,
		SkipHashCheck: true,
		Description:   "The path of the symlink to the outpath",
	}
}

// Helpers

func AddNixOptions(
	ctx context.Context,
	cmdSlice []string,
	dg dschema.DataGetter,
	i interface{},
	addAttr bool,
	addOutlink bool,
	flake bool,
) (cs []string, d diag.Diagnostics) {
	cs = cmdSlice

	if addAttr {
		attr, d0 := dg.Get(ctx, "attr")
		d = append(d, d0...)
		if d.HasError() {
			return
		}
		if attr.(string) != "" {
			cs = append(cs, "--attr", attr.(string))
		}
	}

	arg, d0 := dg.Get(ctx, "arg")
	d = append(d, d0...)
	if d.HasError() {
		return
	}
	if arg != nil {
		for k, v := range arg.(map[string]string) {
			cs = append(cs, "--arg", k, v)
		}
	}
	argstr, d0 := dg.Get(ctx, "argstr")
	d = append(d, d0...)
	if d.HasError() {
		return
	}
	if argstr != nil {
		for k, v := range argstr.(map[string]string) {
			cs = append(cs, "--argstr", k, v)
		}
	}

	nixpkgs, d0 := dg.Get(ctx, "nixpkgs")
	d = append(d, d0...)
	if d.HasError() {
		return
	}
	if nixpkgs.(string) != "" {
		cs = append(cs, "-I", fmt.Sprintf("nixpkgs=%s", nixpkgs.(string)))
	}

	nixOpts, d0 := dg.Get(ctx, "nix_options")
	d = append(d, d0...)
	if d.HasError() {
		return
	}
	if nixOpts != nil {
		for k, v := range nixOpts.(map[string]string) {
			cs = append(cs, "--option", k, v)
		}
	}

	if addOutlink {
		outlink, d0 := dg.Get(ctx, "out_link")
		d = append(d, d0...)
		if d.HasError() {
			return
		}
		if outlink.(string) != "" {
			cs = append(cs, "--out-link", outlink.(string))
		} else {
			if flake {
				cs = append(cs, "--no-link")
			} else {
				cs = append(cs, "--no-out-link")
			}
		}
	}

	return
}

func ParseFlake(
	ctx context.Context,
	dg dschema.DataGetter,
) (flake string, attr string, d diag.Diagnostics) {
	flakeb := strings.Builder{}
	f, d0 := dg.Get(ctx, "flake")
	d = append(d, d0...)
	if d.HasError() {
		return
	}
	flakeb.WriteString(f.(string))

	fp, d0 := dg.Get(ctx, "flake_path")
	d = append(d, d0...)
	if d.HasError() {
		return
	}
	flakeb.WriteString(fp.(string))

	binst, d0 := dg.Get(ctx, "installable")
	d = append(d, d0...)
	if d.HasError() {
		return
	}
	sinst := strings.SplitN(binst.(string), "#", 2)
	if len(sinst) < 2 {
		d = append(d, diag.Diagnostic{
			Severity:      diag.Error,
			AttributePath: cty.GetAttrPath("installable"),
			Summary:       "Must contain # character",
		})
		return
	}
	flakeb.WriteString(sinst[0])
	flake = flakeb.String()
	attr = sinst[1]
	return
}

func AddNixExpression(
	ctx context.Context,
	cmdSlice []string,
	dg dschema.DataGetter,
	i interface{},
	addInline bool,
	flake bool,
) (cs []string, inb *bytes.Buffer, instStr string, d diag.Diagnostics) {
	cs = cmdSlice
	if flake {
		bflake, attr, d0 := ParseFlake(ctx, dg)
		d = append(d, d0...)
		if d.HasError() {
			return
		}
		instStr = strings.Join([]string{bflake, attr}, "#")
		cs = append(cs, instStr)
	} else {
		file, d0 := dg.Get(ctx, "file")
		d = append(d, d0...)
		if d.HasError() {
			return
		}
		if file != "" {
			cs = append(cs, file.(string))
		} else if addInline {
			inline, d0 := dg.Get(ctx, "inline")
			d = append(d, d0...)
			if d.HasError() {
				return
			}
			if inline.(string) != "" {
				inb = bytes.NewBufferString(inline.(string))
				cs = append(cs, "-")
			}
		}
	}
	return
}

func SetOutLink(
	ctx context.Context,
	dg dschema.DataGetter,
	outjson string,
	wd interface{},
	flake bool,
) (d diag.Diagnostics) {
	out_link, d := dg.Get(ctx, "out_link")
	if out_link == "" || d.HasError() {
		return
	}

	var outi interface{}
	err := json.Unmarshal([]byte(outjson), &outi)
	if err != nil {
		d = append(d, diag.FromErr(err)...)
		return
	}
	out, ok := outi.(string)
	if !ok {
		d = append(d, diag.Diagnostic{
			Severity:      diag.Error,
			AttributePath: cty.GetAttrPath("out"),
			Summary:       "not a string",
		})
		return
	}
	outp, err := realpath.Realpath(out)
	if err != nil {
		d = append(d, diag.FromErr(err)...)
		return
	}

	// TODO: See if nix 3.0 interface has a way to register gc roots
	var exe string
	cmdSlice := []string{}
	exe = patches.NixStore()
	cmdSlice = append(
		cmdSlice,
		"--realise",
		"--add-root", out_link.(string), "--indirect",
	)
	cmdSlice = append(cmdSlice, outp)
	log.Printf("[DEBUG] %#v %#v", exe, cmdSlice)
	cmd := exec.CommandContext(ctx, exe, cmdSlice...)
	cmd.Stderr = logwriter.New("[INFO] [setoutlink]", nil)
	cmd.Dir = wd.(string)
	err = cmd.Run()
	d = exeFail(d, exe, cmdSlice, err)
	if d.HasError() {
		return
	}
	return
}

// See https://github.com/NixOS/nix/pull/2622
// TODO: rewrite this if --raw output is added to nix build directly
func getFlakeOutPath(
	ctx context.Context,
	inst string,
	wd interface{},
) (string, diag.Diagnostics) {
	exe := patches.Nix()
	cmdSlice := []string{
		"eval",
		"--raw",
		inst + ".outPath",
	}
	log.Printf("[DEBUG] %#v %#v", exe, cmdSlice)
	cmd := exec.CommandContext(ctx, exe, cmdSlice...)
	cmd.Stderr = logwriter.New("[INFO] [getoutpath]", nil)
	outb := &bytes.Buffer{}
	cmd.Dir = wd.(string)
	cmd.Stdout = outb
	err := cmd.Run()
	d := exeFail(nil, exe, cmdSlice, err)
	if d.HasError() {
		return "", d
	}

	return outb.String(), d
}

func GetOutPath(
	ctx context.Context,
	inst string,
	wd interface{},
	flake bool,
	cmdOut string,
) (outpath string, d diag.Diagnostics) {
	var err error
	if flake {
		outpath, d = getFlakeOutPath(ctx, inst, wd)
		if d.HasError() {
			return
		}
	} else {
		outpath = strings.TrimSpace(cmdOut)
	}

	_, err = realpath.Realpath(outpath)
	if err != nil {
		d = append(d, diag.FromErr(err)...)
		return
	}
	return
}

func GenTFPNConfig(
	ctx context.Context,
	dg dschema.DataGetter,
	buildPath string,
) (d diag.Diagnostics) {
	cfgi, d0 := dg.Get(ctx, "config")
	d = append(d, d0...)
	config := cfgi.(string)
	if config == "" {
		config = "null"
	}
	err := ioutil.WriteFile(
		filepath.Join(buildPath, "tfpn-config.json"),
		[]byte(config),
		0600,
	)
	if err != nil {
		d = append(d, diag.FromErr(err)...)
	}
	return
}

func GenNixOSFile(
	ctx context.Context,
	dg dschema.DataGetter,
	buildPath string,
	cmdSlice []string,
) (cs []string, d diag.Diagnostics) {
	cs = cmdSlice
	nixpkgsi, d0 := dg.Get(ctx, "nixpkgs")
	d = append(d, d0...)
	if d.HasError() {
		return
	}
	nixpkgs := nixpkgsi.(string)
	if nixpkgs == "" {
		nixpkgs = `(import <nixpkgs> {}).path`
	}

	filei, d0 := dg.Get(ctx, "file")
	d = append(d, d0...)
	if d.HasError() {
		return
	}
	file := filei.(string)
	if file == "" {
		file = "./configuration.nix"
	}

	share := patches.Share()
	tfpnModPath := filepath.Join(
		share,
		"nixos",
		"modules",
	)
	cs = append(cs, "-I", tfpnModPath)
	cfgT, err := template.ParseFiles(
		filepath.Join(share, "nixos", "template", "default.nix"),
	)
	if err != nil {
		d = append(d, diag.FromErr(err)...)
		return
	}
	cfgF, err := os.OpenFile(
		filepath.Join(buildPath, "default.nix"),
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0600,
	)
	if err != nil {
		d = append(d, diag.FromErr(err)...)
		return
	}
	defer cfgF.Close()
	err = cfgT.Execute(
		cfgF,
		struct {
			Nixpkgs     string
			File        string
			TfpnModPath string
		}{
			Nixpkgs:     nixpkgs,
			File:        file,
			TfpnModPath: tfpnModPath,
		},
	)
	if err != nil {
		d = append(d, diag.FromErr(err)...)
		return
	}
	return
}

func GenNixOSFlake(
	ctx context.Context,
	dg dschema.DataGetter,
	buildPath string,
	cmdSlice []string,
) (cs []string, d diag.Diagnostics) {
	cs = cmdSlice

	nixpkgsi, d0 := dg.Get(ctx, "nixpkgs")
	d = append(d, d0...)
	if d.HasError() {
		return
	}
	nixpkgs := nixpkgsi.(string)
	if nixpkgs == "" {
		d = append(d, diag.Diagnostic{
			Severity:      diag.Error,
			AttributePath: cty.GetAttrPath("nixpkgs"),
			Summary:       "Must be set",
		})
		return
	}

	flake, attr, d0 := ParseFlake(ctx, dg)
	d = append(d, d0...)
	if d.HasError() {
		return
	}

	share := patches.Share()
	tfpnModPath := filepath.Join(share, "nixos", "modules")

	flakeT, err := template.ParseFiles(
		filepath.Join(share, "nixos", "template", "flake.nix"),
	)
	if err != nil {
		d = append(d, diag.FromErr(err)...)
		return
	}
	flakeF, err := os.OpenFile(
		filepath.Join(buildPath, "flake.nix"),
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0600,
	)
	if err != nil {
		d = append(d, diag.FromErr(err)...)
		return
	}
	defer flakeF.Close()
	err = flakeT.Execute(
		flakeF,
		struct {
			Attr        string
			Flake       string
			Nixpkgs     string
			TfpnModPath string
		}{
			Attr:        attr,
			Flake:       flake,
			Nixpkgs:     nixpkgs,
			TfpnModPath: tfpnModPath,
		},
	)
	if err != nil {
		d = append(d, diag.FromErr(err)...)
		return
	}

	return
}

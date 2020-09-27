// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/leocp1/terraform-provider-packernix/src/pkg/dschema"
	"github.com/leocp1/terraform-provider-packernix/src/pkg/logwriter"
	"github.com/leocp1/terraform-provider-packernix/src/pkg/packerout"
	"github.com/leocp1/terraform-provider-packernix/src/pkg/patches"
)

func ResourceImage() *schema.Resource {
	return &schema.Resource{
		Schema:        SchemaImage(),
		CreateContext: CreateImage,
		ReadContext:   ReadImage,
		UpdateContext: UpdateImage,
		DeleteContext: DeleteImage,
		CustomizeDiff: CustomizeDiffImage,
		Timeouts: &schema.ResourceTimeout{
			// Image creation can be slow and timing can depend on the cloud
			// provider.
			// Here, we default to a week timeout and rely on the passed Packer
			// template to set a more sane default.
			Create: schema.DefaultTimeout(7 * 24 * time.Hour),
		},
		Description: "A Packer machine image.",
	}
}

var ImageDSchema = map[string]dschema.DSchema{
	"template": dschema.StringDSchema(
		false,
		func() *schema.Schema {
			return &schema.Schema{
				Type:             schema.TypeString,
				Required:         true,
				Description:      "A Packer template",
				DiffSuppressFunc: JSONDiffSuppressFunc,
				ValidateDiagFunc: func(
					i interface{},
					p cty.Path,
				) (d diag.Diagnostics) {
					o, d := JSONValidateDiagFunc(i, p)
					if d.HasError() {
						return
					}
					d = append(d, ValidatePackerTemplate(o)...)
					return
				},
			}
		},
	),
	"build_path":  BuildPathDSchema(),
	"env":         &dschema.EnvDSchema{},
	"working_dir": &dschema.WDDSchema{},
}

func SchemaImage() (m map[string]*schema.Schema) {
	m = map[string]*schema.Schema{
		"builder_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "ID of the builder",
		},
		"image": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "ID of the output image",
		},
	}
	dschema.AddSchema(ImageDSchema, m)
	return
}

func ValidatePackerTemplate(tmpl interface{}) (d diag.Diagnostics) {
	tempDiag := func(s string) diag.Diagnostic {
		return diag.Diagnostic{
			Severity:      diag.Error,
			AttributePath: cty.GetAttrPath("template"),
			Summary:       s,
		}
	}
	mt, ok := tmpl.(map[string]interface{})
	if !ok {
		d = append(d, tempDiag("Template not an attribute set"))
	}
	bsi, ok := mt["builders"]
	if !ok {
		d = append(d, tempDiag("No builders attribute"))
	}
	bs, ok := bsi.([]interface{})
	if !ok {
		d = append(d, tempDiag("builders not a list"))
	}
	if len(bs) != 1 {
		d = append(d, tempDiag("builders does not contain one build"))
	}
	b, ok := bs[0].(map[string]interface{})
	if !ok {
		d = append(d, tempDiag("builder not an attribute set"))
	}
	ti, ok := b["type"]
	if !ok {
		d = append(d, tempDiag("builder does not define a type"))
	}
	_, ok = ti.(string)
	if !ok {
		d = append(d, tempDiag("builder type is not a string"))
	}
	return
}

func MakePackerTemplate(
	ctx context.Context,
	rd dschema.DataGetter,
	tfpath string,
	op string,
) (btype string, d diag.Diagnostics) {
	tmplStr, d0 := rd.Get(ctx, "template")
	d = append(d, d0...)
	if d.HasError() {
		return
	}
	var tmpli interface{}
	err := json.Unmarshal(([]byte)(tmplStr.(string)), &tmpli)
	if err != nil {
		d = append(d, diag.FromErr(err)...)
		return
	}
	d = append(d, ValidatePackerTemplate(tmpli)...)
	if d.HasError() {
		return
	}
	bs := tmpli.(map[string]interface{})["builders"]
	b := bs.([]interface{})[0].(map[string]interface{})
	btype = b["type"].(string)
	if op != "create" {
		// Set builder type
		btype = fmt.Sprintf("%s-%s", op, btype)
		b["type"] = btype
		// Remove all attributes except variables and builders
		otmpli := tmpli
		tmpli = map[string]interface{}{}
		bs = otmpli.(map[string]interface{})["builders"]
		tmpli.(map[string]interface{})["builders"] = bs
		vs, ok := otmpli.(map[string]interface{})["variables"]
		if ok {
			tmpli.(map[string]interface{})["variables"] = vs
		}
	}

	tmplF, err := os.OpenFile(
		tfpath,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		0600,
	)
	if err != nil {
		d = append(d, diag.FromErr(err)...)
		return
	}
	defer tmplF.Close()
	jsenc := json.NewEncoder(tmplF)
	err = jsenc.Encode(tmpli)
	if err != nil {
		d = append(d, diag.FromErr(err)...)
		return
	}
	return
}

func RunPacker(
	ctx context.Context,
	rd dschema.DataGetter,
	i interface{},
	op string,
) (pout *packerout.PackerOut, d diag.Diagnostics) {

	var err error

	pout = &packerout.PackerOut{}

	wd, d0 := rd.Get(ctx, "working_dir")
	d = append(d, d0...)
	if d.HasError() {
		return
	}
	bdi, d0 := rd.Get(ctx, "build_path")
	d = append(d, d0...)
	if d.HasError() {
		return
	}
	bd := bdi.(string)
	if bd == "" {
		bd, err = ioutil.TempDir(
			"",
			"terraform-provider-packernix-packer-build",
		)
		if err != nil {
			d = append(d, diag.FromErr(err)...)
			return
		}
		defer os.RemoveAll(bd)
	}
	err = os.MkdirAll(bd, 0700)
	if err != nil {
		d = append(d, diag.FromErr(err)...)
		return
	}
	tfpath := filepath.Join(bd, op+".json")
	env, d0 := rd.Get(ctx, "env")
	d = append(d, d0...)
	if d.HasError() {
		return
	}
	// Lock generated file
	tfUL, d0 := i.(*ProviderContext).FL.TryLock(tfpath)
	d = append(d, d0...)
	if d.HasError() {
		return
	}
	defer tfUL.Unlock()

	// Modify template
	btype, d0 := MakePackerTemplate(ctx, rd, tfpath, op)
	d = append(d, d0...)
	if d.HasError() {
		return
	}

	// validate
	exe := patches.Packer()
	cmdSlice := []string{
		"validate",
		tfpath,
	}
	log.Printf("[DEBUG] %#v %#v", exe, cmdSlice)
	cmd := exec.CommandContext(ctx, exe, cmdSlice...)
	cmd.Dir = wd.(string)
	cmd.Env = env.([]string)
	cmd.Stdout = logwriter.New(fmt.Sprintf("[INFO] [%s] ", exe), nil)
	cmd.Stderr = logwriter.New(fmt.Sprintf("[INFO] [%s] ", exe), nil)
	err = cmd.Run()
	d = exeFail(d, exe, cmdSlice, err)
	if d.HasError() {
		return
	}

	// build
	cmdSlice = []string{
		"-machine-readable",
		"build",
		tfpath,
	}
	log.Printf("[DEBUG] %#v %#v", exe, cmdSlice)
	cmd = exec.CommandContext(ctx, exe, cmdSlice...)
	cmd.Dir = wd.(string)
	cmd.Env = env.([]string)
	err = pout.RunPacker(nil, cmd, btype)
	d = exeFail(d, exe, cmdSlice, err)
	if d.HasError() {
		return
	}

	return
}

func PackerOutId(
	pout *packerout.PackerOut,
) (id string) {
	id = fmt.Sprintf(
		"%s.%s",
		pout.BuilderID,
		pout.ID,
	)
	if id == "." {
		id = ""
	}
	return
}

func SetImageId(
	rd *schema.ResourceData,
	pout *packerout.PackerOut,
) {
	id := PackerOutId(pout)
	err := rd.Set("builder_id", pout.BuilderID)
	if err != nil {
		id = ""
	}
	err = rd.Set("image", pout.ID)
	if err != nil {
		id = ""
	}
	rd.SetId(id)
}

func CheckPreexist(
	rd *schema.ResourceData,
) (found bool) {
	bidi, ok := rd.GetOk("builder_id")
	if !ok {
		return
	}
	bid, ok := bidi.(string)
	if !ok || bid == "" {
		return
	}
	iidi, ok := rd.GetOk("image")
	if !ok {
		return
	}
	iid, ok := iidi.(string)
	if !ok || iid == "" {
		return
	}
	pout := &packerout.PackerOut{
		BuilderID: bid,
		ID:        iid,
	}
	SetImageId(rd, pout)
	return true
}

func CreateImage(
	ctx context.Context,
	rd *schema.ResourceData,
	i interface{},
) (d diag.Diagnostics) {
	if CheckPreexist(rd) {
		d = append(d, diag.Diagnostic{
			Severity: diag.Warning,
			Summary: fmt.Sprintf(
				"Setting existing image %s to the state. No new images built.",
				rd.Id(),
			),
		})
		return
	}

	cg := &dschema.ConfigGetter{
		Ds: ImageDSchema,
		Rd: rd,
		Pd: i.(*ProviderContext),
	}
	pout, d0 := RunPacker(ctx, cg, i, "create")
	d = append(d, d0...)
	if d.HasError() {
		return
	}
	d = append(d, cg.SetAll(ctx)...)
	if d.HasError() {
		rd.SetId("")
		return
	}
	SetImageId(rd, pout)
	return
}

func ReadImage(
	ctx context.Context,
	rd *schema.ResourceData,
	i interface{},
) (d diag.Diagnostics) {
	sg := &dschema.StateGetter{
		Ds: ImageDSchema,
		Rd: rd,
		Pd: i.(*ProviderContext),
	}
	pout, d := RunPacker(ctx, sg, i, "read")
	if d.HasError() {
		return
	}
	count, err := strconv.Atoi(pout.String)
	if err != nil {
		d = append(d, diag.FromErr(err)...)
	}
	if count > 1 {
		d = append(d, diag.Diagnostic{
			Severity: diag.Warning,
			Summary: fmt.Sprintf(
				"%d compatible images found.",
				count,
			),
		})
	}
	if count == 0 {
		rd.SetId("")
		return
	}
	SetImageId(rd, pout)
	return
}

func CustomizeDiffImage(
	ctx context.Context,
	rd *schema.ResourceDiff,
	i interface{},
) (err error) {
	cg := &dschema.ConfigGetter{
		Ds: ImageDSchema,
		Rd: dschema.ResourceDiffAdapter(rd),
		Pd: i.(*ProviderContext),
	}
	d := cg.SetAll(ctx)
	if d.HasError() {
		return dschema.DiagsToErr(d)
	}
	pout, d0 := RunPacker(ctx, cg, i, "read")
	d = append(d, d0...)
	if d.HasError() {
		return dschema.DiagsToErr(d)
	}

	// Allow reusing the found image if either
	//	- a new image is being created
	//	- the state matches.
	cid := PackerOutId(pout)
	sid := rd.Id()
	if sid == "" || cid == sid {
		if cid != "" {
			err = rd.SetNew("builder_id", pout.BuilderID)
			if err != nil {
				return
			}
			err = rd.SetNew("image", pout.ID)
		}
		return
	}

	for k := range ImageDSchema {
		if rd.HasChange(k) {
			err = rd.ForceNew(k)
			if err != nil {
				return
			}
		}
	}
	err = rd.SetNewComputed("builder_id")
	if err != nil {
		return
	}
	err = rd.SetNewComputed("image")

	return
}

func UpdateImage(
	ctx context.Context,
	rd *schema.ResourceData,
	i interface{},
) diag.Diagnostics {
	cg := &dschema.ConfigGetter{
		Ds: ImageDSchema,
		Rd: rd,
		Pd: i.(*ProviderContext),
	}
	return cg.SetAll(ctx)
}

func DeleteImage(
	ctx context.Context,
	rd *schema.ResourceData,
	i interface{},
) (d diag.Diagnostics) {
	sg := &dschema.StateGetter{
		Ds: ImageDSchema,
		Rd: rd,
		Pd: i.(*ProviderContext),
	}
	_, d = RunPacker(ctx, sg, i, "delete")
	if d.HasError() {
		return
	}
	rd.SetId("")
	return
}

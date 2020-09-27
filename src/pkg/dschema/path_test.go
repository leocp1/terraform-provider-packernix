// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package dschema_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/leocp1/terraform-provider-packernix/src/pkg/dschema"
)

type PathDSchemaCase struct {
	name     string
	resource map[string]interface{}
	provider map[string]interface{}
	expected string
	hasDiags bool
}

func GenPathDSchemaCases(wd string) (ts []PathDSchemaCase) {
	for i := 0; i < (1 << 6); i++ {
		tt := PathDSchemaCase{
			resource: map[string]interface{}{},
			provider: map[string]interface{}{},
			expected: filepath.Join(wd, "testdata", "d", "file.txt"),
			hasDiags: false,
		}
		var ns []string

		ns = append(ns, "pwdrel:")
		if (i & (1 << 0)) == 0 {
			ns = append(ns, "y")
			tt.provider["working_dir"] = "testdata"
		} else {
			ns = append(ns, "n")
			tt.provider["working_dir"] = filepath.Join(wd, "testdata")
		}
		ns = append(ns, ",")

		rwdrel := (i & (1 << 1)) == 0
		ns = append(ns, "rwdrel:")
		if rwdrel {
			ns = append(ns, "y")
			tt.resource["working_dir"] = "d"
		} else {
			ns = append(ns, "n")
			tt.resource["working_dir"] = filepath.Join(wd, "testdata", "d")
		}
		ns = append(ns, ",")

		usepwd := (i & (1 << 2)) == 0
		ns = append(ns, "wd:")
		if usepwd {
			ns = append(ns, "p")
			delete(tt.resource, "working_dir")
		} else {
			ns = append(ns, "r")
		}
		ns = append(ns, ",")

		ns = append(ns, "pprel:")
		if (i & (1 << 3)) == 0 {
			ns = append(ns, "y")
			tt.provider["optional"] = "file.txt"
		} else {
			ns = append(ns, "n")
			tt.provider["optional"] = filepath.Join(wd, "testdata", "d", "file.txt")
		}
		ns = append(ns, ",")

		ns = append(ns, "rprel:")
		if (i & (1 << 4)) == 0 {
			ns = append(ns, "y")
			tt.resource["optional"] = "file.txt"
		} else {
			ns = append(ns, "n")
			tt.resource["optional"] = filepath.Join(wd, "testdata", "d", "file.txt")
		}
		ns = append(ns, ",")

		ns = append(ns, "p:")
		usepp := (i & (1 << 5)) == 0
		if usepp {
			ns = append(ns, "p")
			delete(tt.resource, "optional")
		} else {
			ns = append(ns, "r")
			tt.provider["optional"] = "wrong"
		}
		ns = append(ns, ",")

		if usepp || usepwd {
			tt.provider["working_dir"] = filepath.Join(
				tt.provider["working_dir"].(string),
				"d",
			)
			_, rwdSet := tt.resource["working_dir"]
			if rwdSet {
				orwd := tt.resource["working_dir"].(string)
				if filepath.IsAbs(orwd) {
					tt.resource["working_dir"] = filepath.Join(
						orwd,
						"../wrong",
					)
				} else {
					tt.resource["working_dir"] = filepath.Join(
						orwd,
						"../../wrong",
					)
				}
			}
		}

		tt.name = strings.Join(ns, "")

		ts = append(ts, tt)
	}
	return
}

func TestPathDSchema(t *testing.T) {
	ctx := context.Background()

	ds := map[string]DSchema{
		"optional": &PathDSchema{
			HasProviderDefault: true,
			GetFromConfig:      true,
			SkipHashCheck:      true,
		},
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Skipf("os.Getwd failed")
	}
	ts := []PathDSchemaCase{
		{
			name:     "unset",
			expected: "",
		},
		{
			name: "nix store path",
			provider: map[string]interface{}{
				"working_dir": wd,
			},
			resource: map[string]interface{}{
				"optional": "/nix/store/hash-name",
			},
			expected: "/nix/store/hash-name",
		},
		{
			name:     "read slashes",
			provider: map[string]interface{}{},
			resource: map[string]interface{}{
				"optional": "./testdata/../testdata/d/file.txt",
			},
			expected: filepath.Join(wd, "testdata", "d", "file.txt"),
		},
	}
	ts = append(ts, GenPathDSchemaCases(wd)...)
	for i, tt := range ts {
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			pd, d := NewTestPD(ctx, t, ds, tt.provider)
			rd := NewTestRD(t, ds, tt.resource)
			cg := &ConfigGetter{Ds: ds, Rd: rd, Pd: pd}
			got, d0 := cg.Get(ctx, "optional")
			d = append(d, d0...)
			if d.HasError() != tt.hasDiags {
				t.Errorf(
					"expected hasDiags: %#v, but got %#v",
					tt.hasDiags,
					d,
				)
			}
			if d.HasError() {
				return
			}
			if got != tt.expected {
				t.Errorf(
					"expected %#v, but got %#v",
					tt.expected,
					got,
				)
			}
			d = append(d, cg.SetAll(ctx)...)
			if d.HasError() {
				t.Errorf("Set failed: %#v", d)
				return
			}
			p, d0 := cg.Get(ctx, "optional")
			d = append(d, d0...)
			if d.HasError() {
				t.Errorf("Second Get failed: %#v", d)
				return
			}
			if tt.expected != p {
				t.Errorf(
					"Second Get got %#v but expected %#v",
					p,
					tt.expected,
				)
			}
		})
	}
}

func TestHashPath(t *testing.T) {
	ctx := context.Background()

	ds := map[string]DSchema{
		"optional": &PathDSchema{
			GetFromConfig: true,
		},
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Skipf("os.Getwd failed")
	}
	td := filepath.Join(wd, "testdata")
	d := filepath.Join(td, "d")
	wrong := filepath.Join(td, "wrong")
	file := filepath.Join(d, "file.txt")

	pd, diag := NewTestPD(ctx, t, ds, map[string]interface{}{})
	rd := NewTestRD(t, ds, map[string]interface{}{})

	cg := &ConfigGetter{Ds: ds, Rd: rd, Pd: pd}
	// since GetFromConfig is set, Get() will be performed directly, but
	// filepath hashes will be checked as if the values came from the state.
	sg := &StateGetter{Ds: ds, Rd: rd, Pd: pd}

	_, diag = cg.Get(ctx, "optional")
	if diag.HasError() {
		t.Errorf("Check empty test failed")
	}
	rd.Set("optional", d)
	diag = cg.SetAll(ctx)
	if diag.HasError() {
		t.Errorf("Set test failed")
	}
	_, diag = sg.Get(ctx, "optional")
	if diag.HasError() {
		t.Errorf("Check test failed")
	}

	rd.Set("optional", wrong)
	_, diag = sg.Get(ctx, "optional")
	if diag.HasError() {
		t.Errorf("Check dir with same contents test failed")
	}

	rd.Set("optional", file)
	_, diag = sg.Get(ctx, "optional")
	if !diag.HasError() {
		t.Errorf("Check file inside dir failed")
	}

	rd.Set("optional", file)
	diag = cg.SetAll(ctx)
	if diag.HasError() {
		t.Errorf("Set test failed")
	}
	_, diag = sg.Get(ctx, "optional")
	if diag.HasError() {
		t.Errorf("Check test failed")
	}
}

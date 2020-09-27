// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package dschema_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	. "github.com/leocp1/terraform-provider-packernix/src/pkg/dschema"
)

func TestWDDSchema(t *testing.T) {
	ctx := context.Background()

	ds := map[string]DSchema{
		"working_dir": &WDDSchema{},
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Skipf("os.Getwd failed")
	}

	ts := []struct {
		name     string
		provider map[string]interface{}
		resource map[string]interface{}
		expected string
		hasDiags bool
	}{
		{
			name:     "Default",
			provider: map[string]interface{}{},
			resource: map[string]interface{}{},
			expected: wd,
		},
		{
			name:     "Relative resource working_dir",
			provider: map[string]interface{}{},
			resource: map[string]interface{}{
				"working_dir": "testdata",
			},
			expected: filepath.Join(wd, "testdata"),
		},
		{
			name:     "Absolute resource working_dir",
			provider: map[string]interface{}{},
			resource: map[string]interface{}{
				"working_dir": filepath.Join(wd, "testdata"),
			},
			expected: filepath.Join(wd, "testdata"),
		},
		{
			name: "Relative provider working_dir",
			provider: map[string]interface{}{
				"working_dir": "testdata",
			},
			resource: map[string]interface{}{},
			expected: filepath.Join(wd, "testdata"),
		},
		{
			name: "Absolute resource working_dir",
			provider: map[string]interface{}{
				"working_dir": filepath.Join(wd, "testdata"),
			},
			resource: map[string]interface{}{},
			expected: filepath.Join(wd, "testdata"),
		},
		{
			name: "Relative provider and resource working_dir",
			provider: map[string]interface{}{
				"working_dir": ".",
			},
			resource: map[string]interface{}{
				"working_dir": "testdata",
			},
			expected: filepath.Join(wd, "testdata"),
		},
		{
			name:     "resource working_dir does not exist",
			provider: map[string]interface{}{},
			resource: map[string]interface{}{
				"working_dir": filepath.Join(wd, "does", "not", "exist"),
			},
			expected: "",
			hasDiags: true,
		},
		{
			name: "provider working_dir does not exist",
			provider: map[string]interface{}{
				"working_dir": filepath.Join(wd, "does", "not", "exist"),
			},
			resource: map[string]interface{}{},
			expected: "",
			hasDiags: true,
		},
	}
	for i, tt := range ts {
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			pd, d := NewTestPD(ctx, t, ds, tt.provider)
			rd := NewTestRD(t, ds, tt.resource)
			cg := &ConfigGetter{Ds: ds, Rd: rd, Pd: pd}
			got, d0 := cg.Get(ctx, "working_dir")
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
		})
	}
}

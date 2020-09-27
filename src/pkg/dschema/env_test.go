// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package dschema_test

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"

	. "github.com/leocp1/terraform-provider-packernix/src/pkg/dschema"
)

func TestEnvDSchema(t *testing.T) {
	ctx := context.Background()

	ds := map[string]DSchema{
		"env": &EnvDSchema{},
	}

	osenv := os.Environ()

	ts := []struct {
		name     string
		provider map[string]interface{}
		resource map[string]interface{}
		expected []string
		hasDiags bool
	}{
		{
			name:     "Default",
			provider: map[string]interface{}{},
			resource: map[string]interface{}{},
			expected: osenv,
		},
		{
			name: "Clear from provider",
			provider: map[string]interface{}{
				"clear_env": true,
			},
			resource: map[string]interface{}{},
			expected: []string{},
		},
		{
			name: "Clear from resource",
			provider: map[string]interface{}{
				"clear_env": false,
			},
			resource: map[string]interface{}{
				"clear_env": true,
			},
			expected: []string{},
		},
		{
			name: "Resource overrides provider",
			provider: map[string]interface{}{
				"env": map[string]interface{}{
					"A": "first",
				},
				"clear_env": true,
			},
			resource: map[string]interface{}{
				"env": map[string]interface{}{
					"A": "second",
				},
			},
			expected: []string{
				"A=first",
				"A=second",
			},
		},
	}
	for i, tt := range ts {
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			pd, d := NewTestPD(ctx, t, ds, tt.provider)
			rd := NewTestRD(t, ds, tt.resource)
			cg := &ConfigGetter{Ds: ds, Rd: rd, Pd: pd}
			got, d0 := cg.Get(ctx, "env")
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
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf(
					"expected %#v, but got %#v",
					tt.expected,
					got,
				)
			}
		})
	}
}

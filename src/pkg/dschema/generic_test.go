// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package dschema_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	. "github.com/leocp1/terraform-provider-packernix/src/pkg/dschema"
)

func TestGenericDSchema(t *testing.T) {
	ctx := context.Background()

	ds := map[string]DSchema{
		"ndstring": StringDSchema(
			false,
			func() *schema.Schema {
				return &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
					DefaultFunc: func() (interface{}, error) {
						return "", nil
					},
				}
			},
		),
		"string": StringDSchema(
			false,
			func() *schema.Schema {
				return &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
					DefaultFunc: func() (interface{}, error) {
						return "stringd", nil
					},
				}
			},
		),
		"dstring": StringDSchema(
			true,
			func() *schema.Schema {
				return &schema.Schema{
					Type:     schema.TypeString,
					Optional: true,
					DefaultFunc: func() (interface{}, error) {
						return "dstringd", nil
					},
				}
			},
		),
		"ndslice": StringSliceDSchema(
			false,
			func() *schema.Schema {
				return &schema.Schema{
					Type: schema.TypeList,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Optional: true,
				}
			},
		),
		"slice": StringSliceDSchema(
			false,
			func() *schema.Schema {
				return &schema.Schema{
					Type: schema.TypeList,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Optional: true,
					DefaultFunc: func() (interface{}, error) {
						return []interface{}{"sliced"}, nil
					},
				}
			},
		),
		"dslice": StringSliceDSchema(
			true,
			func() *schema.Schema {
				return &schema.Schema{
					Type: schema.TypeList,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
					Optional: true,
					DefaultFunc: func() (interface{}, error) {
						return []interface{}{"dsliced"}, nil
					},
				}
			},
		),
		"ndmap": StringMapDSchema(
			false,
			func() *schema.Schema {
				return &schema.Schema{
					Type:     schema.TypeMap,
					Elem:     schema.TypeString,
					Optional: true,
				}
			},
		),
		"map": StringMapDSchema(
			false,
			func() *schema.Schema {
				return &schema.Schema{
					Type:     schema.TypeMap,
					Elem:     schema.TypeString,
					Optional: true,
					DefaultFunc: func() (interface{}, error) {
						return map[string]interface{}{"k": "mapv"}, nil
					},
				}
			},
		),
		"dmap": StringMapDSchema(
			true,
			func() *schema.Schema {
				return &schema.Schema{
					Type:     schema.TypeMap,
					Elem:     schema.TypeString,
					Optional: true,
					DefaultFunc: func() (interface{}, error) {
						return map[string]interface{}{"k": "dmapv"}, nil
					},
				}
			},
		),
	}

	ts := []struct {
		name     string
		provider map[string]interface{}
		resource map[string]interface{}
		expected map[string]interface{}
		hasDiags bool
	}{
		{
			name:     "empty",
			provider: map[string]interface{}{},
			resource: map[string]interface{}{},
			expected: map[string]interface{}{
				"ndstring": "",
				"string":   "stringd",
				"dstring":  "dstringd",
				"ndslice":  []string{},
				"slice":    []string{"sliced"},
				"dslice":   []string{"dsliced"},
				"ndmap":    map[string]string{},
				"map":      map[string]string{"k": "mapv"},
				"dmap":     map[string]string{"k": "dmapv"},
			},
		},
		{
			name: "merge",
			provider: map[string]interface{}{
				"dstring": "pstring",
				"dslice":  []interface{}{"pelem"},
				"dmap": map[string]interface{}{
					"k":        "kval",
					"override": "wrong",
				},
			},
			resource: map[string]interface{}{
				"dstring": "rstring",
				"dslice":  []interface{}{"relem"},
				"dmap": map[string]interface{}{
					"override": "right",
					"added":    "added",
				},
			},
			expected: map[string]interface{}{
				"dstring": "rstring",
				"dslice":  []string{"relem"},
				"dmap": map[string]string{
					"k":        "kval",
					"override": "right",
					"added":    "added",
				},
			},
		},
	}
	for i, tt := range ts {
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			pd, d := NewTestPD(ctx, t, ds, tt.provider)
			rd := NewTestRD(t, ds, tt.resource)
			cg := &ConfigGetter{Ds: ds, Rd: rd, Pd: pd}
			for k, v := range tt.expected {
				got, d0 := cg.Get(ctx, k)
				if d0.HasError() != tt.hasDiags {
					t.Errorf(
						"expected hasDiags: %#v, but got %#v",
						tt.hasDiags,
						d,
					)
				}
				if d0.HasError() {
					return
				}
				d = append(d, d0...)
				if !reflect.DeepEqual(got, v) {
					t.Errorf(
						"expected %#v, but got %#v",
						v,
						got,
					)
				}
			}
		})
	}
}

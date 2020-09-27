// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/leocp1/terraform-provider-packernix/src/pkg/patches"
	"github.com/leocp1/terraform-provider-packernix/src/pkg/provider"
)

func TestProvider(t *testing.T) {
	if err := provider.Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func ProviderFactory() (*schema.Provider, error) {
	return provider.Provider(), nil
}

func ProviderFactories() map[string]func() (*schema.Provider, error) {
	return map[string]func() (*schema.Provider, error){
		"packernix": ProviderFactory,
	}
}

// Helper function for reading config from a subdirectory of testdata
func ReadConfig(t *testing.T, path string, i interface{}) (c string) {
	t.Helper()

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf(err.Error())
		return
	}

	tmpl, err := template.ParseFiles(
		filepath.Join(wd, "testdata", path),
	)
	if err != nil {
		t.Fatalf(err.Error())
		return
	}

	bc := &bytes.Buffer{}
	err = tmpl.Execute(bc, i)
	if err != nil {
		t.Fatalf(err.Error())
		return
	}

	c = bc.String()
	return
}

// SkipFunc for TestSteps that require flakes
func FlakeSkipFunc(ctx context.Context, t *testing.T) func() (bool, error) {
	return func() (rv bool, err error) {
		rv = !patches.SupportsNixFlake(ctx)
		if rv {
			t.Log("Flake support not found")
		}
		return
	}
}

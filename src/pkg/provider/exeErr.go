// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package provider

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

func exeFail(
	d diag.Diagnostics,
	exe string,
	cmdSlice []string,
	err error,
) diag.Diagnostics {
	if err != nil {
		shcmd := &strings.Builder{}
		shcmd.WriteString("`")
		fmt.Fprintf(shcmd, "%#v", exe)
		for _, a := range cmdSlice {
			fmt.Fprintf(shcmd, " %#v", a)
		}
		fmt.Fprintf(shcmd, "` failed: %s", err.Error())
		d = append(d, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  shcmd.String(),
			Detail:   "Set TF_LOG=DEBUG to see command output",
		})
	}
	return d
}

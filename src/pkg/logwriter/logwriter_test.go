// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package logwriter_test

import (
	"bytes"
	"fmt"
	"log"
	"testing"

	. "github.com/leocp1/terraform-provider-packernix/src/pkg/logwriter"
)

type testCase struct {
	name   string
	prefix string
	in     string
	out    string
}

var tests = []testCase{
	{
		name:   "basic",
		prefix: "[INFO]:",
		in: `line 1
line 2
line 3`,
		out: `[INFO]:line 1
[INFO]:line 2
[INFO]:line 3
`,
	},
	{
		name:   "empty",
		prefix: "[INFO]:",
		in:     "",
		out:    "",
	},
	{
		name:   "empty lines",
		prefix: "[INFO]:",
		in: `


`,
		out: `[INFO]:
[INFO]:
[INFO]:
`,
	},
}

func TestPrefixWriter(t *testing.T) {
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%d: %s", i, tt.name), func(t *testing.T) {
			outb := &bytes.Buffer{}
			logger := log.New(outb, "", 0)
			pw := New(tt.prefix, logger)

			_, err := fmt.Fprint(pw, tt.in)
			if err != nil {
				t.Errorf("failed to write string: %w", err)
				t.FailNow()
			}
			err = pw.Close()
			if err != nil {
				t.Errorf("failed to close writer: %w", err)
			}
			got := outb.String()
			if got != tt.out {
				t.Errorf(
					"failed for %#v ... (got %#v)",
					tt,
					got,
				)
			}
		})
	}
}

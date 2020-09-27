// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Immediately fail if we attempt to write to a file that has been "locked".
//
// This is mainly intended as a sanity check against build_path being set to the
// same directory for different resources
package faillock

import (
	"fmt"
	"sync"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
)

type Faillock struct {
	l sync.Mutex
	s map[string]struct{}
}

// Create a new faillock
func New() *Faillock {
	return &Faillock{
		s: map[string]struct{}{},
	}
}

type Unlocker struct {
	fl *Faillock
	p  string
}

func (ul *Unlocker) Unlock() {
	ul.fl.l.Lock()
	defer ul.fl.l.Unlock()
	delete(ul.fl.s, ul.p)
}

// Attempt to lock a path. Returns a struct used to unlock the path if
// successful
func (fl *Faillock) TryLock(
	p string,
) (ul *Unlocker, d diag.Diagnostics) {
	fl.l.Lock()
	defer fl.l.Unlock()
	_, ok := fl.s[p]
	if ok {
		d = append(d, diag.Diagnostic{
			Severity: diag.Error,
			Summary: fmt.Sprintf(
				"Multiple resources attempted to write to %s",
				p,
			),
		})
	} else {
		fl.s[p] = struct{}{}
		ul = &Unlocker{
			fl: fl,
			p:  p,
		}
	}
	return
}

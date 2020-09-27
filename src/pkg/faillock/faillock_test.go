// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package faillock_test

import (
	"testing"

	. "github.com/leocp1/terraform-provider-packernix/src/pkg/faillock"
)

func TestFaillock(t *testing.T) {
	fl := New()
	ul, d := fl.TryLock("./faillock.go")
	if d.HasError() {
		t.Fatalf("Locking from new faillock failed")
	}
	_, d = fl.TryLock("./faillock.go")
	if !d.HasError() {
		t.Fatalf("Locking a locked file succeeded")
	}
	ul.Unlock()
	ul, d = fl.TryLock("./faillock.go")
	if d.HasError() {
		t.Fatalf("Locking a unlocked file failed")
	}
	ul.Unlock()
}

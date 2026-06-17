package config

import (
	"sync"
	"testing"
)

// TestLock serializes tests across the codebase that mutate the global
// config singleton. Exported so other packages' tests (serialize, client,
// docs, root forge) can synchronize via the same mutex when their tests
// touch GetGlobal(). (T0.1 / EC-1 from edge-case review).
var TestLock sync.Mutex

// ResetGlobalConfigForTest acquires TestLock, resets the global config to
// defaults, and schedules another reset + unlock on test exit.
//
// Call this at the top of any test (in any package) that mutates
// GetGlobal() or relies on a clean global state. The helper lives in a
// non-test file so other packages can import it; outside test binaries
// the `testing` import is harmless (only *testing.T is referenced).
func ResetGlobalConfigForTest(t *testing.T) {
	t.Helper()
	TestLock.Lock()
	GetGlobal().Reset()
	t.Cleanup(func() {
		GetGlobal().Reset()
		TestLock.Unlock()
	})
}

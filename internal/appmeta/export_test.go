package appmeta

import "sync"

var setVersionMu sync.Mutex

// SetVersion sets the version value for tests and returns a function that restores the original value.
func SetVersion(v string) func() {
	setVersionMu.Lock()

	orig := version
	version = v

	return func() { version = orig; setVersionMu.Unlock() }
}

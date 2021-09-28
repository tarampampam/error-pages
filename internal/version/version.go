// Package version is used as a place, where application version defined.
package version

import "strings"

// version value will be set during compilation.
var version = "v0.0.0@undefined"

// Version returns version value (without `v` prefix).
func Version() string {
	v := strings.TrimSpace(version)

	if len(v) > 1 && ((v[0] == 'v' || v[0] == 'V') && (v[1] >= '0' && v[1] <= '9')) {
		return v[1:]
	}

	return v
}

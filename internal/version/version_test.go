package version

import (
	"testing"
)

func TestVersion(t *testing.T) {
	t.Parallel()

	for give, want := range map[string]string{
		// without changes
		"vvv":     "vvv",
		"victory": "victory",
		"voodoo":  "voodoo",
		"foo":     "foo",
		"0.0.0":   "0.0.0",
		"v":       "v",
		"V":       "V",

		// "v" prefix removal
		"v0.0.0": "0.0.0",
		"V0.0.0": "0.0.0",
		"v1":     "1",
		"V1":     "1",

		// with spaces
		" 0.0.0":  "0.0.0",
		"v0.0.0 ": "0.0.0",
		" V0.0.0": "0.0.0",
		"v1 ":     "1",
		" V1":     "1",
		"v ":      "v",
	} {
		version = give

		if v := Version(); v != want {
			t.Errorf("want: %s, got: %s", want, v)
		}
	}
}

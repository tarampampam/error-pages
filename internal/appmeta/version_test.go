package appmeta_test

import (
	"fmt"
	"testing"

	"gh.tarampamp.am/error-pages/v4/internal/appmeta"
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
		t.Run(fmt.Sprintf("%s -> %s", give, want), func(t *testing.T) {
			t.Parallel()
			t.Cleanup(appmeta.SetVersion(give))

			if v := appmeta.Version(); v != want {
				t.Errorf("want: %s, got: %s", want, v)
			}
		})
	}
}

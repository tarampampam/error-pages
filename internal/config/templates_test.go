package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplates_Common(t *testing.T) {
	t.Parallel()

	var tpl = make(templates)

	t.Run("initial state", func(t *testing.T) {
		assert.Empty(t, tpl.Names())
		assert.False(t, tpl.Has("test"))

		var got, ok = tpl.Get("test")

		assert.Empty(t, got)
		assert.False(t, ok)
	})

	t.Run("add a template from variable", func(t *testing.T) {
		const testContent = "content"

		assert.NoError(t, tpl.Add("test", testContent))
		assert.True(t, tpl.Has("test"))

		var got, ok = tpl.Get("test")

		assert.Equal(t, got, testContent)
		assert.True(t, ok)
		assert.Equal(t, []string{"test"}, tpl.Names())
		assert.False(t, tpl.Has("_test99"))

		assert.NoError(t, tpl.Add("_test99", ""))
		assert.NoError(t, tpl.Add("_test11", ""))

		assert.Equal(t, []string{"_test11", "_test99", "test"}, tpl.Names()) // sorted
		assert.True(t, tpl.Has("_test99"))
	})

	t.Run("adding template without a name should fail", func(t *testing.T) {
		assert.ErrorContains(t, tpl.Add("", "content"), "template name cannot be empty")
	})
}

func TestTemplates_AddFromFile(t *testing.T) {
	t.Parallel()

	for name, _tt := range map[string]struct {
		givePath string
		giveName func() []string

		wantError       string
		wantThisName    string
		wantThisContent string
	}{
		"dotfile": {
			givePath:     "./testdata/.dotfile",
			wantThisName: ".dotfile",
		},
		"dotfile with extension": {
			givePath:     "./testdata/.dotfile_with.ext",
			wantThisName: ".dotfile_with",
		},
		"empty file": {
			givePath:     "./testdata/empty.html",
			wantThisName: "empty",
		},
		"file with multiple dots but without a name": {
			givePath:     "./testdata/file.with.multiple.dots",
			wantThisName: "file.with.multiple",
		},
		"name with spaces": {
			givePath:     "./testdata/name with spaces.txt",
			wantThisName: "name with spaces",
		},
		"with content and a name": {
			givePath:        "./testdata/with-content.htm",
			giveName:        func() []string { return []string{"test name"} },
			wantThisName:    "test name",
			wantThisContent: "<!DOCTYPE html><html lang=\"en\"></html>\n",
		},
		"with content but without a name": {
			givePath:        "./testdata/with-content.htm",
			wantThisName:    "with-content",
			wantThisContent: "<!DOCTYPE html><html lang=\"en\"></html>\n",
		},
		"filename with no extension": {
			givePath:     "./testdata/without_extension",
			wantThisName: "without_extension",
		},

		"file not found": {
			givePath:  "./testdata/not-found",
			wantError: "file ./testdata/not-found not found",
		},
		"directory": {
			givePath:  "./testdata",
			wantError: "./testdata is not a file",
		},
	} {
		var tt = _tt

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var (
				tpl      = make(templates)
				giveName []string
			)

			if tt.giveName != nil {
				giveName = tt.giveName()
			}

			var addedName, err = tpl.AddFromFile(tt.givePath, giveName...)

			if tt.wantError == "" {
				assert.NoError(t, err)
				assert.True(t, tpl.Has(tt.wantThisName))
				assert.Equal(t, addedName, tt.wantThisName)

				var content, _ = tpl.Get(tt.wantThisName)

				assert.Equal(t, content, tt.wantThisContent)
			} else {
				assert.ErrorContains(t, err, tt.wantError)

				assert.False(t, tpl.Has(tt.wantThisName))
			}
		})
	}
}

package templates_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/error-pages/templates"
)

func TestBuiltIn(t *testing.T) {
	t.Parallel()

	var content = templates.BuiltIn()

	assert.True(t, len(content) > 0)

	for name, data := range content {
		assert.Regexp(t, `^[a-z0-9_\.-]+$`, name)
		assert.NotEmpty(t, data)
	}
}

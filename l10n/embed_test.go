package l10n_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/error-pages/l10n"
)

func TestL10n(t *testing.T) {
	assert.NotEmpty(t, l10n.L10n())
	assert.Contains(t, l10n.L10n(), "data-l10n")
}

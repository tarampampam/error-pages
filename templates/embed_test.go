package templates_test

import (
	"strings"
	"testing"

	"gh.tarampamp.am/error-pages/v4/templates"
)

func TestBuiltIn(t *testing.T) {
	t.Parallel()

	t.Run("not empty", func(t *testing.T) {
		t.Parallel()

		assertFalse(t, len(templates.BuiltIn()) == 0)
	})

	t.Run("contains all template names", func(t *testing.T) {
		t.Parallel()

		m := templates.BuiltIn()

		for _, name := range []string{
			templates.TemplateNameAppDown,
			templates.TemplateNameCats,
			templates.TemplateNameConnection,
			templates.TemplateNameGhost,
			templates.TemplateNameHackerTerminal,
			templates.TemplateNameL7,
			templates.TemplateNameLostInSpace,
			templates.TemplateNameNoise,
			templates.TemplateNameOrient,
			templates.TemplateNameShuffle,
			templates.TemplateNameWin98,
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				_, ok := m[name]
				assertTrue(t, ok)
			})
		}
	})

	t.Run("all templates have content", func(t *testing.T) {
		t.Parallel()

		for name, content := range templates.BuiltIn() {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				assertFalse(t, content == "")

				// check that the content looks like HTML
				assertContains(t, content, "<html")
				assertContains(t, content, "</html>")
			})
		}
	})

	t.Run("returns independent map", func(t *testing.T) {
		t.Parallel()

		m := templates.BuiltIn()
		delete(m, templates.TemplateNameAppDown)

		_, ok := templates.BuiltIn()[templates.TemplateNameAppDown]
		assertTrue(t, ok)
	})
}

// --------------------------------------------------------------------------------------------------------------------

// assertTrue is a helper function that asserts that the given condition is true.
func assertTrue(t *testing.T, condition bool) {
	t.Helper()

	if !condition {
		t.Errorf("expected condition to be true, but it was false")
	}
}

// assertFalse is a helper function that asserts that the given condition is false.
func assertFalse(t *testing.T, condition bool) {
	t.Helper()

	if condition {
		t.Errorf("expected condition to be false, but it was true")
	}
}

// assertContains is a helper function that asserts that the given string contains the specified substring.
func assertContains(t *testing.T, s, substr string) {
	t.Helper()

	if !strings.Contains(s, substr) {
		t.Errorf("expected string to contain substring, but it did not: %q does not contain %q", s, substr)
	}
}

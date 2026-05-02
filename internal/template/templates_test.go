package tpl_test

import (
	"testing"
	"time"

	"gh.tarampamp.am/error-pages/v4/internal/formats"
	tpl "gh.tarampamp.am/error-pages/v4/internal/template"
	"gh.tarampamp.am/error-pages/v4/internal/testutil/assert"
	"gh.tarampamp.am/error-pages/v4/templates"
)

func TestNewTemplates(t *testing.T) {
	t.Parallel()

	t.Run("defaults load without error", func(t *testing.T) {
		t.Parallel()

		_, err := tpl.NewTemplates()
		assert.NoError(t, err)
	})

	t.Run("option errors", func(t *testing.T) {
		t.Parallel()

		for name, tt := range map[string]struct {
			giveOpt       tpl.TemplatesOption
			wantErrSubstr string
		}{
			"unknown HTML template name": {
				giveOpt:       tpl.WithHTMLTemplateName("does-not-exist"),
				wantErrSubstr: "not found among built-in templates",
			},
			"invalid custom HTML template": {
				giveOpt:       tpl.WithCustomHTMLTemplate("{{.Invalid"),
				wantErrSubstr: "custom HTML template parsing",
			},
			"invalid custom JSON template": {
				giveOpt:       tpl.WithCustomJSONTemplate("{{.Invalid"),
				wantErrSubstr: "custom JSON template parsing",
			},
			"invalid custom XML template": {
				giveOpt:       tpl.WithCustomXMLTemplate("{{.Invalid"),
				wantErrSubstr: "custom XML template parsing",
			},
			"invalid custom plain-text template": {
				giveOpt:       tpl.WithCustomPlainTextTemplate("{{.Invalid"),
				wantErrSubstr: "custom plain text template parsing",
			},
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				_, err := tpl.NewTemplates(tt.giveOpt)
				assert.ErrorContains(t, err, tt.wantErrSubstr)
			})
		}
	})

	t.Run("WithHTMLTemplateName accepts every built-in name", func(t *testing.T) {
		t.Parallel()

		for name := range templates.BuiltInHTML() {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				_, err := tpl.NewTemplates(tpl.WithHTMLTemplateName(name))
				assert.NoError(t, err)
			})
		}
	})

	t.Run("RotationModeRandomOnStartup keeps the same template across all Get calls", func(t *testing.T) {
		t.Parallel()

		ts, err := tpl.NewTemplates(tpl.WithRotationMode(tpl.RotationModeRandomOnStartup))
		assert.NoError(t, err)

		first, firstErr := ts.Get(formats.HTMLFormat)
		assert.NoError(t, firstErr)
		assert.True(t, first != nil)

		for range 20 {
			got, getErr := ts.Get(formats.HTMLFormat)
			assert.NoError(t, getErr)
			assert.Equal(t, first, got) // pointer identity: same template object every time
		}
	})
}

func TestTemplates_Get(t *testing.T) {
	t.Parallel()

	t.Run("HTML/disabled rotation returns the same template on every call", func(t *testing.T) {
		t.Parallel()

		ts, err := tpl.NewTemplates(tpl.WithRotationMode(tpl.RotationModeDisabled))
		assert.NoError(t, err)

		first, firstErr := ts.Get(formats.HTMLFormat)
		assert.NoError(t, firstErr)
		assert.True(t, first != nil)

		for range 10 {
			got, getErr := ts.Get(formats.HTMLFormat)
			assert.NoError(t, getErr)
			assert.Equal(t, first, got)
		}
	})

	t.Run("HTML/custom template overrides all rotation modes", func(t *testing.T) {
		t.Parallel()

		for _, mode := range []tpl.RotationMode{
			tpl.RotationModeDisabled,
			tpl.RotationModeRandomOnStartup,
			tpl.RotationModeRandomOnEachRequest,
			tpl.RotationModeRandomHourly,
			tpl.RotationModeRandomDaily,
		} {
			t.Run(string(mode), func(t *testing.T) {
				t.Parallel()

				ts, err := tpl.NewTemplates(
					tpl.WithCustomHTMLTemplate(`custom: {{code}}`),
					tpl.WithRotationMode(mode),
				)
				assert.NoError(t, err)

				first, firstErr := ts.Get(formats.HTMLFormat)
				assert.NoError(t, firstErr)
				assert.True(t, first != nil)

				for range 5 {
					got, getErr := ts.Get(formats.HTMLFormat)
					assert.NoError(t, getErr)
					assert.Equal(t, first, got) // same custom template pointer every time regardless of mode
				}
			})
		}
	})

	t.Run("HTML/random-on-each-request returns different templates over many calls", func(t *testing.T) {
		t.Parallel()

		ts, err := tpl.NewTemplates(tpl.WithRotationMode(tpl.RotationModeRandomOnEachRequest))
		assert.NoError(t, err)

		seen := make(map[*tpl.Template]bool)

		for range 50 {
			got, getErr := ts.Get(formats.HTMLFormat)
			assert.NoError(t, getErr)
			assert.True(t, got != nil)
			seen[got] = true
		}

		// with 11 built-in templates, P(all 50 calls hit the same one) = (1/11)^49 ≈ 10^-52
		assert.True(t, len(seen) > 1)
	})

	t.Run("HTML/random-hourly", func(t *testing.T) {
		t.Parallel()

		var current time.Time

		ts, err := tpl.NewTemplates(
			tpl.WithRotationMode(tpl.RotationModeRandomHourly),
			tpl.WithClock(func() time.Time { return current }),
		)
		assert.NoError(t, err)

		// period 1: 2024-01-15 10:xx UTC
		current = time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

		period1, err := ts.Get(formats.HTMLFormat)
		assert.NoError(t, err)
		assert.True(t, period1 != nil)

		// calls within the same hour must return the exact same pointer
		for m := 31; m < 60; m++ {
			current = time.Date(2024, 1, 15, 10, m, 0, 0, time.UTC)
			got, getErr := ts.Get(formats.HTMLFormat)
			assert.NoError(t, getErr)
			assert.Equal(t, period1, got)
		}

		// period 2: 2024-01-15 11:xx UTC - crosses the hour boundary
		current = time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)

		period2, err := ts.Get(formats.HTMLFormat)
		assert.NoError(t, err)
		assert.True(t, period2 != nil)

		for m := 1; m < 10; m++ {
			current = time.Date(2024, 1, 15, 11, m, 0, 0, time.UTC)
			got, getErr := ts.Get(formats.HTMLFormat)
			assert.NoError(t, getErr)
			assert.Equal(t, period2, got)
		}

		// period 3: 2024-01-16 11:xx UTC - same clock-hour, different date
		current = time.Date(2024, 1, 16, 11, 0, 0, 0, time.UTC)

		period3, err := ts.Get(formats.HTMLFormat)
		assert.NoError(t, err)
		assert.True(t, period3 != nil)

		for m := 1; m < 5; m++ {
			current = time.Date(2024, 1, 16, 11, m, 0, 0, time.UTC)
			got, getErr := ts.Get(formats.HTMLFormat)
			assert.NoError(t, getErr)
			assert.Equal(t, period3, got)
		}
	})

	t.Run("HTML/random-daily", func(t *testing.T) {
		t.Parallel()

		var current time.Time

		ts, err := tpl.NewTemplates(
			tpl.WithRotationMode(tpl.RotationModeRandomDaily),
			tpl.WithClock(func() time.Time { return current }),
		)
		assert.NoError(t, err)

		// period 1: 2024-01-15
		current = time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

		period1, err := ts.Get(formats.HTMLFormat)
		assert.NoError(t, err)
		assert.True(t, period1 != nil)

		// calls within the same day must return the exact same pointer
		for h := 11; h < 20; h++ {
			current = time.Date(2024, 1, 15, h, 0, 0, 0, time.UTC)
			got, getErr := ts.Get(formats.HTMLFormat)
			assert.NoError(t, getErr)
			assert.Equal(t, period1, got)
		}

		// period 2: 2024-01-16 - crosses the day boundary
		current = time.Date(2024, 1, 16, 10, 0, 0, 0, time.UTC)

		period2, err := ts.Get(formats.HTMLFormat)
		assert.NoError(t, err)
		assert.True(t, period2 != nil)

		for h := 11; h < 15; h++ {
			current = time.Date(2024, 1, 16, h, 0, 0, 0, time.UTC)
			got, getErr := ts.Get(formats.HTMLFormat)
			assert.NoError(t, getErr)
			assert.Equal(t, period2, got)
		}

		// period 3: 2024-02-16 - same day-of-month, different month
		current = time.Date(2024, 2, 16, 10, 0, 0, 0, time.UTC)

		period3, err := ts.Get(formats.HTMLFormat)
		assert.NoError(t, err)
		assert.True(t, period3 != nil)

		for h := 11; h < 15; h++ {
			current = time.Date(2024, 2, 16, h, 0, 0, 0, time.UTC)
			got, getErr := ts.Get(formats.HTMLFormat)
			assert.NoError(t, getErr)
			assert.Equal(t, period3, got)
		}
	})

	t.Run("HTML/unknown rotation mode returns error", func(t *testing.T) {
		t.Parallel()

		ts, err := tpl.NewTemplates(tpl.WithRotationMode("unknown-rotation-mode"))
		assert.NoError(t, err) // construction succeeds; error surfaces only on Get

		got, getErr := ts.Get(formats.HTMLFormat)
		assert.ErrorContains(t, getErr, "unknown HTML rotation mode")
		assert.True(t, got == nil)
	})

	t.Run("non-HTML formats", func(t *testing.T) {
		t.Parallel()

		for name, tt := range map[string]struct {
			giveFormat formats.Format
			giveOpts   []tpl.TemplatesOption
		}{
			"json/built-in":       {giveFormat: formats.JSONFormat},
			"xml/built-in":        {giveFormat: formats.XMLFormat},
			"plain-text/built-in": {giveFormat: formats.PlainTextFormat},
			"json/custom":         {giveFormat: formats.JSONFormat, giveOpts: []tpl.TemplatesOption{tpl.WithCustomJSONTemplate(`{"c": {{code}}}`)}},
			"xml/custom":          {giveFormat: formats.XMLFormat, giveOpts: []tpl.TemplatesOption{tpl.WithCustomXMLTemplate(`<c>{{code}}</c>`)}},
			"plain-text/custom":   {giveFormat: formats.PlainTextFormat, giveOpts: []tpl.TemplatesOption{tpl.WithCustomPlainTextTemplate(`{{code}}`)}},
		} {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				ts, err := tpl.NewTemplates(tt.giveOpts...)
				assert.NoError(t, err)

				first, firstErr := ts.Get(tt.giveFormat)
				assert.NoError(t, firstErr)
				assert.True(t, first != nil)

				// same pointer on every call - template objects are not recreated per request
				for range 5 {
					got, getErr := ts.Get(tt.giveFormat)
					assert.NoError(t, getErr)
					assert.Equal(t, first, got)
				}
			})
		}
	})

	t.Run("unknown format returns ErrFormatIsNotSupported", func(t *testing.T) {
		t.Parallel()

		ts, err := tpl.NewTemplates()
		assert.NoError(t, err)

		got, getErr := ts.Get(formats.Format(255))
		assert.ErrorIs(t, getErr, tpl.ErrFormatIsNotSupported)
		assert.True(t, got == nil)
	})
}

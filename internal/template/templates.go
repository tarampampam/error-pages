package tpl

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"sync/atomic"
	"time"

	"gh.tarampamp.am/error-pages/v4/internal/formats"
	"gh.tarampamp.am/error-pages/v4/templates"
)

// RotationMode represents the rotation mode for templates.
type RotationMode string

const (
	RotationModeDisabled            RotationMode = "disabled"               // do not rotate templates, default
	RotationModeRandomOnStartup     RotationMode = "random-on-startup"      // pick a random template on startup
	RotationModeRandomOnEachRequest RotationMode = "random-on-each-request" // pick a random template on each request
	RotationModeRandomHourly        RotationMode = "random-hourly"          // once an hour switch to a random template
	RotationModeRandomDaily         RotationMode = "random-daily"           // once a day switch to a random template
)

// Templates contains the HTML/JSON/XML/etc templates for the app.
type Templates struct {
	// clockFn is injectable so that callers can control the time source; primarily useful for testing
	// deterministic rotation behavior without sleeping
	clockFn func() time.Time

	html struct {
		builtIn struct {
			m     map[string]*Template
			names []string // to avoid map iteration on each request for random selection
		}

		custom          *Template
		rotationMode    RotationMode
		useTemplateName string

		templateChangedAt  atomic.Pointer[time.Time]
		pickedTemplateName atomic.Pointer[string]
	}
	json      *Template
	xml       *Template
	plainText *Template
}

// TemplatesOption is a functional option for configuring a [Templates] instance via [NewTemplates].
type TemplatesOption func(*Templates) error

// WithCustomHTMLTemplate sets a custom HTML template, replacing all built-in HTML themes.
func WithCustomHTMLTemplate(src string) TemplatesOption {
	src = strings.TrimSpace(src)

	if src == "" {
		return func(t *Templates) error { return nil }
	}

	return func(t *Templates) error {
		tpl, err := New(src + "\n")
		if err != nil {
			return fmt.Errorf("custom HTML template parsing: %w", err)
		}

		t.html.custom = tpl

		return nil
	}
}

// WithHTMLTemplateName selects one of the built-in HTML templates by name. Returns an error if name does not
// match any loaded built-in template.
func WithHTMLTemplateName(name string) TemplatesOption {
	return func(t *Templates) error {
		if t.html.builtIn.m == nil {
			return fmt.Errorf("cannot set HTML template name %q: built-in templates not loaded yet", name)
		}

		if _, ok := t.html.builtIn.m[name]; !ok {
			return fmt.Errorf("HTML template with name %q not found among built-in templates", name)
		}

		t.html.useTemplateName = name

		return nil
	}
}

// WithCustomJSONTemplate sets a custom JSON response template, overriding the built-in default.
func WithCustomJSONTemplate(src string) TemplatesOption {
	src = strings.TrimSpace(src)

	if src == "" {
		return func(t *Templates) error { return nil }
	}

	return func(t *Templates) error {
		tpl, err := New(src + "\n")
		if err != nil {
			return fmt.Errorf("custom JSON template parsing: %w", err)
		}

		t.json = tpl

		return nil
	}
}

// WithCustomXMLTemplate sets a custom XML response template, overriding the built-in default.
func WithCustomXMLTemplate(src string) TemplatesOption {
	src = strings.TrimSpace(src)

	if src == "" {
		return func(t *Templates) error { return nil }
	}

	return func(t *Templates) error {
		tpl, err := New(src + "\n")
		if err != nil {
			return fmt.Errorf("custom XML template parsing: %w", err)
		}

		t.xml = tpl

		return nil
	}
}

// WithCustomPlainTextTemplate sets a custom plain-text response template, overriding the built-in default.
func WithCustomPlainTextTemplate(src string) TemplatesOption {
	src = strings.TrimSpace(src)

	if src == "" {
		return func(t *Templates) error { return nil }
	}

	return func(t *Templates) error {
		tpl, err := New(src + "\n")
		if err != nil {
			return fmt.Errorf("custom plain text template parsing: %w", err)
		}

		t.plainText = tpl

		return nil
	}
}

// WithRotationMode sets the HTML template rotation strategy. Has no effect when a custom HTML template is
// configured via [WithCustomHTMLTemplate].
func WithRotationMode(m RotationMode) TemplatesOption {
	return func(t *Templates) error {
		t.html.rotationMode = m

		return nil
	}
}

// WithClock overrides the time source used for [RotationModeRandomHourly] and [RotationModeRandomDaily]
// rotation modes. fn is called on each [Templates.Get] invocation to obtain the current time. Primarily
// useful for testing deterministic rotation behavior.
func WithClock(fn func() time.Time) TemplatesOption {
	return func(t *Templates) error {
		t.clockFn = fn

		return nil
	}
}

// NewTemplates creates a new Templates instance by parsing the embedded templates.
func NewTemplates(opts ...TemplatesOption) (*Templates, error) {
	t := Templates{
		clockFn: time.Now, // default clock function
	}

	builtIn := templates.BuiltInHTML()
	t.html.builtIn.m = make(map[string]*Template, len(builtIn))
	t.html.builtIn.names = make([]string, 0, len(builtIn))

	for name, src := range builtIn {
		tpl, err := New(src)
		if err != nil {
			return nil, fmt.Errorf("built-in HTML template %q parsing: %w", name, err)
		}

		t.html.builtIn.m[name] = tpl
		t.html.builtIn.names = append(t.html.builtIn.names, name)
	}

	slices.Sort(t.html.builtIn.names) // to ensure consistent order of template names

	if len(t.html.builtIn.names) > 0 {
		t.html.useTemplateName = t.html.builtIn.names[0] // default to the first template
	}

	for _, opt := range opts {
		if err := opt(&t); err != nil {
			return nil, err
		}
	}

	if t.json == nil {
		v, err := New(templates.JSON)
		if err != nil {
			return nil, fmt.Errorf("built-in JSON template parsing: %w", err)
		}

		t.json = v
	}

	if t.xml == nil {
		v, err := New(templates.XML)
		if err != nil {
			return nil, fmt.Errorf("built-in XML template parsing: %w", err)
		}

		t.xml = v
	}

	if t.plainText == nil {
		v, err := New(templates.PlaintText)
		if err != nil {
			return nil, fmt.Errorf("built-in plain text template parsing: %w", err)
		}

		t.plainText = v
	}

	if t.html.rotationMode == RotationModeRandomOnStartup {
		t.html.useTemplateName = t.getRandomBuiltInTemplateName()
	}

	return &t, nil
}

// getRandomBuiltInTemplateName returns a random built-in template name.
// It returns an empty string if there are no built-in templates available.
func (t *Templates) getRandomBuiltInTemplateName() string {
	// to avoid panic in case of no built-in templates (should not happen, but just in case)
	if len(t.html.builtIn.names) == 0 {
		return ""
	}

	return t.html.builtIn.names[rand.IntN(len(t.html.builtIn.names))] //nolint:gosec
}

// ErrNoHTMLTpl is returned by [Templates.Get] for [formats.HTMLFormat] when no built-in templates are loaded
// and no custom template has been configured via [WithCustomHTMLTemplate].
var ErrNoHTMLTpl = errors.New("no HTML template available: no built-in templates loaded and no custom template set")

// ErrFormatIsNotSupported is returned by [Templates.Get] when the requested [formats.Format] is not recognized.
var ErrFormatIsNotSupported = errors.New("format is not supported")

// Get returns the [Template] for the given format. For [formats.HTMLFormat], the selected template depends on
// the configured [RotationMode]:
//   - [RotationModeDisabled] and [RotationModeRandomOnStartup]: returns the fixed template (set at construction).
//   - [RotationModeRandomOnEachRequest]: picks a random built-in template on every call.
//   - [RotationModeRandomHourly]: rotates to a new random template once per UTC hour.
//   - [RotationModeRandomDaily]: rotates to a new random template once per UTC day.
//
// A custom HTML template set via [WithCustomHTMLTemplate] always takes precedence over rotation.
func (t *Templates) Get(format formats.Format) (*Template, error) {
	switch format {
	case formats.HTMLFormat:
		if t.html.custom != nil {
			return t.html.custom, nil
		}

		switch t.html.rotationMode {
		case RotationModeDisabled, RotationModeRandomOnStartup:
			if t.html.useTemplateName == "" {
				return nil, ErrNoHTMLTpl
			}

			return t.html.builtIn.m[t.html.useTemplateName], nil
		case RotationModeRandomOnEachRequest:
			randomName := t.getRandomBuiltInTemplateName()
			if randomName == "" {
				return nil, ErrNoHTMLTpl
			}

			return t.html.builtIn.m[randomName], nil
		case RotationModeRandomHourly, RotationModeRandomDaily:
			if len(t.html.builtIn.names) == 0 {
				return nil, ErrNoHTMLTpl
			}

			now := t.clockFn().UTC()
			lastChangedAt := t.html.templateChangedAt.Load()

			if lastChangedAt == nil { // the template was not changed yet (first request)
				randomName := t.getRandomBuiltInTemplateName()
				t.html.templateChangedAt.Store(&now)
				t.html.pickedTemplateName.Store(&randomName)

				return t.html.builtIn.m[randomName], nil
			}

			const hoursInDay = 24

			shouldRotate := (t.html.rotationMode == RotationModeRandomHourly &&
				lastChangedAt.Truncate(time.Hour) != now.Truncate(time.Hour)) ||
				(t.html.rotationMode == RotationModeRandomDaily &&
					lastChangedAt.Truncate(hoursInDay*time.Hour) != now.Truncate(hoursInDay*time.Hour))

			if shouldRotate {
				randomName := t.getRandomBuiltInTemplateName()
				t.html.templateChangedAt.Store(&now)
				t.html.pickedTemplateName.Store(&randomName)

				return t.html.builtIn.m[randomName], nil
			}

			if lastUsed := t.html.pickedTemplateName.Load(); lastUsed != nil {
				return t.html.builtIn.m[*lastUsed], nil
			}

			randomName := t.getRandomBuiltInTemplateName()
			t.html.templateChangedAt.Store(&now)
			t.html.pickedTemplateName.Store(&randomName)

			return t.html.builtIn.m[randomName], nil
		default:
			return nil, fmt.Errorf("unknown HTML rotation mode %q", t.html.rotationMode)
		}
	case formats.JSONFormat:
		return t.json, nil
	case formats.XMLFormat:
		return t.xml, nil
	case formats.PlainTextFormat:
		return t.plainText, nil
	}

	return nil, ErrFormatIsNotSupported
}

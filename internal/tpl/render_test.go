package tpl_test

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarampampam/error-pages/internal/tpl"
)

func Test_Render(t *testing.T) {
	renderer := tpl.NewTemplateRenderer()
	defer func() { _ = renderer.Close() }()

	for name, tt := range map[string]struct {
		giveContent string
		giveProps   tpl.Properties
		wantContent string
		wantError   bool
	}{
		"common case": {
			giveContent: "{{code}}: {{ message }} {{description}}",
			giveProps:   tpl.Properties{Code: "404", Message: "Not found", Description: "Blah"},
			wantContent: "404: Not found Blah",
		},
		"html markup": {
			giveContent: "<!-- comment --><html><body>{{code}}: {{ message }} {{description}}</body></html>",
			giveProps:   tpl.Properties{Code: "201", Message: "lorem ipsum"},
			wantContent: "<!-- comment --><html><body>201: lorem ipsum </body></html>",
		},
		"with line breakers": {
			giveContent: "\t {{code}}: {{ message }} {{description}}\n",
			giveProps:   tpl.Properties{},
			wantContent: "\t :  \n",
		},
		"golang template": {
			giveContent: "\t {{code}} {{ .Code }}{{ if .Message }} Yeah {{end}}",
			giveProps:   tpl.Properties{Code: "201", Message: "lorem ipsum"},
			wantContent: "\t 201 201 Yeah ",
		},
		"wrong golang template": {
			giveContent: "{{ if foo() }} Test {{ end }}",
			giveProps:   tpl.Properties{},
			wantError:   true,
		},

		"json common case": {
			giveContent: `{"code": {{code | json}}, "message": {"here":[ {{ message | json }} ]}, "desc": "{{description}}"}`,
			giveProps:   tpl.Properties{Code: `404'"{`, Message: "Not found\t\r\n"},
			wantContent: `{"code": "404'\"{", "message": {"here":[ "Not found\t\r\n" ]}, "desc": ""}`,
		},
		"json golang template": {
			giveContent: `{"code": "{{code}}", "message": {"here":[ "{{ if .Message }} Yeah {{end}}" ]}}`,
			giveProps:   tpl.Properties{Code: "201", Message: "lorem ipsum"},
			wantContent: `{"code": "201", "message": {"here":[ " Yeah " ]}}`,
		},
	} {
		t.Run(name, func(t *testing.T) {
			content, err := renderer.Render([]byte(tt.giveContent), tt.giveProps)

			if tt.wantError == true {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantContent, string(content))
			}
		})
	}
}

func TestTemplateRenderer_Render_Concurrent(t *testing.T) {
	renderer := tpl.NewTemplateRenderer()

	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			props := tpl.Properties{
				Code:        strconv.Itoa(rand.Intn(599-300+1) + 300), //nolint:gosec
				Message:     "Not found",
				Description: "Blah",
			}

			content, err := renderer.Render([]byte("{{code}}: {{ message }} {{description}}"), props)

			assert.NoError(t, err)
			assert.NotEmpty(t, content)
		}()
	}

	wg.Wait()

	assert.NoError(t, renderer.Close())
	assert.EqualError(t, renderer.Close(), tpl.ErrClosed.Error())
}

func BenchmarkRenderHTML(b *testing.B) {
	b.ReportAllocs()

	renderer := tpl.NewTemplateRenderer()
	defer func() { _ = renderer.Close() }()

	for i := 0; i < b.N; i++ {
		_, _ = renderer.Render(
			[]byte("{{code}}: {{ message }} {{description}}"),
			tpl.Properties{Code: "404", Message: "Not found", Description: "Blah"},
		)
	}
}

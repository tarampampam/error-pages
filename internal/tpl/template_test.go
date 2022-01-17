package tpl_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarampampam/error-pages/internal/tpl"
)

func Test_Render(t *testing.T) {
	assert.Equal(t,
		"404: Not found Blah",
		string(tpl.Render(
			[]byte("{{code}}: {{ message }} {{description}}"),
			tpl.Properties{Code: "404", Message: "Not found", Description: "Blah"},
		)),
	)

	assert.Equal(t,
		"201: lorem ipsum ",
		string(tpl.Render(
			[]byte("{{code}}: {{ message }} {{description}}"),
			tpl.Properties{Code: "201", Message: "lorem ipsum"},
		)),
	)

	assert.Equal(t,
		"\t :  \n",
		string(tpl.Render(
			[]byte("\t {{code}}: {{ message }} {{description}}\n"),
			tpl.Properties{},
		)),
	)
}

func BenchmarkRender(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		tpl.Render(
			[]byte("{{code}}: {{ message }} {{description}}"),
			tpl.Properties{Code: "404", Message: "Not found", Description: "Blah"},
		)
	}
}

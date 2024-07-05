package template_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/error-pages/internal/template"
)

func TestMiniHTML(t *testing.T) {
	t.Parallel()

	var wg sync.WaitGroup

	for range 100 { // race condition provocation
		wg.Add(1)

		go func() {
			defer wg.Done()

			for give, want := range map[string]string{
				"": "",
				`<!-- Simple HTML page -->
<!DOCTYPE html>
<html>
<head>
	<title>Test</title>
</head>
<body>
	<h1 align="center">Test</h1>
</body>
</html>`: `<!doctype html><html><head><title>Test</title></head><body><h1 align="center">Test</h1></body></html>`,
				`<!-- css styles -->
<html>
<head>
	<style>
		.foo:hover {
			color: #f0a; /* comment */
		}
	</style>
</head>
<body>
	<p style="color: red" class="bar">Text</p>
</body>
</html>`: `<html><head><style>.foo:hover{color:#f0a}</style></head><body><p style="color:red" class="bar">Text</p></body></html>`,
				`<!-- svg -->
<svg xmlns="http://www.w3.org/2000/svg">
	<g>
		<circle cx="50" cy="50" r="40" stroke="black" stroke-width="3" fill="red" />
	</g>
</svg>`: `<svg><g><circle cx="50" cy="50" r="40" stroke="#000" stroke-width="3" fill="red"/></g></svg>`,
				`<!-- js -->
<html>
<body>
<script>
	// comment
	console.log('Hello, World!');

	let foo = 1;
	foo++;
</script>
</body>
</html>`: `<html><body><script>console.log("Hello, World!");let foo=1;foo++</script></body></html>`,
				`<!-- js module not changed -->
<html>
<body>
<script type="module">
	// comment
	console.log('Hello, World!');

	let foo = 1;
	foo++;
</script>
</body>
</html>`: `<html><body><script type="module">
	// comment
	console.log('Hello, World!');

	let foo = 1;
	foo++;
</script></body></html>`,
			} {
				var got, err = template.MiniHTML(give)

				assert.NoError(t, err)
				assert.Equal(t, want, got)
			}
		}()
	}

	wg.Wait()
}

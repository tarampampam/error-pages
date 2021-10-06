package tpl_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/tarampampam/error-pages/internal/tpl"

	"github.com/stretchr/testify/assert"
)

func TestErrorPages_GetPage(t *testing.T) {
	e := tpl.NewErrorPages()

	e.AddTemplate("foo", []byte("{{code}}: {{ message }} {{description}}"))
	e.AddPage("200", "ok", "all is ok")
	e.AddTemplate("bar", []byte("{{ code }} _ {{message}} ({{ description }})"))
	e.AddPage("201", "lorem", "ipsum")

	content, err := e.GetPage("foo", "200")
	assert.NoError(t, err)
	assert.Equal(t, "200: ok all is ok", string(content))

	content, err = e.GetPage("foo", "201")
	assert.NoError(t, err)
	assert.Equal(t, "201: lorem ipsum", string(content))

	content, err = e.GetPage("bar", "200")
	assert.NoError(t, err)
	assert.Equal(t, "200 _ ok (all is ok)", string(content))

	content, err = e.GetPage("bar", "201")
	assert.NoError(t, err)
	assert.Equal(t, "201 _ lorem (ipsum)", string(content))

	content, err = e.GetPage("foo", "666")
	assert.ErrorIs(t, err, tpl.ErrUnknownPageCode)
	assert.Nil(t, content)

	content, err = e.GetPage("baz", "200")
	assert.ErrorIs(t, err, tpl.ErrUnknownTemplate)
	assert.Nil(t, content)
}

func TestErrorPages_GetPage_Concurrent(t *testing.T) {
	e := tpl.NewErrorPages()

	init := func() {
		e.AddTemplate("foo", []byte("{{ code }}: {{ message }} {{ description }}"))
		e.AddPage("200", "ok", "all is ok")
		e.AddPage("201", "lorem", "ipsum")
	}

	var wg sync.WaitGroup

	init()

	for i := 0; i < 1234; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()

			init() // make re-initialization
		}()

		go func() {
			defer wg.Done()

			content, err := e.GetPage("foo", "200")
			assert.NoError(t, err)
			assert.Equal(t, "200: ok all is ok", string(content))

			content, err = e.GetPage("foo", "201")
			assert.NoError(t, err)
			assert.Equal(t, "201: lorem ipsum", string(content))

			content, err = e.GetPage("foo", "666")
			assert.Error(t, err)
			assert.Nil(t, content)

			content, err = e.GetPage("bar", "200")
			assert.Error(t, err)
			assert.Nil(t, content)
		}()
	}

	wg.Wait()
}

func TestErrorPages_IteratePages(t *testing.T) {
	e := tpl.NewErrorPages()

	e.AddTemplate("foo", []byte("{{ code }}: {{ message }} {{ description }}"))
	e.AddTemplate("bar", []byte("{{ code }}: {{ message }} {{ description }}"))
	e.AddPage("200", "ok", "all is ok")
	e.AddPage("400", "Bad Request", "")

	visited := make(map[string]map[string]bool) // map[template]codes

	assert.NoError(t, e.IteratePages(func(template, code string, content []byte) error {
		if _, ok := visited[template]; !ok {
			visited[template] = make(map[string]bool)
		}

		visited[template][code] = true

		assert.NotNil(t, content)

		return nil
	}))

	assert.Len(t, visited, 2)
	assert.Len(t, visited["foo"], 2)
	assert.True(t, visited["foo"]["200"])
	assert.True(t, visited["foo"]["400"])
	assert.Len(t, visited["bar"], 2)
	assert.True(t, visited["bar"]["200"])
	assert.True(t, visited["bar"]["400"])
}

func TestErrorPages_IteratePages_WillReturnTheError(t *testing.T) {
	e := tpl.NewErrorPages()

	e.AddTemplate("foo", []byte("{{ code }}: {{ message }} {{ description }}"))
	e.AddPage("200", "ok", "all is ok")

	assert.EqualError(t, e.IteratePages(func(template, code string, content []byte) error {
		return errors.New("foo error")
	}), "foo error")
}

package tpl_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarampampam/error-pages/internal/tpl"
)

func TestErrors_Get(t *testing.T) {
	e := tpl.NewErrors(
		map[string][]byte{"foo": []byte("{{ code }}: {{ message }} {{ description }}")},
		map[string]tpl.Annotator{"200": {"ok", "all is ok"}},
	)

	content, err := e.Get("foo", "200")
	assert.NoError(t, err)
	assert.Equal(t, "200: ok all is ok", string(content))

	content, err = e.Get("foo", "666")
	assert.EqualError(t, err, "code \"666\" does not exists")
	assert.Nil(t, content)

	content, err = e.Get("bar", "200")
	assert.EqualError(t, err, "template \"bar\" does not exists")
	assert.Nil(t, content)
}

func TestErrors_GetConcurrent(t *testing.T) {
	e := tpl.NewErrors(
		map[string][]byte{"foo": []byte("{{ code }}: {{ message }} {{ description }}")},
		map[string]tpl.Annotator{"200": {"ok", "all is ok"}},
	)

	var wg sync.WaitGroup

	for i := 0; i < 1234; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			content, err := e.Get("foo", "200")
			assert.NoError(t, err)
			assert.Equal(t, "200: ok all is ok", string(content))

			content, err = e.Get("foo", "666")
			assert.Error(t, err)
			assert.Nil(t, content)
		}()
	}

	wg.Wait()
}

func TestErrors_VisitAll(t *testing.T) {
	e := tpl.NewErrors(
		map[string][]byte{
			"foo": []byte("{{ code }}: {{ message }} {{ description }}"),
			"bar": []byte("{{ code }}: {{ message }} {{ description }}"),
		},
		map[string]tpl.Annotator{
			"200": {"ok", "all is ok"},
			"400": {"Bad Request", "The server did not understand the request"},
		},
	)

	visited := make(map[string]map[string]bool) // map[template]codes

	assert.NoError(t, e.VisitAll(func(template, code string, content []byte) error {
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

func TestErrors_VisitAllWillReturnTheError(t *testing.T) {
	e := tpl.NewErrors(
		map[string][]byte{
			"foo": []byte("{{ code }}: {{ message }} {{ description }}"),
		},
		map[string]tpl.Annotator{
			"200": {"ok", "all is ok"},
		},
	)

	assert.EqualError(t, e.VisitAll(func(template, code string, content []byte) error {
		return errors.New("foo error")
	}), "foo error")
}

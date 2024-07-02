package error_page_test

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"gh.tarampamp.am/error-pages/internal/http/handlers/error_page"
	"gh.tarampamp.am/error-pages/internal/template"
)

func TestRenderedCache_CRUD(t *testing.T) {
	t.Parallel()

	var cache = error_page.NewRenderedCache(time.Millisecond)

	t.Run("has", func(t *testing.T) {
		assert.False(t, cache.Has("template", template.Props{}))
		cache.Put("template", template.Props{}, []byte("content"))
		assert.True(t, cache.Has("template", template.Props{}))

		assert.False(t, cache.Has("template", template.Props{Code: 1}))
		assert.False(t, cache.Has("foo", template.Props{Code: 1}))
	})

	t.Run("exists", func(t *testing.T) {
		var got, ok = cache.Get("template", template.Props{})

		assert.True(t, ok)
		assert.Equal(t, []byte("content"), got)

		cache.Clear()

		assert.False(t, cache.Has("template", template.Props{}))
	})

	t.Run("not exists", func(t *testing.T) {
		var got, ok = cache.Get("template", template.Props{Code: 2})

		assert.False(t, ok)
		assert.Nil(t, got)
	})

	t.Run("race condition provocation", func(t *testing.T) {
		var wg sync.WaitGroup

		for i := 0; i < 100; i++ {
			wg.Add(2)

			go func(i int) {
				defer wg.Done()

				cache.Get("template", template.Props{})
				cache.Put("template"+strconv.Itoa(i), template.Props{}, []byte("content"))
				cache.Has("template", template.Props{})
			}(i)

			go func() {
				defer wg.Done()

				cache.ClearExpired()
			}()
		}

		wg.Wait()
	})
}

func TestRenderedCache_Expiring(t *testing.T) {
	t.Parallel()

	var cache = error_page.NewRenderedCache(10 * time.Millisecond)

	cache.Put("template", template.Props{}, []byte("content"))
	cache.ClearExpired()
	assert.True(t, cache.Has("template", template.Props{}))

	<-time.After(10 * time.Millisecond)

	assert.True(t, cache.Has("template", template.Props{})) // expired, but not cleared yet
	cache.ClearExpired()
	assert.False(t, cache.Has("template", template.Props{})) // cleared
}

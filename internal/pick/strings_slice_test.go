package pick_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarampampam/error-pages/internal/pick"
)

func TestStringsSlice_Pick(t *testing.T) {
	t.Parallel()

	t.Run("first", func(t *testing.T) {
		t.Parallel()

		for i := uint8(0); i < 100; i++ {
			assert.Equal(t, "", pick.NewStringsSlice([]string{}, pick.First).Pick())
		}

		p := pick.NewStringsSlice([]string{"foo", "bar", "baz"}, pick.First)

		for i := uint8(0); i < 100; i++ {
			assert.Equal(t, "foo", p.Pick())
		}
	})

	t.Run("random once", func(t *testing.T) {
		t.Parallel()

		for i := uint8(0); i < 100; i++ {
			assert.Equal(t, "", pick.NewStringsSlice([]string{}, pick.RandomOnce).Pick())
		}

		var (
			p      = pick.NewStringsSlice([]string{"foo", "bar", "baz"}, pick.RandomOnce)
			picked = p.Pick()
		)

		for i := uint8(0); i < 100; i++ {
			assert.Equal(t, picked, p.Pick())
		}
	})

	t.Run("random every time", func(t *testing.T) {
		t.Parallel()

		for i := uint8(0); i < 100; i++ {
			assert.Equal(t, "", pick.NewStringsSlice([]string{}, pick.RandomEveryTime).Pick())
		}

		for i := uint8(0); i < 100; i++ {
			p := pick.NewStringsSlice([]string{"foo", "bar", "baz"}, pick.RandomEveryTime)

			assert.NotEqual(t, p.Pick(), p.Pick())
		}
	})
}

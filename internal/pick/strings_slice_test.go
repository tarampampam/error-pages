package pick_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tarampampam/error-pages/internal/pick"
)

func TestStringsSlice_Pick(t *testing.T) {
	t.Run("first", func(t *testing.T) {
		for i := uint8(0); i < 100; i++ {
			assert.Equal(t, "", pick.NewStringsSlice([]string{}, pick.First).Pick())
		}

		p := pick.NewStringsSlice([]string{"foo", "bar", "baz"}, pick.First)

		for i := uint8(0); i < 100; i++ {
			assert.Equal(t, "foo", p.Pick())
		}
	})

	t.Run("random once", func(t *testing.T) {
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
		for i := uint8(0); i < 100; i++ {
			assert.Equal(t, "", pick.NewStringsSlice([]string{}, pick.RandomEveryTime).Pick())
		}

		for i := uint8(0); i < 100; i++ {
			p := pick.NewStringsSlice([]string{"foo", "bar", "baz"}, pick.RandomEveryTime)

			assert.NotEqual(t, p.Pick(), p.Pick())
		}
	})
}

func TestNewStringsSliceWithInterval_Pick(t *testing.T) {
	t.Run("first", func(t *testing.T) {
		for i := uint8(0); i < 50; i++ {
			p := pick.NewStringsSliceWithInterval([]string{}, pick.First, time.Millisecond)
			assert.Equal(t, "", p.Pick())
			assert.NoError(t, p.Close())
			assert.Panics(t, func() { p.Pick() })
		}

		p := pick.NewStringsSliceWithInterval([]string{"foo", "bar", "baz"}, pick.First, time.Millisecond)

		for i := uint8(0); i < 50; i++ {
			assert.Equal(t, "foo", p.Pick())

			<-time.After(time.Millisecond * 2)
		}

		assert.NoError(t, p.Close())
		assert.Error(t, p.Close())
		assert.Panics(t, func() { p.Pick() })
	})

	t.Run("random once", func(t *testing.T) {
		for i := uint8(0); i < 50; i++ {
			p := pick.NewStringsSliceWithInterval([]string{}, pick.RandomOnce, time.Millisecond)
			assert.Equal(t, "", p.Pick())
			assert.NoError(t, p.Close())
			assert.Panics(t, func() { p.Pick() })
		}

		var (
			p      = pick.NewStringsSliceWithInterval([]string{"foo", "bar", "baz"}, pick.RandomOnce, time.Millisecond)
			picked = p.Pick()
		)

		for i := uint8(0); i < 50; i++ {
			assert.Equal(t, picked, p.Pick())

			<-time.After(time.Millisecond * 2)
		}

		assert.NoError(t, p.Close())
		assert.Error(t, p.Close())
		assert.Panics(t, func() { p.Pick() })
	})

	t.Run("random every time", func(t *testing.T) {
		for i := uint8(0); i < 50; i++ {
			p := pick.NewStringsSliceWithInterval([]string{}, pick.RandomEveryTime, time.Millisecond)
			assert.Equal(t, "", p.Pick())
			assert.NoError(t, p.Close())
			assert.Panics(t, func() { p.Pick() })
		}

		var changed int

		for i := uint8(0); i < 50; i++ {
			p := pick.NewStringsSliceWithInterval([]string{"foo", "bar", "baz"}, pick.RandomEveryTime, time.Millisecond) //nolint:lll

			one, two := p.Pick(), p.Pick()
			assert.Equal(t, one, two)

			<-time.After(time.Millisecond * 2)

			three, four := p.Pick(), p.Pick()
			assert.Equal(t, three, four)

			if one != three {
				changed++
			}

			assert.NoError(t, p.Close())
			assert.Error(t, p.Close())
			assert.Panics(t, func() { p.Pick() })
		}

		assert.GreaterOrEqual(t, changed, 25)
	})
}

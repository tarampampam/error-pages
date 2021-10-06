package pick_test

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tarampampam/error-pages/internal/pick"
)

func TestStringsSlice_Pick_First(t *testing.T) {
	for name, items := range map[string][]string{
		"0 item":  {},
		"1 item":  {"foo"},
		"3 items": {"foo", "bar", "baz"},
	} {
		t.Run(name, func(t *testing.T) {
			p := pick.NewStringsSlice(items, pick.First)

			for i := 0; i < 100; i++ {
				if len(items) == 0 {
					assert.Equal(t, "", p.Pick())
				} else {
					assert.Equal(t, "foo", p.Pick())
				}
			}
		})
	}
}

func TestStringsSlice_Pick_RandomOnce(t *testing.T) {
	p := pick.NewStringsSlice([]string{}, pick.RandomOnce)
	assert.Equal(t, "", p.Pick())

	p = pick.NewStringsSlice([]string{"foo"}, pick.RandomOnce)
	assert.Equal(t, "foo", p.Pick())

	dataSet := randomStringsSlice(t, 2048) // if this test will fail - Increase this value
	p = pick.NewStringsSlice(dataSet, pick.RandomOnce)
	picked := p.Pick()

	assert.NotEqual(t, dataSet[0], p.Pick())

	for i := 0; i < 32; i++ {
		assert.Equal(t, picked, p.Pick())
	}
}

func TestStringsSlice_Pick_RandomEveryTime(t *testing.T) {
	p := pick.NewStringsSlice([]string{}, pick.RandomEveryTime)
	assert.Equal(t, "", p.Pick())

	p = pick.NewStringsSlice([]string{"foo"}, pick.RandomEveryTime)
	assert.Equal(t, "foo", p.Pick())

	dataSet := randomStringsSlice(t, 2048) // if this test will fail - Increase this value
	p = pick.NewStringsSlice(dataSet, pick.RandomEveryTime)

	lastPicked := p.Pick()

	for i := 0; i < 32; i++ {
		picked := p.Pick()

		assert.NotEqual(t, lastPicked, picked)
		lastPicked = picked
	}
}

func TestStringsSlice_Pick_UnsupportedMode(t *testing.T) {
	p := pick.NewStringsSlice([]string{}, 255)
	assert.Equal(t, "", p.Pick())

	p = pick.NewStringsSlice([]string{"foo"}, 255)
	assert.Equal(t, "foo", p.Pick())

	p = pick.NewStringsSlice([]string{"foo", "bar"}, 255)

	assert.Panics(t, func() { p.Pick() })
}

func randomStringsSlice(t *testing.T, itemsCount int) []string {
	t.Helper()

	var (
		rnd     = rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
		items   = make([]string, itemsCount)
		letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-+=")
	)

	for i := 0; i < len(items); i++ {
		b := make([]rune, 32)

		for j := range b {
			b[j] = letters[rnd.Intn(len(letters))]
		}

		items[i] = string(b)
	}

	return items
}

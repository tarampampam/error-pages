package pick_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tarampampam/error-pages/internal/pick"
)

func TestPicker_NextIndex_First(t *testing.T) {
	for i := uint32(0); i < 100; i++ {
		p := pick.NewPicker(i, pick.First)

		for j := uint8(0); j < 100; j++ {
			assert.Equal(t, uint32(0), p.NextIndex())
		}
	}
}

func TestPicker_NextIndex_RandomOnce(t *testing.T) {
	for i := uint8(0); i < 10; i++ {
		assert.Equal(t, uint32(0), pick.NewPicker(0, pick.RandomOnce).NextIndex())
	}

	for i := uint8(10); i < 100; i++ {
		p := pick.NewPicker(uint32(i), pick.RandomOnce)

		next := p.NextIndex()
		assert.LessOrEqual(t, next, uint32(i))

		for j := uint8(0); j < 100; j++ {
			assert.Equal(t, next, p.NextIndex())
		}
	}
}

func TestPicker_NextIndex_RandomEveryTime(t *testing.T) {
	for i := uint8(0); i < 10; i++ {
		assert.Equal(t, uint32(0), pick.NewPicker(0, pick.RandomEveryTime).NextIndex())
	}

	for i := uint8(1); i < 100; i++ {
		p := pick.NewPicker(uint32(i), pick.RandomEveryTime)

		for j := uint8(0); j < 100; j++ {
			one, two := p.NextIndex(), p.NextIndex()

			assert.LessOrEqual(t, one, uint32(i))
			assert.LessOrEqual(t, two, uint32(i))
			assert.NotEqual(t, one, two)
		}
	}
}

func TestPicker_NextIndex_Unsupported(t *testing.T) {
	assert.Panics(t, func() { pick.NewPicker(1, 255).NextIndex() })
}

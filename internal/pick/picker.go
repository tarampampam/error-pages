package pick

import (
	"math/rand"
	"sync"
	"time"
)

type pickMode = byte

const (
	First           pickMode = 1 + iota // Always pick the first element (index = 0)
	RandomOnce                          // Pick random element once (any future Pick calls will return the same element)
	RandomEveryTime                     // Always Pick the random element
)

type picker struct {
	mode   pickMode
	rand   *rand.Rand // will be nil for the First pick mode
	maxIdx uint32

	mu      sync.Mutex
	lastIdx uint32
}

const unsetIdx uint32 = 4294967295

func NewPicker(maxIdx uint32, mode pickMode) *picker {
	var p = &picker{
		maxIdx:  maxIdx,
		mode:    mode,
		lastIdx: unsetIdx,
	}

	if mode != First {
		p.rand = rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	}

	return p
}

// NextIndex returns an index for the next element (based on pickMode).
func (p *picker) NextIndex() uint32 {
	if p.maxIdx == 0 {
		return 0
	}

	switch p.mode {
	case First:
		return 0

	case RandomOnce:
		if p.lastIdx == unsetIdx {
			p.mu.Lock()
			defer p.mu.Unlock()

			p.lastIdx = uint32(p.rand.Intn(int(p.maxIdx)))
		}

		return p.lastIdx

	case RandomEveryTime:
		var idx = uint32(p.rand.Intn(int(p.maxIdx + 1)))

		p.mu.Lock()
		defer p.mu.Unlock()

		if idx == p.lastIdx {
			p.lastIdx++
		} else {
			p.lastIdx = idx
		}

		if p.lastIdx > p.maxIdx { // overflow?
			p.lastIdx = 0
		}

		return p.lastIdx

	default:
		panic("picker.NextIndex(): unsupported mode")
	}
}

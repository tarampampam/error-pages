package pick

import (
	"math/rand"
	"time"
)

type pickMode byte

const (
	First           pickMode = 1 + iota // Always pick the first element
	RandomOnce                          // Pick random element once (any future Pick calls will return the same element)
	RandomEveryTime                     // Always Pick the random element
)

type StringsSlice struct {
	items       []string
	mode        pickMode
	lastUsedIdx int        // -1 when unset, needed for RandomOnce mode
	rnd         *rand.Rand // will be nil for the First mode
}

// NewStringsSlice creates new StringsSlice.
func NewStringsSlice(items []string, mode pickMode) *StringsSlice {
	var rnd *rand.Rand

	if mode != First {
		rnd = rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	}

	return &StringsSlice{
		items:       items,
		mode:        mode,
		lastUsedIdx: -1,
		rnd:         rnd,
	}
}

// Pick an element from the strings slice.
func (s *StringsSlice) Pick() string {
	if l := len(s.items); l == 0 {
		return ""
	} else if l == 1 {
		return s.items[0]
	}

	switch s.mode {
	case First:
		return s.items[0]

	case RandomOnce:
		if s.lastUsedIdx == -1 {
			s.lastUsedIdx = s.rnd.Intn(len(s.items))
		}

		return s.items[s.lastUsedIdx]

	case RandomEveryTime:
		return s.items[s.rnd.Intn(len(s.items))]

	default:
		panic("pick: unsupported mode")
	}
}

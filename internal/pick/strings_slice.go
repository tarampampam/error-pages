package pick

import (
	"errors"
	"sync"
	"time"
)

type StringsSlice struct {
	s []string
	p *picker
}

// NewStringsSlice creates new StringsSlice.
func NewStringsSlice(items []string, mode pickMode) *StringsSlice {
	maxIdx := len(items) - 1

	if maxIdx < 0 {
		maxIdx = 0
	}

	return &StringsSlice{s: items, p: NewPicker(uint32(maxIdx), mode)}
}

// Pick an element from the strings slice.
func (s *StringsSlice) Pick() string {
	if len(s.s) == 0 {
		return ""
	}

	return s.s[s.p.NextIndex()]
}

type StringsSliceWithInterval struct {
	s []string
	p *picker
	d time.Duration

	idxMu sync.RWMutex
	idx   uint32

	close    chan struct{}
	closedMu sync.RWMutex
	closed   bool
}

// NewStringsSliceWithInterval creates new StringsSliceWithInterval.
func NewStringsSliceWithInterval(items []string, mode pickMode, interval time.Duration) *StringsSliceWithInterval {
	maxIdx := len(items) - 1

	if maxIdx < 0 {
		maxIdx = 0
	}

	if interval <= time.Duration(0) {
		panic("NewStringsSliceWithInterval: wrong interval")
	}

	s := &StringsSliceWithInterval{
		s:     items,
		p:     NewPicker(uint32(maxIdx), mode),
		d:     interval,
		close: make(chan struct{}, 1),
	}

	s.next()

	go s.rotate()

	return s
}

func (s *StringsSliceWithInterval) rotate() {
	defer close(s.close)

	timer := time.NewTimer(s.d)
	defer timer.Stop()

	for {
		select {
		case <-s.close:
			return

		case <-timer.C:
			s.next()
			timer.Reset(s.d)
		}
	}
}

func (s *StringsSliceWithInterval) next() {
	idx := s.p.NextIndex()

	s.idxMu.Lock()
	s.idx = idx
	s.idxMu.Unlock()
}

// Pick an element from the strings slice.
func (s *StringsSliceWithInterval) Pick() string {
	if s.isClosed() {
		panic("StringsSliceWithInterval.Pick(): closed")
	}

	if len(s.s) == 0 {
		return ""
	}

	s.idxMu.RLock()
	defer s.idxMu.RUnlock()

	return s.s[s.idx]
}

func (s *StringsSliceWithInterval) isClosed() (closed bool) {
	s.closedMu.RLock()
	closed = s.closed
	s.closedMu.RUnlock()

	return
}

func (s *StringsSliceWithInterval) Close() error {
	if s.isClosed() {
		return errors.New("closed")
	}

	s.closedMu.Lock()
	s.closed = true
	s.closedMu.Unlock()

	s.close <- struct{}{}

	return nil
}

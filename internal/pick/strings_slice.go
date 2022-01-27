package pick

type StringsSlice struct {
	s []string
	p *picker
}

// NewStringsSlice creates new StringsSlice.
func NewStringsSlice(items []string, mode pickMode) *StringsSlice {
	return &StringsSlice{s: items, p: NewPicker(uint32(len(items)-1), mode)}
}

// Pick an element from the strings slice.
func (s *StringsSlice) Pick() string {
	if len(s.s) == 0 {
		return ""
	}

	return s.s[s.p.NextIndex()]
}

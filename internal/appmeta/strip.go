package appmeta

// stripVersionPrefix removes the 'v' or 'V' prefix from a version string if it exists and is followed by a digit.
func stripVersionPrefix(v string) string {
	if len(v) > 1 && (v[0] == 'v' || v[0] == 'V') && v[1] >= '0' && v[1] <= '9' {
		return v[1:]
	}

	return v
}

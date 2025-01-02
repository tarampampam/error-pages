package config

import (
	"slices"
	"strconv"
)

type (
	CodeDescription struct {
		// Message is a short description of the HTTP error.
		Message string

		// Description is a longer description of the HTTP error.
		Description string
	}

	// Codes is a map of HTTP codes to their descriptions.
	//
	// The codes may be written in a non-strict manner. For example, they may be "4xx", "4XX", or "4**".
	// If the map contains both "404" and "4xx" keys, and we search for "404", the "404" key will be returned.
	// However, if we search for "405", "400", or any non-existing code that starts with "4" and its length is 3,
	// the value under the key "4xx" will be retrieved.
	//
	// The length of the code (in string format) is matter.
	Codes map[string]CodeDescription // map[http_code]description
)

// Find searches the closest match for the given HTTP code, written in a non-strict manner. Read [Codes] for more
// information.
func (c Codes) Find(httpCode uint16) (CodeDescription, bool) { //nolint:gocyclo
	if len(c) == 0 { // empty map, fast return
		return CodeDescription{}, false
	}

	var code = strconv.FormatUint(uint64(httpCode), 10)

	if desc, ok := c[code]; ok { // search for the exact match
		return desc, true
	}

	var (
		keysMap   = make(map[string][]rune, len(c))
		codeRunes = []rune(code)
	)

	for key := range c { // take only the keys that are the same length and start with the same character or a wildcard
		if kr := []rune(key); len(kr) > 0 && len(kr) == len(codeRunes) && isWildcardOr(kr[0], codeRunes[0]) {
			keysMap[key] = kr
		}
	}

	if len(keysMap) == 0 { // no matches found using the first rune comparison
		return CodeDescription{}, false
	}

	var matchedMap = make(map[string]uint16, len(keysMap)) // map[mapKey]wildcardMatchedCount

	for mapKey, keyRunes := range keysMap { // search for the closest match
		var wildcardMatchedCount uint16 = 0

		for i := 0; i < len(codeRunes); i++ { // loop through each httpCode rune
			var keyRune, codeRune = keyRunes[i], codeRunes[i]

			if wm := isWildcard(keyRune); wm || keyRune == codeRune {
				if wm {
					wildcardMatchedCount++
				}

				if i == len(codeRunes)-1 { // is the last rune?
					matchedMap[mapKey] = wildcardMatchedCount
				}

				continue
			}

			break
		}
	}

	if len(matchedMap) == 0 { // no matches found
		return CodeDescription{}, false
	} else if len(matchedMap) == 1 { // only one match found
		for mapKey := range matchedMap {
			return c[mapKey], true
		}
	}

	// multiple matches found, find the most specific one based on the wildcard matched count (pick the one with the
	// least wildcards)
	var (
		minCount uint16
		key      string
	)

	for mapKey, count := range matchedMap {
		if minCount == 0 || count < minCount {
			minCount, key = count, mapKey
		}
	}

	return c[key], true
}

func isWildcard(r rune) bool       { return r == '*' || r == 'x' || r == 'X' }
func isWildcardOr(r, or rune) bool { return isWildcard(r) || r == or }

// Codes returns all HTTP codes sorted alphabetically.
func (c Codes) Codes() []string {
	var codes = make([]string, 0, len(c))

	for code := range c {
		codes = append(codes, code)
	}

	slices.Sort(codes)

	return codes
}

// Has checks if the HTTP code exists.
func (c Codes) Has(code string) (found bool) { _, found = c[code]; return } //nolint:nlreturn

// Get returns the HTTP code description by the specified code, if it exists.
func (c Codes) Get(code string) (data CodeDescription, ok bool) { data, ok = c[code]; return } //nolint:nlreturn

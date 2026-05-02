package random

import "math/rand/v2"

// String generates a random alphanumeric string of the specified length.
// Uses the package-level goroutine-safe source from math/rand/v2.
func String(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, length)

	for i := range b {
		b[i] = charset[rand.IntN(len(charset))] //nolint:gosec
	}

	return string(b)
}

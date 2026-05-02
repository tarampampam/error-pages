package middleware

import (
	"net/http"
	"slices"
)

// Apply applies the provided middlewares to the given handler. The middlewares are applied in the order
// they are provided, meaning that the first middleware in the slice will be the outermost one, and the last
// middleware will be the innermost one.
//
// Apply(h, A, B, C) results in A(B(C(h))).
func Apply(h http.Handler, mw ...func(http.Handler) http.Handler) http.Handler {
	if h == nil {
		panic("nil handler")
	}

	for _, v := range slices.Backward(mw) {
		if v == nil {
			continue // skip nil middlewares
		}

		h = v(h)
	}

	return h
}

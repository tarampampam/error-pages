package middleware

import (
	"net/http"
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

	for i := len(mw) - 1; i >= 0; i-- {
		if mw[i] == nil {
			continue // skip nil middlewares
		}

		h = mw[i](h)
	}

	return h
}

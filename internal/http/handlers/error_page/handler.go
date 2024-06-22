package error_page

import "net/http"

func New() http.Handler {
	var body = []byte("error page")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write(body)
	})
}

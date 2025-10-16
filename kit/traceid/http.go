package traceid

import "net/http"

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := FromContext(r.Context()); !ok {
			ctx := WithContext(r.Context(), New())
			r = r.WithContext(ctx)
		}
		next.ServeHTTP(w, r)
	})
}

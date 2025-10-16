package hlog

import (
	"log/slog"
	"net/http"
	"slices"
	"time"
)

type MiddlewareFunc func(http.Handler) http.Handler

func Middleware(lg *slog.Logger) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				rr := RecordResponse(next, w, r)
				lg.LogAttrs(
					r.Context(),
					slog.LevelInfo, "request",
					slog.Any("path", r.URL),
					slog.String("method", r.Method),
					slog.Duration("duration", rr.Duration),
					slog.Int("status", rr.StatusCode),
				)
			},
		)
	}
}

func Duration(d *time.Duration) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			defer func() { *d = time.Since(start) }()
			next.ServeHTTP(w, r)
		})
	}
}

type Response struct {
	Duration   time.Duration
	StatusCode int
}

func RecordResponse(
	h http.Handler,
	w http.ResponseWriter, r *http.Request,
) Response {
	var rr Response
	mws := []MiddlewareFunc{
		Duration(&rr.Duration),
		StatusCode(&rr.StatusCode),
	}
	for _, wrap := range slices.Backward(mws) {
		h = wrap(h)
	}
	h.ServeHTTP(w, r)
	return rr
}

type Interceptor struct {
	http.ResponseWriter
	OnWriteHeader func(code int)
}

func (ic *Interceptor) WriteHeader(code int) {
	if ic.OnWriteHeader != nil {
		ic.OnWriteHeader(code)
	}
	ic.ResponseWriter.WriteHeader(code)
}

func (ic *Interceptor) Unwrap() http.ResponseWriter {
	return ic.ResponseWriter
}

func StatusCode(n *int) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*n = http.StatusOK
			w = &Interceptor{
				ResponseWriter: w,
				OnWriteHeader: func(code int) {
					*n = code
				},
			}
			next.ServeHTTP(w, r)
		})
	}
}

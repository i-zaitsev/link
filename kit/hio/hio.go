package hio

import "net/http"

type Handler func(w http.ResponseWriter, r *http.Request) Handler

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if next := h(w, r); next != nil {
		next.ServeHTTP(w, r)
	}
}

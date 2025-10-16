package hio

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Responder struct {
	err func(error) Handler
}

func NewResponder(err func(error) Handler) Responder {
	return Responder{err}
}

func (r Responder) Error(format string, args ...any) Handler {
	return r.err(fmt.Errorf(format, args...))
}

func (r Responder) Redirect(code int, url string) Handler {
	return func(w http.ResponseWriter, r *http.Request) Handler {
		http.Redirect(w, r, url, code)
		return nil
	}
}

func (r Responder) Text(code int, message string) Handler {
	return func(w http.ResponseWriter, r *http.Request) Handler {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(code)
		_, _ = fmt.Fprintf(w, message)
		return nil
	}
}

func (r Responder) JSON(code int, from any) Handler {
	data, err := json.Marshal(from)
	if err != nil {
		return r.Error("encoding json: %w", err)
	}
	return func(w http.ResponseWriter, r *http.Request) Handler {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_, _ = w.Write(data)
		return nil
	}
}

package rest

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/i-zaitsev/link"
	"github.com/i-zaitsev/link/kit/hio"
)

func Shorten(lg *slog.Logger, links *link.Shortener) http.Handler {
	with := newResponder(lg)
	return hio.Handler(func(w http.ResponseWriter, r *http.Request) hio.Handler {
		var lnk link.Link
		if err := hio.DecodeJSON(
			hio.MaxBytesReader(w, r.Body, 4096),
			&lnk,
		); err != nil {
			return with.Error("decoding: %w: %w", err, link.ErrBadRequest)
		}
		key, err := links.Shorten(r.Context(), lnk)
		if err != nil {
			return with.Error("shortening: %w", err)
		}
		return with.JSON(http.StatusCreated, map[string]link.Key{
			"key": key,
		})
	})
}

func Resolve(lg *slog.Logger, links *link.Shortener) http.Handler {
	with := newResponder(lg)
	return hio.Handler(func(w http.ResponseWriter, r *http.Request) hio.Handler {
		lnk, err := links.Resolve(r.Context(), link.Key(r.PathValue("key")))
		if err != nil {
			return with.Error("resolving: %w", err)
		}
		return with.Redirect(http.StatusFound, lnk.URL)
	})
}

func httpError(
	w http.ResponseWriter,
	r *http.Request,
	lg *slog.Logger,
	err error,
) {
	code := http.StatusInternalServerError
	switch {
	case errors.Is(err, link.ErrBadRequest):
		code = http.StatusBadRequest
	case errors.Is(err, link.ErrConflict):
		code = http.StatusConflict
	case errors.Is(err, link.ErrNotFound):
		code = http.StatusNotFound
	}
	if code == http.StatusInternalServerError {
		lg.ErrorContext(r.Context(), "internal", "error", err)
		err = link.ErrInternal
	}
	http.Error(w, err.Error(), code)
}

func newResponder(lg *slog.Logger) hio.Responder {
	err := func(err error) hio.Handler {
		return func(w http.ResponseWriter, r *http.Request) hio.Handler {
			httpError(w, r, lg, err)
			return nil
		}
	}
	return hio.NewResponder(err)
}

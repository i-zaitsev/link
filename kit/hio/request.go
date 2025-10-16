package hio

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func DecodeJSON(from io.Reader, to any) error {
	data, err := io.ReadAll(from)
	if err != nil {
		return fmt.Errorf("reading: %w", err)
	}
	if err := json.Unmarshal(data, to); err != nil {
		return fmt.Errorf("unmarshaling json: %w", err)
	}
	v, ok := to.(interface{ Validate() error })
	if ok {
		if err := v.Validate(); err != nil {
			return fmt.Errorf("validating: %w", err)
		}
	}
	return nil
}

func MaxBytesReader(w http.ResponseWriter, rc io.ReadCloser, max int64) io.ReadCloser {
	type unwrapper interface {
		Unwrap() http.ResponseWriter
	}
	for {
		if v, ok := w.(unwrapper); !ok {
			break
		} else {
			w = v.Unwrap()
		}
	}
	return http.MaxBytesReader(w, rc, max)
}

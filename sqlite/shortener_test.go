package sqlite

import (
	"errors"
	"testing"

	"github.com/i-zaitsev/link"
)

func TestShortenerShorten(t *testing.T) {
	t.Parallel()
	lnk := link.Link{
		Key: "foo",
		URL: "https://example.com",
	}
	shortener := NewShortener(DialTestDB(t))
	key, err := shortener.Shorten(t.Context(), lnk)
	if err != nil {
		t.Fatalf("Shorten: unexpected error: %v", err)
	}
	if key != "foo" {
		t.Errorf(`got key %q, want "foo"`, key)
	}
	_, err = shortener.Shorten(t.Context(), lnk)
	if !errors.Is(err, link.ErrConflict) {
		t.Fatalf("\ngot err = %v\nwant ErrConflict for duplicate key", err)
	}
}

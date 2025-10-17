package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/i-zaitsev/link"
)

type Shortener struct {
	db *sql.DB
}

func NewShortener(db *sql.DB) *Shortener {
	return &Shortener{db}
}

func (s *Shortener) Shorten(ctx context.Context, lnk link.Link) (link.Key, error) {
	var err error
	if lnk.Key, err = link.Shorten(lnk); err != nil {
		return "", fmt.Errorf("%w: %w", err, link.ErrBadRequest)
	}
	_, err = s.db.ExecContext(
		ctx,
		`INSERT INTO links (short_key, uri) VALUES ($1, $2)`,
		lnk.Key, lnk.URL,
	)
	if err != nil {
		return "", fmt.Errorf("saving: %w: %w", err, link.ErrInternal)
	}
	return lnk.Key, nil
}

func DialTestDB(tb testing.TB) *sql.DB {
	tb.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", tb.Name())
	db, err := Dial(tb.Context(), dsn)
	if err != nil {
		tb.Fatalf("DialTestDB: %v", err)
	}
	tb.Cleanup(func() {
		if err := db.Close(); err != nil {
			tb.Errorf("DialTestDB: closing db: %v", err)
		}
	})
	return db
}

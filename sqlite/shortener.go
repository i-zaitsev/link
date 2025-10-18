package sqlite

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/base64"
	"errors"
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
		lnk.Key, base64String(lnk.URL),
	)
	if isPrimaryKeyViolation(err) {
		return "", fmt.Errorf("saving: %w", link.ErrConflict)
	}
	if err != nil {
		return "", fmt.Errorf("saving: %w: %w", err, link.ErrInternal)
	}
	return lnk.Key, nil
}

func (s *Shortener) Resolve(ctx context.Context, key link.Key) (link.Link, error) {
	keyValidation := func(key link.Key) error {
		if key.Empty() {
			return fmt.Errorf("resolve: %w", link.ErrBadRequest)
		}
		if err := key.Validate(); err != nil {
			return fmt.Errorf("validating: %w: %w", err, link.ErrBadRequest)
		}
		return nil
	}

	dbValidation := func(err error) error {
		if errors.Is(err, sql.ErrNoRows) {
			return link.ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("retrieving %w: %w", err, link.ErrInternal)
		}
		return nil
	}

	if err := keyValidation(key); err != nil {
		return link.Link{}, err
	}

	var uri base64String
	err := s.db.QueryRowContext(ctx, `SELECT uri FROM links WHERE short_key = $1`, key).Scan(&uri)
	if err := dbValidation(err); err != nil {
		return link.Link{}, err
	}

	return link.Link{
		Key: key,
		URL: uri.String(),
	}, nil
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

type base64String string

func (b base64String) Value() (driver.Value, error) {
	return base64.StdEncoding.EncodeToString([]byte(b)), nil
}

func (b *base64String) Scan(src any) error {
	s, ok := src.(string)
	if !ok {
		return fmt.Errorf("decoding: %q is %T, not string", s, src)
	}
	dst, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return fmt.Errorf("decoding: %q: %w", s, err)
	}
	*b = base64String(dst)
	return nil
}

func (b base64String) String() string {
	return string(b)
}

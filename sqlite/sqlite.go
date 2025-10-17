package sqlite

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"

	"modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

func Dial(ctx context.Context, dsn string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("opening: %w", err)
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("pinging: %w", err)
	}
	if _, err := db.ExecContext(ctx, schema); err != nil {
		return nil, fmt.Errorf("applying schema: %w", err)
	}
	return db, nil
}

func isPrimaryKeyViolation(err error) bool {
	var sqliteErr *sqlite.Error
	if errors.As(err, &sqliteErr) {
		return sqliteErr.Code() == 1555
	}
	return false
}

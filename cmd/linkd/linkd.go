package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/i-zaitsev/link/kit/hlog"
	"github.com/i-zaitsev/link/kit/traceid"
	"github.com/i-zaitsev/link/rest"
	"github.com/i-zaitsev/link/sqlite"
)

type config struct {
	http struct {
		addr     string
		timeouts struct{ read, idle time.Duration }
	}
	db struct{ dsn string }
	lg *slog.Logger
}

func main() {
	var cfg config
	flag.StringVar(
		&cfg.http.addr,
		"http.addr", "localhost:8080", "http address to listen on",
	)
	flag.DurationVar(
		&cfg.http.timeouts.read,
		"http.timeout.read", 20*time.Second, "read timeout",
	)
	flag.DurationVar(
		&cfg.http.timeouts.idle,
		"http.timeout.idle", 40*time.Second, "idle timeout",
	)
	flag.StringVar(
		&cfg.db.dsn,
		"db.dsn", "file:links.db?mode=rwc", "database DSN",
	)
	flag.Parse()

	cfg.lg = slog.New(slog.NewTextHandler(os.Stderr, nil)).With("app", "linkd")
	cfg.lg.Info("starting", "addr", cfg.http.addr)

	if err := run(context.Background(), cfg); err != nil {
		cfg.lg.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg config) error {
	db, err := sqlite.Dial(ctx, cfg.db.dsn)
	if err != nil {
		return fmt.Errorf("dialing database: %w", err)
	}

	shortener := sqlite.NewShortener(db)

	lg := slog.New(traceid.NewLogHandler(cfg.lg.Handler()))

	mux := http.NewServeMux()
	mux.Handle("POST /shorten", rest.Shorten(lg, shortener))
	mux.Handle("GET /r/{key}", rest.Resolve(lg, shortener))
	mux.HandleFunc("GET /health", rest.Health)

	loggerMiddleware := hlog.Middleware(lg)

	server := &http.Server{
		Handler:     traceid.Middleware(loggerMiddleware(mux)),
		Addr:        cfg.http.addr,
		ReadTimeout: cfg.http.timeouts.read,
		IdleTimeout: cfg.http.timeouts.idle,
	}

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server closed unexpectedly: %w", err)
	}

	return nil
}

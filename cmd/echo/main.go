// Command echo runs the echo HTTP service.
//
// The service exposes /healthz (GET) and /echo (POST) endpoints. It
// demonstrates the scaffold's conventions: net/http stdlib server, slog
// logging, graceful shutdown driven by signal.NotifyContext, and ctx as
// the first parameter.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gyc567/go-modern-ai-stack/internal/echo"
)

const (
	defaultPort         = ":8080"
	defaultReadTimeout  = 5 * time.Second
	defaultWriteTimeout = 10 * time.Second
	defaultShutdownWait = 10 * time.Second
)

func main() {
	if err := run(); err != nil {
		slog.Error("echo service exited with error", "err", err)
		os.Exit(1)
	}
}

func run() error {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(log)

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	mux := http.NewServeMux()
	mux.Handle("GET /healthz", echo.NewHealthHandler())
	mux.Handle("POST /echo", echo.NewEchoHandler(log))

	srv := &http.Server{
		Addr:         port,
		Handler:      mux,
		ReadTimeout:  defaultReadTimeout,
		WriteTimeout: defaultWriteTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.InfoContext(ctx, "echo service listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("listen: %w", err)
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		log.InfoContext(context.Background(), "shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), defaultShutdownWait)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	log.InfoContext(context.Background(), "echo service stopped cleanly")
	return nil
}

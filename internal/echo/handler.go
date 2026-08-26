// Package echo implements the /healthz and /echo HTTP handlers.
//
// Handlers follow the scaffold's conventions: ctx-aware, slog-based logging,
// explicit error wrapping, and bounded body reads.
package echo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

// MaxBodyBytes caps the /echo request body size.
const MaxBodyBytes = 1 << 20 // 1 MiB

// EchoResponse is the JSON body returned by POST /echo.
type EchoResponse struct {
	Length      int    `json:"length"`
	ContentType string `json:"content_type"`
	Body        string `json:"body"`
}

// NewHealthHandler returns a handler that responds 200 OK with body "ok".
func NewHealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok")
	})
}

// NewEchoHandler returns a handler that echoes back the request body as JSON.
func NewEchoHandler(log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		contentType := r.Header.Get("Content-Type")
		log.InfoContext(ctx, "echo request received", "content_type", contentType)

		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, MaxBodyBytes))
		if err != nil {
			status := http.StatusBadRequest
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				status = http.StatusRequestEntityTooLarge
			}
			respondError(ctx, w, log, status, fmt.Errorf("read body: %w", err))
			return
		}
		defer func() { _ = r.Body.Close() }()

		resp := EchoResponse{
			Length:      len(body),
			ContentType: contentType,
			Body:        string(body),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.ErrorContext(ctx, "encode response", "err", fmt.Errorf("encode: %w", err))
			return
		}
		log.InfoContext(ctx, "echo served", "length", resp.Length)
	})
}

func respondError(ctx context.Context, w http.ResponseWriter, log *slog.Logger, code int, err error) {
	log.WarnContext(ctx, "handler error", "code", code, "err", err)
	http.Error(w, http.StatusText(code), code)
}

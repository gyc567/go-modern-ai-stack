package echo_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/gyc567/go-modern-ai-stack/internal/echo"
)

func TestNewHealthHandler(t *testing.T) {
	t.Parallel()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	echo.NewHealthHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusOK)
	}
	if got, want := rec.Body.String(), "ok"; got != want {
		t.Fatalf("body: got %q, want %q", got, want)
	}
	if got, want := rec.Header().Get("Content-Type"), "text/plain; charset=utf-8"; got != want {
		t.Fatalf("content-type: got %q, want %q", got, want)
	}
}

func TestNewEchoHandler(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name         string
		method       string
		contentType  string
		body         string
		wantStatus   int
		wantInBody   string
		wantEchoBody string
	}{
		{
			name:         "text body",
			method:       http.MethodPost,
			contentType:  "text/plain",
			body:         "hello world",
			wantStatus:   http.StatusOK,
			wantInBody:   "hello world",
			wantEchoBody: "hello world",
		},
		{
			name:         "empty body",
			method:       http.MethodPost,
			contentType:  "text/plain",
			body:         "",
			wantStatus:   http.StatusOK,
			wantInBody:   "",
			wantEchoBody: "",
		},
		{
			name:         "json body",
			method:       http.MethodPost,
			contentType:  "application/json",
			body:         `{"a":1}`,
			wantStatus:   http.StatusOK,
			wantInBody:   `{"a":1}`,
			wantEchoBody: `{"a":1}`,
		},
		{
			name:         "no content type",
			method:       http.MethodPost,
			contentType:  "",
			body:         "raw",
			wantStatus:   http.StatusOK,
			wantInBody:   "raw",
			wantEchoBody: "raw",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			log := slog.New(slog.NewTextHandler(io.Discard, nil))
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, "/echo", strings.NewReader(tc.body))
			if tc.contentType != "" {
				req.Header.Set("Content-Type", tc.contentType)
			}

			echo.NewEchoHandler(log).ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Fatalf("status: got %d, want %d; body=%s", rec.Code, tc.wantStatus, rec.Body.String())
			}

			if tc.wantStatus != http.StatusOK {
				return
			}

			var got echo.EchoResponse
			if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			want := echo.EchoResponse{
				Length:      len(tc.wantInBody),
				ContentType: tc.contentType,
				Body:        tc.wantEchoBody,
			}
			if diff := cmp.Diff(want, got); diff != "" {
				t.Fatalf("response mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNewEchoHandler_TooLarge(t *testing.T) {
	t.Parallel()

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	big := bytes.Repeat([]byte("a"), echo.MaxBodyBytes+1)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewReader(big))
	req.Header.Set("Content-Type", "text/plain")

	echo.NewEchoHandler(log).ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status: got %d, want %d", rec.Code, http.StatusRequestEntityTooLarge)
	}
}

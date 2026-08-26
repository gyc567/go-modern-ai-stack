# Tutorial: Building Go Services with go-modern-ai-stack

A step-by-step walkthrough of the scaffold. By the end you'll have added a new service, written tests that pass under the race detector, and learned how to make AI agents follow your conventions.

## 1. What is this scaffold?

`go-modern-ai-stack` is an opinionated Go scaffold designed for **AI-agent-assisted development**. The thesis:

> AI agents produce better code when the project's conventions are explicit, executable, and verifiable.

The scaffold achieves this through three mechanisms:

1. **`AGENTS.md`** — a single file that AI agents read before writing code. It contains the project's conventions, anti-patterns, and default choices.
2. **An example service** — `cmd/echo` and `internal/echo` demonstrate every convention. An AI agent can pattern-match against this when generating new code.
3. **A constrained toolbox** — `net/http`, `log/slog`, `database/sql`, no ORM, no config library. Fewer decisions = fewer mistakes = more consistency.

The trade-off is reduced flexibility. You commit to:

- No third-party HTTP framework (chi, gin, echo, fiber).
- No ORM (gorm, ent, sqlx-as-ORM).
- No config library (viper, koanf).
- No logging library (zap, zerolog).
- No dependency injection container (wire, fx).

If your project truly needs any of these, this scaffold is the wrong starting point. If your project can live with stdlib + `go-cmp` for tests, this scaffold will save you weeks of re-litigating decisions.

## 2. Prerequisites

- **Go 1.23 or later.** Verify with `go version`.
- **Familiarity with the Go toolchain.** You should be comfortable with `go build`, `go test`, `go vet`, `gofmt`.
- **An AI agent CLI (recommended).** Codex CLI, Claude Code, or any agent that reads project-level instructions files. Without an AI agent, this scaffold still works — it just doesn't deliver its main value.

Optional but useful:

- **`gh` CLI** — for GitHub operations.
- **`make`** — for the v0.2 Makefile targets.

## 3. Quickstart

### 3.1 Clone and verify

```bash
git clone https://github.com/gyc567/go-modern-ai-stack.git
cd go-modern-ai-stack
go build ./...
go test -race -count=1 ./...
```

You should see:

```
?   	github.com/gyc567/go-modern-ai-stack/cmd/echo	[no test files]
ok  	github.com/gyc567/go-modern-ai-stack/internal/echo	0.5s
```

### 3.2 Run the example service

```bash
go run ./cmd/echo
```

In another terminal:

```bash
curl http://localhost:8080/healthz
# ok
curl -X POST -d 'hello world' http://localhost:8080/echo
# {"length":11,"content_type":"","body":"hello world"}
```

Stop the service with `Ctrl-C`. The server logs "shutdown signal received" and exits cleanly.

### 3.3 What just happened?

- The server started on `:8080` (default; override with `PORT=:9000 go run ./cmd/echo`).
- `GET /healthz` returned 200 with body "ok".
- `POST /echo` read the body, parsed it as raw bytes, and returned a JSON envelope.
- `Ctrl-C` triggered `SIGINT`, which `signal.NotifyContext` converted to context cancellation, which `srv.Shutdown` honored gracefully.

## 4. Project layout

```
go-modern-ai-stack/
├── AGENTS.md                  ← AI agent rules (read first)
├── README.md                  ← Human quickstart
├── LICENSE                    ← MIT
├── docs/
│   ├── tutorial.en.md         ← This tutorial (English)
│   └── tutorial.zh-CN.md      ← Tutorial (Chinese)
├── go.mod / go.sum            ← Go module
├── .gitignore
├── cmd/
│   └── echo/                  ← Service binary entry point
│       └── main.go
├── internal/
│   └── echo/                  ← Bounded context: echo handlers
│       ├── handler.go
│       └── handler_test.go
└── testdata/                  ← Fixtures (currently empty)
```

### 4.1 Why this layout?

- **`cmd/<service>/main.go`** — one binary per service. The entry point is obvious from the directory tree.
- **`internal/<bounded-context>/`** — Go's compiler enforces that `internal/` is not importable from outside the module. Business logic stays scoped to its service.
- **`pkg/`** — empty by default. Add code here only when it's reused across two or more services. Resist the urge to use it as a "utility junk drawer" — that leads to a tangled dependency graph.
- **`testdata/`** — fixtures, golden files, large test inputs. The Go toolchain skips this directory during compilation.

### 4.2 Naming

- **Package name = directory name**, lowercase, singular, no underscores. `userservice`, not `userService` or `user_service`.
- **File names**: lowercase, snake_case. `user_handler.go`, not `UserHandler.go`.
- **Avoid stutter**: `userservice.User`, not `userservice.UserService`.

## 5. Walkthrough: the echo service

Read the example in this order: `cmd/echo/main.go`, then `internal/echo/handler.go`, then `internal/echo/handler_test.go`.

### 5.1 `cmd/echo/main.go` — the binary shape

Every service binary follows this skeleton:

```go
func main() {
    if err := run(); err != nil {
        slog.Error("...", "err", err)
        os.Exit(1)
    }
}

func run() error {
    log := slog.New(slog.NewJSONHandler(os.Stdout, ...))
    slog.SetDefault(log)

    // Build mux, server, etc.
    mux := http.NewServeMux()
    mux.Handle("GET /healthz", echo.NewHealthHandler())
    mux.Handle("POST /echo", echo.NewEchoHandler(log))

    srv := &http.Server{
        Addr:         port,
        Handler:      mux,
        ReadTimeout:  defaultReadTimeout,
        WriteTimeout: defaultWriteTimeout,
    }

    // Graceful shutdown via signal.NotifyContext
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
    defer stop()

    errCh := make(chan error, 1)
    go func() {
        log.InfoContext(ctx, "service listening", "addr", srv.Addr)
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
    return nil
}
```

Three invariants this enforces:

1. **`main()` is the only place that calls `os.Exit`**. This makes the binary testable as a function (you can call `run()` from a test in theory).
2. **`run()` returns an error** instead of panicking or calling `os.Exit` directly. This separates "exit logic" from "shutdown logic".
3. **Graceful shutdown is wired through `signal.NotifyContext`**. The shutdown timeout (default 10s) is bounded — if the server doesn't drain in time, `srv.Shutdown` returns an error.

### 5.2 `internal/echo/handler.go` — handler patterns

Two handlers: `/healthz` (cheap liveness check) and `/echo` (the feature).

**Handler constructor takes dependencies explicitly**:

```go
func NewEchoHandler(log *slog.Logger) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // ...
    })
}
```

No globals. The logger is passed in. In tests, you can pass a discarding logger.

**Context flows from request**:

```go
ctx := r.Context()
log.InfoContext(ctx, "echo served", "length", len(body))
```

Every log line carries the request context, which propagates trace IDs, deadlines, and cancellation signals.

**Errors are wrapped with operation context**:

```go
if err != nil {
    status := http.StatusBadRequest
    var maxBytesErr *http.MaxBytesError
    if errors.As(err, &maxBytesErr) {
        status = http.StatusRequestEntityTooLarge
    }
    respondError(ctx, w, log, status, fmt.Errorf("read body: %w", err))
    return
}
```

`%w` (not `%v`) preserves the original error for `errors.Is` and `errors.As`. This matters for retry logic, HTTP error-to-status mapping, and surfacing underlying causes in logs.

**Body size is bounded**:

```go
body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, MaxBodyBytes))
```

Without this, a malicious or buggy client could send a 10GB body and exhaust your memory. `MaxBodyBytes = 1 << 20` caps it at 1 MiB.

### 5.3 `internal/echo/handler_test.go` — test patterns

Three test functions:

1. `TestNewHealthHandler` — single-shot test for the trivial handler.
2. `TestNewEchoHandler` — table-driven test for the echo feature.
3. `TestNewEchoHandler_TooLarge` — boundary test for the body size limit.

**Table-driven pattern**:

```go
cases := []struct {
    name         string
    method       string
    contentType  string
    body         string
    wantStatus   int
    wantInBody   string
    wantEchoBody string
}{
    {name: "text body", method: http.MethodPost, contentType: "text/plain", body: "hello world", ...},
    {name: "empty body", method: http.MethodPost, contentType: "text/plain", body: "", ...},
    {name: "json body", method: http.MethodPost, contentType: "application/json", body: `{"a":1}`, ...},
    {name: "no content type", method: http.MethodPost, contentType: "", body: "raw", ...},
}

for _, tc := range cases {
    t.Run(tc.name, func(t *testing.T) {
        t.Parallel()
        // ... use tc ...
    })
}
```

**`t.Parallel()` on outer AND inner tests**: enables concurrent test execution.

**Use `github.com/google/go-cmp` for deep comparison**:

```go
want := echo.EchoResponse{Length: 11, ContentType: "text/plain", Body: "hello world"}
if diff := cmp.Diff(want, got); diff != "" {
    t.Fatalf("mismatch (-want +got):\n%s", diff)
}
```

`cmp.Diff` produces human-readable output showing exactly which fields differ.

**Discard logs in tests**:

```go
log := slog.New(slog.NewTextHandler(io.Discard, nil))
```

If you don't do this, every test run floods your terminal with log lines.

## 6. Building your first service

Let's add a `greeter` service that responds to `GET /hello/{name}` with `"Hello, {name}!"`.

### 6.1 Plan the bounded context

The bounded context is `greeter`. All greeter code goes in `internal/greeter/`. If `greeter` grows to need multiple sub-domains (e.g., `greeting`, `audience`), you split them: `internal/greeting/`, `internal/audience/`.

### 6.2 Create `internal/greeter/handler.go`

```go
// Package greeter implements the /hello/{name} HTTP handler.
package greeter

import (
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "log/slog"
    "net/http"
)

// Greeting is the JSON response body for GET /hello/{name}.
type Greeting struct {
    Name    string `json:"name"`
    Message string `json:"message"`
}

// errEmptyName is a sentinel for missing path values.
var errEmptyName = errors.New("empty name")

// NewGreetingHandler returns a handler that responds with a personalized greeting.
func NewGreetingHandler(log *slog.Logger) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        name := r.PathValue("name")
        if name == "" {
            respondError(r.Context(), w, log, http.StatusBadRequest,
                fmt.Errorf("name required: %w", errEmptyName))
            return
        }

        msg := fmt.Sprintf("Hello, %s!", name)
        log.InfoContext(r.Context(), "greeting served", "name", name)

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        if err := json.NewEncoder(w).Encode(Greeting{Name: name, Message: msg}); err != nil {
            log.ErrorContext(r.Context(), "encode response", "err", fmt.Errorf("encode: %w", err))
            return
        }
    })
}

func respondError(ctx context.Context, w http.ResponseWriter, log *slog.Logger, code int, err error) {
    log.WarnContext(ctx, "handler error", "code", code, "err", err)
    http.Error(w, http.StatusText(code), code)
}
```

### 6.3 Create `internal/greeter/handler_test.go`

```go
package greeter_test

import (
    "encoding/json"
    "io"
    "log/slog"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/google/go-cmp/cmp"

    "github.com/gyc567/go-modern-ai-stack/internal/greeter"
)

func TestNewGreetingHandler(t *testing.T) {
    t.Parallel()

    cases := []struct {
        name       string
        path       string
        wantStatus int
        want       greeter.Greeting
    }{
        {
            name:       "normal",
            path:       "/hello/Alice",
            wantStatus: http.StatusOK,
            want:       greeter.Greeting{Name: "Alice", Message: "Hello, Alice!"},
        },
        {
            name:       "encoded space",
            path:       "/hello/Bob%20Smith",
            wantStatus: http.StatusOK,
            want:       greeter.Greeting{Name: "Bob Smith", Message: "Hello, Bob Smith!"},
        },
        {
            name:       "empty path segment",
            path:       "/hello/",
            wantStatus: http.StatusBadRequest,
        },
    }

    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            t.Parallel()

            log := slog.New(slog.NewTextHandler(io.Discard, nil))
            rec := httptest.NewRecorder()
            req := httptest.NewRequest(http.MethodGet, tc.path, nil)

            greeter.NewGreetingHandler(log).ServeHTTP(rec, req)

            if rec.Code != tc.wantStatus {
                t.Fatalf("status: got %d, want %d; body=%s",
                    rec.Code, tc.wantStatus, rec.Body.String())
            }

            if tc.wantStatus != http.StatusOK {
                return
            }

            var got greeter.Greeting
            if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
                t.Fatalf("decode response: %v", err)
            }
            if diff := cmp.Diff(tc.want, got); diff != "" {
                t.Fatalf("response mismatch (-want +got):\n%s", diff)
            }
        })
    }
}
```

### 6.4 Verify

```bash
go build ./...
go test -race -count=1 ./...
go vet ./...
gofmt -l .
```

All four commands should pass. If they don't, see the troubleshooting section.

### 6.5 Wire it into a binary

Create `cmd/greeter/main.go` — copy `cmd/echo/main.go` and change the route registration:

```go
mux.Handle("GET /hello/{name}", greeter.NewGreetingHandler(log))
```

Build and run:

```bash
go run ./cmd/greeter
# In another terminal:
curl http://localhost:8080/hello/World
# {"name":"World","message":"Hello, World!"}
```

## 7. Working with AI agents

This is the core value proposition of the scaffold.

### 7.1 Point the agent at AGENTS.md

Most modern AI agents (Codex CLI, Claude Code, Cursor, Aider, Cline) read a project-level instructions file. The convention `AGENTS.md` is widely supported.

Open your agent and prompt:

> "Add a new service at `internal/todo/` that manages a list of TODO items in memory. The service should expose `POST /todos` (create), `GET /todos` (list), and `DELETE /todos/{id}` (delete). Follow AGENTS.md strictly."

### 7.2 Verify the agent's output

After the agent writes code, run the verification loop:

```bash
go build ./...
go test -race -count=1 ./...
go vet ./...
gofmt -l .
```

If any command fails:

- **Read the error.** The compiler and linter are deterministic; the error tells you exactly what's wrong.
- **Identify the rule the agent missed.** Cross-reference with AGENTS.md.
- **Fix or reject.** You can either fix it yourself (faster, sets an example) or ask the agent to fix it (teaches the agent).

### 7.3 Common mistakes AI agents make

Even with AGENTS.md, agents will:

1. **Forget to wrap errors**: `return err` instead of `return fmt.Errorf("op: %w", err)`.
2. **Drop ctx in handlers**: use bare `log.Info` instead of `log.InfoContext(ctx, ...)`.
3. **Use `time.Sleep` in tests**: should use polling with a deadline.
4. **Introduce a config library**: viper, koanf, envconfig.
5. **Use `panic` in business paths**: should return errors.
6. **Forget `t.Parallel()`**: on outer or inner tests.
7. **Use `reflect.DeepEqual`**: should use `cmp.Diff` from go-cmp.
8. **Create a `pkg/utils/`**: should put code in `internal/<bounded-context>/`.

When you spot a **recurring** mistake, **add a rule to AGENTS.md with a BAD/GOOD example**. This converts one-time correction into permanent guidance for future sessions.

### 7.4 Building agent-friendly feedback loops

The pattern that works:

```
1. Prompt the agent with a clear task.
2. Run the verification commands.
3. If something fails:
   a. Identify which AGENTS.md rule was missed.
   b. Ask the agent to fix using that rule's BAD/GOOD example as reference.
   c. If the agent makes the same mistake again, escalate by adding the rule to AGENTS.md (more explicit, with another BAD/GOOD).
4. After 2-3 successful generations of similar code, the agent has internalized the convention.
```

This is slower than "fix it yourself" but builds durable agent behavior.

## 8. Conventions in depth

This section expands on AGENTS.md with the rationale and common pitfalls.

### 8.1 Errors: always wrap

```go
// BAD: caller has no idea what failed
if err != nil {
    return err
}

// GOOD: caller sees the operation context
if err != nil {
    return fmt.Errorf("read %s: %w", path, err)
}
```

`%w` (not `%v`) is critical — it lets `errors.Is` and `errors.As` unwrap the original error. This matters for retry logic, HTTP error-to-status mapping, and surfacing underlying causes.

### 8.2 Context: first parameter

```go
// BAD: caller cannot cancel
func Load() (Config, error) { ... }

// GOOD: caller can cancel, set deadline, or attach values
func Load(ctx context.Context) (Config, error) { ... }
```

In handlers, the context is `r.Context()`. Never create a new one with `context.Background()` inside a handler — you'd lose the request's deadline and cancellation.

### 8.3 Goroutines: errgroup

```go
// BAD: silent failure, no cancellation
go func() {
    work()
}()
go func() {
    moreWork()
}()

// GOOD: errors propagate, first failure cancels the rest
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error { return work(ctx) })
g.Go(func() error { return moreWork(ctx) })
if err := g.Wait(); err != nil {
    return err
}
```

Use `errgroup` whenever you have multiple goroutines that should fail together. Don't use raw channels for cancellation signaling.

### 8.4 HTTP: stdlib ServeMux

```go
// BAD: third-party dependency, unfamiliar API to other Go devs
r := chi.NewRouter()

// GOOD: stdlib, explicit, sufficient for most services
mux := http.NewServeMux()
mux.HandleFunc("GET /users/{id}", h.GetUser)
mux.HandleFunc("POST /users", h.CreateUser)
```

Go 1.22+ supports method patterns (`"GET /path"`) and path parameters (`{id}`) directly in `http.ServeMux`. This is sufficient for ~90% of services. The remaining 10% that need complex routing, middleware chaining, or router-level abstractions are the exception — and they should be the trigger to revisit this convention.

### 8.5 Logging: slog with context

```go
// BAD: no context propagation, hard to trace
log.Println("hello")

// GOOD: request-scoped, structured, greppable
slog.InfoContext(ctx, "hello", "user_id", id, "request_id", reqID)
```

`slog.InfoContext` automatically extracts trace IDs from the context if you use OpenTelemetry. Even without OTel, every log line carries the request's lifetime — invaluable when debugging "what happened to this request?".

### 8.6 Tests: parallel + race

```bash
go test -race -count=1 ./...
```

Three flags matter:

- **`-race`**: enables the race detector. Catches data races that would otherwise manifest as flaky production behavior.
- **`-count=1`**: forces tests to run without test cache. Important in CI; locally you can drop this.
- **`./...`**: run all tests in the module.

`t.Parallel()` makes individual tests run concurrently. Combined with `-race`, this catches synchronization bugs that only appear under concurrent execution.

## 9. When to deviate from conventions

The conventions are **defaults**, not laws. Break them when:

1. **A specific rule doesn't apply to your domain**. Example: if your service is a WebSocket-heavy real-time API, the `net/http.ServeMux` advice may not fit.
2. **Three or more distinct bounded contexts have hit the same wall**. The "3+" rule documented in AGENTS.md.
3. **You document the deviation in AGENTS.md**. This prevents the next contributor (human or AI) from undoing your work.

Example diff:

```diff
- ### 11. Database
- `database/sql` with `pgx` (Postgres) or `lib/pq`. No ORM.
+ ### 11. Database
+ `database/sql` with `pgx` (Postgres) or `lib/pq`. No ORM **except** in `internal/analytics/` where we use `sqlc` (recorded 2026-08-26: 30+ read queries with joins).
```

The exception is visible, dated, and scoped.

## 10. Troubleshooting

### "build failed: undefined: ..."

You forgot to import a package, or the import path is wrong. Check `go.mod` for the correct module path.

### "test failed: ... got X, want Y"

Read the failing assertion carefully. If it's a structural difference, `cmp.Diff` will show you exactly which fields differ. If it's a status code, the handler may be returning the wrong code (check the AGENTS.md error convention).

### "go vet: ... loop variable captured by func"

You captured a `for` loop variable by reference. In Go 1.22+, loop variables are per-iteration, so this is rare. In older Go, fix by passing the variable as a parameter:

```go
for _, p := range paths {
    p := p // Go <1.22 only
    g.Go(func() error {
        return process(ctx, p)
    })
}
```

### "agent ignored AGENTS.md rule X"

Three options:

1. **Rephrase the rule** — make it more explicit. Add another BAD/GOOD example.
2. **Explicitly cite the rule in your prompt** — "Per AGENTS.md rule 5, this handler must use consumer-side interfaces."
3. **Reject and iterate** — don't accept the code; ask the agent to redo.

## 11. What's next

This is v0.1. Future versions:

- **v0.2**: Makefile, golangci-lint configuration, audit scripts for AI-specific anti-patterns, pre-commit hook.
- **v0.3**: Dev tool version pinning via `tools/tools.go`, coverage gates.

Until then, the workflow is: write code → run verification → update AGENTS.md when you see a recurring mistake.

For more on the design philosophy, see the README and AGENTS.md.

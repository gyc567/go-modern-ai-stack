# AGENTS.md

This file is read by AI agents (Codex CLI, Claude Code) before working in this repository. It defines the conventions an agent must follow. The README.md is written for humans; this file is written for agents.

## Meta

1. Modifications to AGENTS.md require explicit human approval.
2. AGENTS.md has priority over any LLM default behavior.
3. Read this file before running `make verify` (added in v0.2) or before scaffolding a new service.

## Layout

- `cmd/<service>/main.go` — binary entry points (one per service).
- `internal/<bounded-context>/...` — business logic; not importable from outside the module.
- `pkg/` — only for code reused across two or more services. **Empty until then.**
- `testdata/` — fixtures and golden files; not compiled.

## Conventions

### 1. Errors

Always wrap with operation context. Bare `return err` is a bug.

BAD: `return err`
GOOD: `return fmt.Errorf("read %s: %w", path, err)`

### 2. Context

Any function doing I/O or blocking work takes `ctx context.Context` as its first parameter.

- `context.TODO()` is forbidden everywhere.
- `context.Background()` is only allowed in `cmd/main.go` (for `signal.NotifyContext` and shutdown-timeout contexts).

BAD: `func Load() (Config, error)`
GOOD: `func Load(ctx context.Context) (Config, error)`

### 3. Resource cleanup

`defer` must immediately follow acquisition. **`defer` inside a `for` loop is a bug** — wrap the body in a function.

BAD:
```go
for _, p := range paths {
    f, _ := os.Open(p)
    defer f.Close() // accumulates
}
```
GOOD:
```go
for _, p := range paths {
    func() {
        f, _ := os.Open(p)
        defer f.Close()
    }()
}
```

### 4. Goroutines

`go func()` must accept a context. Coordinate with `golang.org/x/sync/errgroup`.

BAD: `go func() { work() }()`
GOOD:
```go
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error { return work(ctx) })
```

### 5. Interfaces

Consumer-side: define interfaces where they are used, not where they are implemented.

BAD: producer defines `type UserStore interface { ... }`
GOOD: consumer defines `type userStore interface { ... }`

### 6. Naming

- Package = directory name, lowercase, singular, no underscores.
- Files: lowercase, snake_case.
- Avoid stutter: `userservice.User`, not `userservice.UserService`.

### 7. Structs over maps

Business data flows as structs with explicit fields. `map[string]any` is forbidden as a data carrier.

### 8. HTTP

Go 1.22+ `net/http.ServeMux` with explicit method patterns. No third-party routers (chi, gin, echo, fiber).

BAD: `r := chi.NewRouter()`
GOOD: `mux.HandleFunc("GET /users/{id}", h.GetUser)`

### 9. Logging

`log/slog`, context-aware. Outside `cmd/`, `fmt.Println` is forbidden.

BAD: `log.Println("hello")`
GOOD: `slog.InfoContext(ctx, "hello", "user_id", id)`

### 10. Configuration

Env vars + one `LoadConfig(ctx) (Config, error)` per service. No viper, koanf, or YAML configs.

### 11. Database

`database/sql` with `pgx` (Postgres) or `lib/pq`. No ORM.

### 12. Tests

Table-driven. `t.Parallel()` on the outer test and each subtest. Baseline: `go test -race -count=1 ./...`. Deep comparison with `github.com/google/go-cmp`, not `reflect.DeepEqual`.

### 13. Test packaging

- Black-box (public API): `<file>_test.go`, package `<name>_test`.
- White-box (unexported): `<file>_internal_test.go`, package `<name>`.

### 14. Dependencies

After `go get` or any `go.mod` change, run `go mod tidy`. Never edit `go.sum` by hand.

### 15. HTTP server lifecycle

`srv.Shutdown(ctx)` must be wired to `signal.NotifyContext`. `os.Exit` is only allowed in `main()`.

## Anti-patterns

Reject these immediately:

- `panic` in business code paths (only allowed in `cmd/main.go` startup failure).
- `time.Sleep` in tests (use polling helpers with deadline).
- `interface{}` as a data carrier.
- Goroutine accessing shared state without a mutex or channel.
- `_ = err` without an explicit `// reason:` comment.
- Functions longer than ~50 lines.
- `init()` functions.

## Defaults (do not change without updating AGENTS.md)

| Concern | Default | Revisit when |
|---|---|---|
| HTTP framework | `net/http` stdlib | 3+ bounded contexts need path-param features stdlib can't express |
| Logger | `log/slog` | 3+ contexts need structured fields stdlib can't express |
| DB access | `database/sql` + `pgx` | 3+ contexts write repetitive SQL builder code |
| Config | env vars + `LoadConfig` | 3+ contexts need hierarchical config |
| Mocking | hand-written fakes in test files | 3+ test files duplicate the same mock |
| HTTP testing | `net/http/httptest` | 3+ tests need TLS / HTTP/2 fixtures |

"3+" means three distinct bounded contexts, not three files in one context.

## What not to introduce

- Custom framework, plugin system, or base package.
- ORM, DI container, or mock generator.
- Third-party HTTP router, config library, or logging library.
- `pkg/` until code is reused across two or more services.

## Tool versions

- Go 1.23+ (see `go.mod`).
- Dev tools (golangci-lint, goimports, go-cmp) pinned in `tools/tools.go` (added in v0.3).

## License

MIT.

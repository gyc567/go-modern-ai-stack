# go-modern-ai-stack

Opinionated Go scaffold tuned for AI-agent-assisted development.

The source of truth for AI agents is `AGENTS.md` — read it first.

## Requirements

- Go 1.23+

## Quickstart

```
go build ./...
go test ./...
```

The full pre-commit gate (`make verify`) and linter wiring land in v0.2.

## Layout

- `cmd/<service>/main.go` — binary entry points
- `internal/<bounded-context>/...` — business logic
- `pkg/` — only for code reused across services (empty until then)
- `testdata/` — fixtures and golden files

See `AGENTS.md` for the full rule set.

## Versions

- v0.1 — scaffold + example echo service + AGENTS.md (this version)
- v0.2 — Makefile + lint + audit + pre-commit hook
- v0.3 — dev-tool version pinning + first real feature service

## License

MIT — see `LICENSE`.

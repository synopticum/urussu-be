# AGENTS.md

Guidance for AI coding agents working on **urussu-be**. Assumes no prior
knowledge of the project.

This file deliberately covers only architecture, conventions and gotchas.
Setup, run, make-target, database and deployment instructions live in
[README.md](README.md) — read it when you need them, don't duplicate them
here.

## Project overview

`urussu-be` is a Go backend service backed by PostgreSQL. It exposes map data
(dots/points of interest on a historical city map, with layers, images,
objects and paths) over HTTP/JSON. The defining architectural traits:

- **The API is defined once, in Protocol Buffers** (`api/urussu/v1/*.proto`).
  From these contracts are generated: Go protobuf/gRPC stubs, grpc-gateway
  REST handlers, and the OpenAPI (swagger 2.0) schema consumed by the
  frontend.
- **There is no standalone gRPC listener.** grpc-gateway handlers call the
  service implementations in-process; only one HTTP server runs (default
  port 8080), serving `/api/*` (REST), `/swagger.json` (the generated
  OpenAPI schema, embedded into the binary via `go:embed`) and `/swagger/`
  (an embedded Swagger UI rendering that schema).

## Technology stack

- **Go 1.26** (module `urussu-be`), `log/slog` for logging.
- **PostgreSQL 16** via `github.com/jackc/pgx/v5` connection pool.
- **grpc-gateway v2** for REST; `google.golang.org/grpc` + `protobuf` for stubs.
- **Buf** for codegen (workspace `buf.yaml`, plugins `buf.gen.yaml`, remote
  deps pinned in `buf.lock`).
- **golang-migrate** for SQL migrations (`migrations/`).

All codegen tools (`buf`, `protoc-gen-*`, `migrate`) are pinned as Go **tool
dependencies** in `go.mod` and invoked via `go tool` — never ask for global
installs. Exception: `go tool` builds `migrate` without build tags, so
migrations run via
`go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate`
(the Makefile already does this).

## Code organization

The full annotated layout is in README → "Project layout". What matters when
editing code:

- `api/urussu/v1/*.proto` — API contracts, **source of truth** for the API.
- `gen/` — generated code, committed but **never edit by hand**; regenerate
  with `make generate`.
- `pkg/third_party/` — vendored proto imports for IDE import resolution
  only; refreshed by `make generate`, not used by buf itself.
- Hand-written code lives only under `cmd/` and `internal/`.
- `cmd/urussu-be/main.go` — wiring: config load, logger, pgx pool, HTTP
  server, graceful shutdown (SIGINT/SIGTERM, 5s).
- `migrations/` — golang-migrate pairs; schema **and seed data** (seed
  content is largely in Russian). Never edit applied migrations — create new
  ones (`make migrate-create name=create_xxx`).

Layered flow: `handlers/grpc` → `service` → `repository/postgres`.
Conventions:

- Each layer defines the interface it consumes (e.g. `grpc.DotsService`,
  `service.DotsRepository`) **on the consumer side** so it can be tested
  with a stub.
- Handlers embed `UnimplementedXxxServiceServer` and map
  `domain.ErrNotFound` → `codes.NotFound`; other errors are logged and
  returned as `codes.Internal`.

## Commands

- `make generate` — regenerate Go code + swagger schema from proto. **Run
  after any proto change** (also re-vendors `pkg/third_party/`).
- `make lint` — `buf lint` (STANDARD ruleset).
- `go build ./... && go vet ./...` — compile/vet; keep both green.
- `go run ./cmd/urussu-be` — run locally against the dockerized DB.

Everything else (run options, compose stacks, migrate-up/down, DB reset,
verification endpoints) is in README → "Running" / "Make targets" /
"Database".

## Adding a new API endpoint

1. Add the RPC to `api/urussu/v1/*.proto` with a `google.api.http` annotation
   (`get: "/api/v1/..."`). Keep the `go_package` option as
   `urussu-be/gen/urussu/v1;urussuv1`.
2. `make generate`.
3. Implement the handler in `internal/handlers/grpc/` following the existing
   layer/interface conventions.
4. Register it in `cmd/urussu-be/main.go` via
   `urussuv1.RegisterXxxServiceHandlerServer(ctx, gwMux, ...)` on the gateway
   mux mounted at `/api/`.

The endpoint then appears in the REST API and `/swagger.json` automatically.

## Configuration

Defaults live in `internal/config/config.env` — the single source consumed by
the binary (`go:embed`), the Makefile (`include`) and docker compose
(`--env-file`); the overridable env vars are listed in README. Gotchas:

- `config.env` must use full-line comments only — Make keeps whitespace
  before inline comments, corrupting values.
- The DB host is deliberately not in `config.env`: `localhost` for local
  runs (URL assembled from `POSTGRES_*` parts), `db` for the compose `app`
  service.
- Config is validated at startup (fail fast) — keep it that way.

## Code style guidelines

- Standard Go (`gofmt`); package-level doc comments explain the package's
  role; comments are in English and explain *why*, not *what*.
- Errors are wrapped with context (`fmt.Errorf("query dot %s: %w", ...)`);
  sentinel errors via `errors.Is` against `domain.ErrNotFound`.
- No extra frameworks — check `go.mod` before assuming a library exists;
  keep dependencies minimal.
- Database schema changes go through new migration files, never by editing
  applied migrations.

## Testing

- **There are currently no tests** in the repo (`go test ./...` finds none).
  The design anticipates them: consumer-side interfaces at every layer are
  meant to be stubbed in unit tests.
- Practical verification today: `go build ./... && go vet ./...`, then run
  the app against the dockerized DB and hit the REST endpoints (see README).

## Security considerations

- Credentials in `config.env` / `docker-compose.yml` (`urussu`/`urussu`) are
  **dev-only**; real secrets belong in a gitignored `.env`, never committed.
- Never log the raw `DATABASE_URL` — use `postgres.RedactURL`.
- SQL is parameterized through pgx placeholders — keep it that way.
- HTTP server sets `ReadHeaderTimeout`; graceful shutdown is implemented —
  preserve both when touching `main.go`.

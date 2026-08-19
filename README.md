# urussu-be

Go backend service backed by PostgreSQL. Exposes data over HTTP/JSON from a
single set of Protocol Buffers contracts (via grpc-gateway, served
in-process — there is no standalone gRPC listener). The OpenAPI
(swagger) schema consumed by the frontend is generated from the same
contracts — the API is defined once, in proto.

Working on the repo with an AI coding agent? See [AGENTS.md](AGENTS.md) for
architecture decisions, code conventions and gotchas.

## Project layout

- `api/urussu/v1/*.proto` — the API contracts (source of truth).
- `buf.yaml` / `buf.gen.yaml` / `buf.lock` — Buf workspace, plugin and
  dependency configuration.
- `gen/` — generated Go code (`*.pb.go`, `*_grpc.pb.go`, `*.pb.gw.go`).
- `gen/openapiv2/` — generated OpenAPI (swagger 2.0) schema, embedded into
  the binary (`embed.go`) and served as `/swagger.json`.
- `internal/handlers/swagger` — serves `/swagger.json` and an embedded
  Swagger UI at `/swagger/` (assets vendored under `ui/`).
- `pkg/third_party/` — vendored proto imports (googleapis, grpc-gateway
  options) so IDEs can resolve them; refreshed by `make generate`.
- `cmd/urussu-be/main.go` — entry point: wiring, HTTP server, graceful shutdown.
- `internal/config` — configuration (defaults from embedded `config.env` +
  env overrides).
- `internal/domain` — core domain models.
- `internal/service` — use-case layer between handlers and repositories.
- `internal/repository/postgres` — PostgreSQL repositories (pgx pool).
- `internal/handlers/grpc` — gRPC service implementations.
- `migrations/` — golang-migrate SQL migrations; the database schema and
  seed data are built by applying them (see `make migrate-up`).
- `Dockerfile` / `docker-compose.yml` — multi-stage build and `db` + `app`
  services.

## Adding a new API

The API contract lives in `api/urussu/v1/*.proto`; the workflow for adding
an endpoint (proto change → `make generate` → handler → registration) is
documented in [AGENTS.md](AGENTS.md#adding-a-new-api-endpoint) so it is
maintained in one place.

## Prerequisites

- Docker (with Compose v2)
- Go 1.26+ (for the local-dev workflow)

No global codegen tools are needed: `buf` and all `protoc-gen-*` plugins are
pinned as Go tool dependencies in `go.mod` and invoked via `go tool`.

## Running

### Option 1: fully dockerized

```sh
docker compose up --build
make migrate-up   # first start only: create schema + seed data
```

### Option 2: dockerized DB + local app (recommended for development)

```sh
docker compose up -d db   # start Postgres in the background
make migrate-up           # first start only: create schema + seed data
go run ./cmd/urussu-be    # run the app on the host
```

### Endpoints

- `GET http://localhost:8080/swagger.json` — generated OpenAPI schema.
- `GET http://localhost:8080/swagger/` — Swagger UI rendering that schema.

## Configuration

Built-in defaults live in `internal/config/config.env` — the single source
shared by the Go binary (embedded via `go:embed`), the Makefile (`include`),
and docker-compose (`--env-file`). The app runs with no configuration at
all. Every option can be overridden via an environment variable:

- `HTTP_PORT` — HTTP port.
- `DATABASE_URL` — PostgreSQL URL.
- `LOG_LEVEL` — `debug` | `info` | `warn` | `error`.
- `CORS_ALLOWED_ORIGINS` — comma-separated origins allowed to call the API
  cross-origin (use when the frontend is served from a different origin);
  empty disables CORS.
- `JWT_SECRET` — HMAC secret signing the JWT access tokens. **Required and
  env-only**: it has no default in `config.env` (secrets never live in the
  repo), so the app refuses to start without it. It must be at least 32
  bytes (HS256 needs a 256-bit key). For local development export any
  throwaway value, e.g. `JWT_SECRET=dev-secret-change-me-dev-secret-change-me`.

With a secret set, all endpoints except `POST /api/v1/auth/register` and
`POST /api/v1/auth/login` require `Authorization: Bearer <token>`; login
returns a JWT with `sub` (user id), `role`, `iat` and `exp` claims.

The Docker image runs on defaults plus env (see `docker-compose.yml`,
which assembles its `DATABASE_URL` from the same `POSTGRES_*` values).
Note that compose is run with `--env-file internal/config/config.env`,
which replaces the default `.env` for variable interpolation — so
`JWT_SECRET` must come from the shell environment, e.g.
`JWT_SECRET=dev-secret-change-me-dev-secret-change-me make docker-up`.

## IDE setup (GoLand / IntelliJ)

Imported protos (`google/api/annotations.proto`,
`protoc-gen-openapiv2/options/annotations.proto`) come from the remote Buf
registry, so the IDE cannot resolve them on its own. `make generate` vendors
them into `pkg/third_party/` (committed to the repo, refreshed on every
`make generate`). To fix import resolution:

1. Run `make generate` (or just pull — the files are committed).
2. In GoLand: **Settings → Languages & Frameworks → Protocol Buffers** and
   add import paths `api` and `pkg/third_party` (project-relative).

## Make targets
- `make generate` — regenerate Go code + swagger schema from proto.
- `make lint` — lint the proto files (`buf lint`).
- `make build` — build the binary to `bin/urussu-be`.
- `make run` — run locally with built-in defaults.
- `make docker-up` / `make docker-down` — manage the compose stack.
- `make migrate-up` / `make migrate-down` — apply / roll back DB migrations.
- `make migrate-create name=create_xxx` — scaffold a new migration pair.
- `make migrate-version` — show the current migration version.

## Database

- **Engine:** PostgreSQL 16 (Alpine image)
- **Credentials (dev only):** user/password/database = `urussu`/`urussu`/`urussu_v2`
- **Tables:** `users`, `layers`, `dots`, `images`, `objects`, `paths` —
  schema and seed data are defined by the SQL migrations in `migrations/`
  (golang-migrate), applied with `make migrate-up`.
- **Persistence:** data lives in the `pgdata` named volume. Migrations are
  applied from the host and are safe to re-run (`migrate` tracks the
  current version in the database itself).

### Resetting the database

```sh
docker compose down -v   # stop containers AND delete the data volume
docker compose up -d db  # fresh, empty urussu_v2 database
make migrate-up          # rebuild schema + seed data
```

### Inspecting the data

```sh
docker compose exec db psql -U urussu -d urussu_v2
```

### Useful commands

```sh
docker compose stop db      # pause the database (data persists)
docker compose start db     # resume it
docker compose logs -f db   # follow Postgres logs
```

## Notes

- Credentials in `docker-compose.yml` are for local development only —
  change them before deploying anywhere.

## Possible improvements

- Add `buf breaking` to CI to protect the contracts.

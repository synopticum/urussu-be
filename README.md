# urussu-be

Go backend service backed by PostgreSQL. Exposes data over HTTP/JSON from a
single set of Protocol Buffers contracts (via grpc-gateway, served
in-process — there is no standalone gRPC listener). The OpenAPI
(swagger) schema consumed by the frontend is generated from the same
contracts — the API is defined once, in proto.

## Project layout

- `api/urussu/v1/*.proto` — the API contracts (source of truth).
- `buf.yaml` / `buf.gen.yaml` / `buf.lock` — Buf workspace, plugin and
  dependency configuration.
- `gen/` — generated Go code (`*.pb.go`, `*_grpc.pb.go`, `*.pb.gw.go`).
- `gen/openapiv2/` — generated OpenAPI (swagger 2.0) schema, served at
  runtime as `/swagger.json`.
- `pkg/third_party/` — vendored proto imports (googleapis, grpc-gateway
  options) so IDEs can resolve them; refreshed by `make generate`.
- `cmd/urussu-be/main.go` — entry point: wiring, HTTP server, graceful shutdown.
- `internal/config` — configuration (built-in defaults + env overrides).
- `internal/domain` — core domain models.
- `internal/service` — use-case layer between handlers and repositories.
- `internal/repository/postgres` — PostgreSQL repositories (pgx pool).
- `internal/handlers/grpc` — gRPC service implementations.
- `migrations/` — golang-migrate SQL migrations; the database schema and
  seed data are built by applying them (see `make migrate-up`).
- `Dockerfile` / `docker-compose.yml` — multi-stage build and `db` + `app`
  services.

## Adding a new API

1. Describe the method in `api/urussu/v1/*.proto` (with a
   `google.api.http` annotation for the REST mapping).
2. Run `make generate` — regenerates Go code and the swagger schema.
3. Implement the handler in `internal/handlers/grpc/`.
4. Register it in `cmd/urussu-be/main.go`
   (`RegisterXxxServiceHandlerServer` on the gateway mux).

The method then appears automatically in the REST API and in
`/swagger.json`.

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

- `GET http://localhost:8080/api/v1/dots?limit=10` — list dots from
  PostgreSQL as JSON (REST via grpc-gateway).
- `GET http://localhost:8080/swagger.json` — generated OpenAPI schema.

## Configuration

Built-in defaults (defined in `internal/config/config.go`) cover local
development, so the app runs with no configuration at all. Every option
can be overridden via an environment variable:

- `HTTP_PORT` — HTTP port (default `8080`).
- `DATABASE_URL` — PostgreSQL URL.
- `SWAGGER_PATH` — path to the generated OpenAPI schema.
- `LOG_LEVEL` — `debug` | `info` | `warn` | `error`.

The Docker image runs on defaults plus env (see `docker-compose.yml`).

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

- Add CORS middleware on the HTTP mux if the frontend is served from a
  different origin.
- Serve a swagger UI (e.g. swagger-ui embedded via `go:embed`) next to the
  raw `/swagger.json`.
- Add `buf breaking` to CI to protect the contracts.

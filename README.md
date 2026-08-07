# urussu-be

Go backend service backed by PostgreSQL. Exposes data over HTTP/JSON from a
single set of Protocol Buffers contracts (via grpc-gateway, served
in-process — there is no standalone gRPC listener). The OpenAPI
(swagger) schema consumed by the frontend is generated from the same
contracts — the API is defined once, in proto.

## Project layout

- `api/urussu-be/v1/*.proto` — the API contracts (source of truth).
- `buf.yaml` / `buf.gen.yaml` / `buf.lock` — Buf workspace, plugin and
  dependency configuration.
- `gen/` — generated Go code (`*.pb.go`, `*_grpc.pb.go`, `*.pb.gw.go`).
- `gen/openapiv2/` — generated OpenAPI (swagger 2.0) schema, served at
  runtime as `/swagger.json`.
- `pkg/third_party/` — vendored proto imports (googleapis, grpc-gateway
  options) so IDEs can resolve them; refreshed by `make generate`.
- `cmd/urussu-be/main.go` — entry point: wiring, HTTP server, graceful shutdown.
- `internal/config` — `config.toml` loading.
- `internal/repository/postgres` — PostgreSQL repositories (pgx pool).
- `internal/delivery/grpc/handler` — gRPC service implementations.
- `config.toml` — runtime configuration.
- `db/init.sql` — PostgreSQL-compatible dump converted from the SQLite source.
  Loaded automatically by the Postgres container on first start.
- `Dockerfile` / `docker-compose.yml` — multi-stage build and `db` + `app`
  services.
- `sqlite/urussu.db` — original data source, read-only; never modified.

## Adding a new API

1. Describe the method in `api/urussu-be/v1/*.proto` (with a
   `google.api.http` annotation for the REST mapping).
2. Run `make generate` — regenerates Go code and the swagger schema.
3. Implement the handler in `internal/delivery/grpc/handler/`.
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
```

### Option 2: dockerized DB + local app (recommended for development)

```sh
docker compose up -d db   # start Postgres in the background
go run ./cmd/urussu-be    # run the app on the host
```

### Endpoints

- `GET http://localhost:8080/api/v1/dots?limit=10` — list dots from
  PostgreSQL as JSON (REST via grpc-gateway).
- `GET http://localhost:8080/swagger.json` — generated OpenAPI schema.

## Configuration

The app is configured via `config.toml` (path can be changed with the
`-config` flag). See `config.toml` for the available options: HTTP
port, database URL, swagger schema path, log level.

For container deployments, `DATABASE_URL` and `HTTP_PORT`
environment variables override the file values. If the default `config.toml`
is missing, built-in defaults are used.

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
- `make run` — run locally with `config.toml`.
- `make docker-up` / `make docker-down` — manage the compose stack.

## Database

- **Engine:** PostgreSQL 16 (Alpine image)
- **Credentials (dev only):** user/password/database = `urussu`/`urussu`/`urussu`
- **Tables:** `collection`, `comments`, `dots`, `objects`, `paths`, `users`
  (schema mirrors the SQLite source, one row per legacy document)
- **Persistence:** data lives in the `pgdata` named volume; `db/init.sql`
  runs only when the volume is empty (first start).

### Resetting the database to a fresh import

```sh
docker compose down -v   # stop containers AND delete the data volume
docker compose up -d db  # init.sql is re-imported on first start
```

### Inspecting the data

```sh
docker compose exec db psql -U urussu -d urussu
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
- The `doc` columns hold JSON as `TEXT`. When building real queries against
  the documents, consider converting them to `JSONB`:
  `ALTER TABLE dots ALTER COLUMN doc TYPE jsonb USING doc::jsonb;`
- To regenerate `db/init.sql` from the SQLite source:

  ```sh
  sqlite3 sqlite/urussu.db .dump | grep -v '^PRAGMA' > db/init.sql
  ```

## Possible improvements

- Extract a `internal/service` layer between handlers and repositories once
  business logic appears (handlers currently call repositories directly).
- Add CORS middleware on the HTTP mux if the frontend is served from a
  different origin.
- Serve a swagger UI (e.g. swagger-ui embedded via `go:embed`) next to the
  raw `/swagger.json`.
- Add `buf breaking` to CI to protect the contracts.

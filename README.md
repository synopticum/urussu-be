# urussu-be

Go backend service backed by PostgreSQL.

## Project layout

- `main.go` — sample app: connects to PostgreSQL and runs read queries
  (per-table row counts + sample dots).
- `db/init.sql` — PostgreSQL-compatible dump converted from the SQLite source.
  Loaded automatically by the Postgres container on first start.
- `Dockerfile` — multi-stage build of the app (Go 1.26 → Alpine).
- `docker-compose.yml` — `db` (postgres:16-alpine) and `app` services.
- `sqlite/urussu.db` — original data source, read-only; never modified.

## Prerequisites

- Docker (with Compose v2)
- Go 1.26+ (only for the local-dev workflow)

## Running

### Option 1: fully dockerized

Runs both the database and the app in containers:

```sh
docker compose up --build
```

The `app` container waits for the `db` healthcheck, connects, prints the
sample query results, and exits with code 0.

### Option 2: dockerized DB + local app (recommended for development)

Run only the database in Docker, run the app directly with Go:

```sh
docker compose up -d db   # start Postgres in the background
go run .                  # run the app on the host
```

This works because:

- `docker-compose.yml` publishes the DB on host port `5432:5432`.
- The app reads the DSN from `DATABASE_URL`; when unset it falls back to
  `postgres://urussu:urussu@localhost:5432/urussu?sslmode=disable`.
  Compose sets `DATABASE_URL` for the containerized app so the same binary
  works in both modes.

To use a different database locally:

```sh
DATABASE_URL="postgres://user:pass@host:5432/dbname?sslmode=disable" go run .
```

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

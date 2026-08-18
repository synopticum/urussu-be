.PHONY: generate lint build run docker-up docker-down

# Shared defaults — single source, also embedded in the binary and passed
# to docker compose via --env-file.
include internal/config/config.env

COMPOSE := docker compose --env-file internal/config/config.env

# Regenerate Go code and the OpenAPI schema from api/**/*.proto.
# Also vendors the imported well-known protos into pkg/third_party/ so IDEs
# can resolve imports (not used by buf itself). Committed to the repo.
generate:
	go tool buf dep update
	go tool buf generate
	go tool buf export buf.build/googleapis/googleapis -o pkg/third_party --path google/api
	go tool buf export buf.build/grpc-ecosystem/grpc-gateway -o pkg/third_party --path protoc-gen-openapiv2/options

lint:
	go tool buf lint

build:
	go build -o bin/urussu-be ./cmd/urussu-be

run:
	go run ./cmd/urussu-be

docker-up:
	$(COMPOSE) up --build

docker-down:
	$(COMPOSE) down

# DB migrations
# NOTE: go tool builds migrate without build tags, so no database drivers get
# compiled in ("unknown driver postgres"). go run -tags postgres builds the
# same CLI with the postgres driver enabled.
DATABASE_URL ?= postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable
MIGRATE := go run -tags postgres github.com/golang-migrate/migrate/v4/cmd/migrate -path migrations -database "$(DATABASE_URL)"

migrate-create:   # usage: make migrate-create name=create_layers
	$(MIGRATE) create -ext sql -dir migrations -seq $(name)

migrate-up:
	$(MIGRATE) up

migrate-down:     # rolls back the single latest migration
	$(MIGRATE) down 1

migrate-down-all:
	$(MIGRATE) down -all

migrate-version:
	$(MIGRATE) version

migrate-goto:     # usage: make migrate-goto version=3
	$(MIGRATE) goto $(version)

migrate-force:    # usage: make migrate-force version=3 (clear dirty state)
	$(MIGRATE) force $(version)

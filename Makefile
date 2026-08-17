.PHONY: generate lint build run docker-up docker-down

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
	go run ./cmd/urussu-be -config config.toml

docker-up:
	docker compose up --build

docker-down:
	docker compose down

# DB migrations
DATABASE_URL ?= postgres://urussu:urussu@localhost:5432/urussu_v2?sslmode=disable
MIGRATE := go tool migrate -path db/migrations -database "$(DATABASE_URL)"

migrate-create:   # usage: make migrate-create name=create_layers
	go tool migrate create -ext sql -dir migrations -seq $(name)

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

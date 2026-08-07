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

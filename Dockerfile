FROM golang:1.26-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/urussu-be

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /server /app/server
COPY config.toml /app/config.toml
COPY gen/openapiv2 /app/gen/openapiv2
ENTRYPOINT ["/app/server", "-config", "/app/config.toml"]

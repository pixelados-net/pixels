# Pixels

[![CI](https://github.com/niflaot/pixels/actions/workflows/ci.yml/badge.svg)](https://github.com/niflaot/pixels/actions/workflows/ci.yml)

Pixels is a fast, idiomatic Go emulator for the pixel protocol. The project is intentionally small at the core, with reusable infrastructure in `pkg/`, realm behavior in `internal/`, packet logic in `networking/`, and controlled plugin-facing APIs in `sdk/`.

## Status

Pixels is being bootstrapped. The current module provides the first package boundaries, documentation rules, and CI checks that compile, vet, test, and enforce coverage.

## Layout

```text
pkg/                    reusable global components
internal/               emulator-only realm features
networking/codec        pixel-protocol frame and payload coding
networking/connection   transport-agnostic sessions and handlers
networking/crypto       cryptographic contracts and implementations
networking/inbound      client-to-server packet decoders
networking/outbound     server-to-client packet encoders
sdk/                    controlled plugin creation surface
```

## Development

Copy `.env.example` to `.env` for local overrides, or use environment variables directly.

Run the emulator:

```sh
go run ./cmd
```

Run the full local check:

```sh
go test ./...
```

Run the CI-equivalent coverage check:

```sh
go test -race -covermode=atomic -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## Configuration

| Variable | Default | Description |
| --- | --- | --- |
| `PIXELS_ENV` | `development` | Runtime environment name. |
| `PIXELS_HOST` | `127.0.0.1` | Protocol listener host. |
| `PIXELS_PORT` | `3000` | Protocol listener port. |
| `PIXELS_ACCESS_KEY` | `pixels-dev-key` | API key for private endpoints through `X-API-Key`. |
| `LOG_LEVEL` | `info` | Zap log level. |
| `LOG_FORMAT` | `console` | Zap encoder, either `console` or `json`. |
| `REDIS_ADDRESS` | `127.0.0.1:6379` | Redis server address. |
| `REDIS_USERNAME` | empty | Redis ACL username. |
| `REDIS_PASSWORD` | empty | Redis password. |
| `REDIS_DATABASE` | `0` | Redis database number. |
| `PIXELS_POSTGRES_HOST` | `localhost` | PostgreSQL server host. |
| `PIXELS_POSTGRES_PORT` | `5432` | PostgreSQL server port. |
| `PIXELS_POSTGRES_DATABASE` | `pixels` | PostgreSQL database name. |
| `PIXELS_POSTGRES_USER` | `pixels` | PostgreSQL user. |
| `PIXELS_POSTGRES_PASSWORD` | `pixels` | PostgreSQL password for development. |
| `PIXELS_POSTGRES_SSL_MODE` | `disable` | PostgreSQL SSL mode. |
| `PIXELS_POSTGRES_MAX_CONNS` | `10` | Maximum PostgreSQL pool connections. |
| `PIXELS_POSTGRES_MIN_CONNS` | `1` | Minimum PostgreSQL pool connections. |
| `PIXELS_POSTGRES_CONNECT_TIMEOUT` | `5s` | PostgreSQL connection timeout. |
| `PIXELS_POSTGRES_STATEMENT_TIMEOUT` | `5s` | PostgreSQL statement timeout. |
| `PIXELS_POSTGRES_HEALTH_TIMEOUT` | `2s` | PostgreSQL health check timeout. |
| `SSO_DEFAULT_TTL` | `5m` | Default one-time SSO ticket lifetime. |
| `SSO_KEY` | `pixels-development-sso-key-change-me` | HMAC key used to derive Redis storage keys for SSO tickets. |
| `SSO_PREFIX` | `pixels:sso` | Redis key prefix for SSO ticket records. |
| `PIXELS_WS_QUEUE_SIZE` | `256` | Maximum queued outbound WebSocket packets per connection. |
| `PIXELS_WS_WRITE_TIMEOUT` | `5s` | Maximum duration for one WebSocket write. |
| `PIXELS_WS_READ_TIMEOUT` | `75s` | Maximum duration for one WebSocket read. |
| `PIXELS_WS_PING_INTERVAL` | `30s` | Interval between server heartbeat pings. |
| `PIXELS_WS_PONG_TIMEOUT` | `60s` | Maximum duration without a client pong before disconnecting. |
| `PIXELS_WS_CLOSE_GRACE` | `2s` | Maximum graceful close flushing duration. |

## Database

Pixels uses PostgreSQL for durable state and Liquibase for schema migrations.
Schema migrations are composed from `database/changelog.xml`, while realm-owned
migrations live with their realm, such as `internal/player/database`.

Run schema validation:

```sh
docker run --rm --network host -v "$PWD:/workspace" -w /workspace liquibase/liquibase:4.31 --defaults-file=database/liquibase.example.properties validate
```

Run schema updates:

```sh
docker run --rm --network host -v "$PWD:/workspace" -w /workspace liquibase/liquibase:4.31 --defaults-file=database/liquibase.example.properties update
```

Seed changelogs are separate from the default schema changelog so development
or test fixtures are never applied accidentally. Run them explicitly with a
context:

```sh
docker run --rm --network host -v "$PWD:/workspace" -w /workspace liquibase/liquibase:4.31 --defaults-file=database/liquibase.seed.example.properties --context-filter=development update
```

## HTTP Surface

- `GET /status` returns public server status.
- `GET /ws` is the public websocket entrypoint.
- `GET /docs` serves Scalar API docs only when `PIXELS_ENV=development`.
- `POST /api/sso/tickets` creates one-time SSO tickets and requires `X-API-Key`.
- Private routes require `X-API-Key: <PIXELS_ACCESS_KEY>`.

## Development Security

When `PIXELS_ENV=development`, connection encryption is optional. Development clients can skip Diffie by not sending the Diffie handshake packets and by sending the SSO ticket over the plain pixel protocol after the normal release or metadata packets.

When `PIXELS_ENV=production`, authentication requires an active secure channel before the SSO ticket is accepted.

## Packet API

Inbound packets are decoded from `codec.Packet` into typed payloads:

```go
payload, err := ticket.Decode(packet)
if err != nil {
	return err
}

_ = payload.Ticket
```

Outbound packets are encoded with typed parameters:

```go
packet, err := status.Encode(true, false, status.WithIsAuthentic(true))
if err != nil {
	return err
}

frame, err := codec.AppendFrame(nil, packet)
```

## Build Metadata

The registered project version lives in `pkg/build`. Release builds can inject the source commit:

```sh
go build -ldflags "-X github.com/niflaot/pixels/pkg/build.CommitHash=$(git rev-parse HEAD)" ./cmd
```

The runtime build version combines the registered version and the first eight characters of the commit hash.
Without `-ldflags`, the commit hash defaults to `dev`, so local `go run ./cmd` prints a version like `0.1.0-dev`.

Project rules for agents and contributors live in `AGENTS.md`.

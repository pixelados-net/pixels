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
| `TOON_CONSOLE` | `false` | Emits each log entry as one inline TOON line for protocol tracing, without timestamp/caller and with normalized packet fields. |
| `PIXELS_I18N_PATH` | `i18n/translations.json` | JSON translation catalog path. |
| `PIXELS_I18N_DEFAULT_LOCALE` | `es` | Default locale used when no player locale is available. |
| `PIXELS_I18N_FALLBACK_LOCALE` | `en` | Fallback locale used when a key is missing in the default locale. |
| `PIXELS_I18N_MISSING_MODE` | `key` | Missing translation behavior, either `key` or `empty`. |
| `PIXELS_CURRENCY_CATALOG_PATH` | `currency/types.json` | JSON catalog of enabled protocol currency types. |
| `PIXELS_CURRENCY_LEDGER_TYPES` | `-1` | Comma-separated currency types whose mutations require audit ledger entries. |
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
| `PIXELS_FURNITURE_TELEPORT_BYPASS_LOCKED` | `false` | Allows paired furniture teleports through locked room modes; room bans still apply. |
| `PIXELS_SUBSCRIPTION_TICK_INTERVAL` | `1m` | Frequency of due-boundary membership lifecycle reconciliation. |
| `PIXELS_SUBSCRIPTION_PAYDAY_INTERVAL` | `744h` | HC payday accounting period; 744 hours equals 31 days. |
| `PIXELS_SUBSCRIPTION_KICKBACK_PERCENTAGE` | `0.10` | Eligible catalog credits returned each payday. |
| `PIXELS_SUBSCRIPTION_PAYDAY_CURRENCY_TYPE` | `-1` | Currency receiving payday rewards; `-1` is credits. |

## HC And VIP

Pixels stores three club levels: `0` for none, `1` for HC, and `2` for VIP.
VIP is the higher tier and is purchased from the same `HC > Habbo Club`
catalog page as HC. Nitro's `vip_buy` layout does not render the product code or
the VIP flag in each row, so the development offers are distinguished by price:
31 days of HC costs 25 credits and 31 days of VIP costs 39 credits. The 90-day
offers are extension deals costing 65 and 99 credits respectively.

An active VIP membership remains VIP when HC time is added; purchasing VIP while
HC is active upgrades the current entitlement. Starting HC after an expired VIP
membership creates a new HC streak and does not retain the expired VIP level.
First membership date, current uninterrupted streak, total club time, and total
VIP time are stored separately.

Payday runs every 31 days by default. It grants the highest applicable streak
bonus plus the configured percentage of eligible catalog credit spending for
that exact period. Complete missed periods are reconstructed after downtime,
and online players receive a newly due payday immediately. Offline rewards are
claimed exactly once at the next authentication. Monthly gifts use complete
31-day accumulated club periods and are independent from prepaid time remaining.

Development seeds provide focused scenarios:

| Player | Scenario |
| --- | --- |
| `demo` | Active HC, one due payday, one eligible linked purchase, and one gift. |
| `alice` | Active VIP with accumulated VIP history and one gift. |
| `bob` | Active HC with two missed paydays and purchases in separate periods. |
| `carol` | Expired historical membership for reactivation testing. |

Use `GET /api/admin/subscriptions/{playerId}` to inspect durable membership,
current payday projection, available gifts, and payday history.

## Database

Pixels uses PostgreSQL for durable state and Liquibase for schema migrations.
Schema migrations are composed from `database/changelog.xml`, while realm-owned
migrations live with their realm, such as `internal/realm/player/database`.

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
- `GET /client/ui-config.json` serves Nitro's configured currency type extension.
- `GET /client/texts/:locale/ExternalTexts.json` serves localized Nitro currency names.
- `POST /api/sso/tickets` creates one-time SSO tickets and accepts
  `Idempotency-Key` for replay-safe retries.
- `/api/admin/players/{playerId}/punishments` lists and applies global bans,
  mutes, warnings, trade locks, and kicks; `DELETE /api/admin/punishments/{id}`
  revokes through the same centralized sanction engine.
- `/api/admin/moderation/*` exposes the issue queue, call-for-help topics,
  moderator presets, and sanction ladder; every operation is documented in `/docs`.
- `POST /api/admin/players` atomically creates a player, profile, and default
  permission assignment; `Idempotency-Key` is required.
- `GET /api/admin/players/by-username/{username}` finds a player by canonical
  username.
- `GET`, `PATCH`, and `DELETE /api/admin/players/{id}` read, update, or
  soft-delete a player. Reads expose an `ETag` for conditional requests.
- `GET /api/admin/currencies/wallet?playerId={id}` reads a player's configured wallet.
- `POST /api/admin/currencies/grant` grants currency.
- `POST /api/admin/currencies/deduct` deducts currency.
- `POST /api/admin/currencies/set` replaces a balance.
- `GET /api/admin/currencies/types` lists configured currency types.
- `POST /api/admin/notifications/send` sends a localized bubble or alert.
- Private routes require `X-API-Key: <PIXELS_ACCESS_KEY>`.

Currency mutation bodies accept `playerId`, `currencyType`, `amount`, optional
`reason`, optional `locale`, and `alert`. Alert delivery defaults to `false`.
When `alert` is `true` and the player is online, Pixels sends a localized
generic alert after committing the balance. The response distinguishes
`alertRequested` from `alertSent`, so an offline player can still receive the
persistent mutation without reporting a false failure.

The two `/client` resources are public and allow cross-origin reads. Add
`http://127.0.0.1:8080/client/ui-config.json` to Nitro's `config.urls`, and add
the desired locale URL, such as
`http://127.0.0.1:8080/client/texts/es/ExternalTexts.json`, to its
`external.texts.url` list. They are partial configuration documents and can be
loaded after Nitro's normal files.

Currency balances are sent during authentication and when Nitro requests packet
`273`. Credits use protocol type `-1`; the initial catalog also enables duckets
(`0`) and diamonds (`5`). Apply the schema changelog before running this
bootstrap against an existing database.

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

# Getting Started

This page covers what you need before running Pixels, the minimum configuration to get a working instance up, and how to build and run it as a container. Use [[ENVIRONMENT-VARIABLES]] for every available setting and [[PRODUCTION-SETUP]] before exposing a deployment publicly.

## Dependencies

Pixels itself ships as a single Go binary, but it needs a handful of backing services to actually do anything:

| Dependency | Used for | Notes |
|---|---|---|
| **PostgreSQL** | Durable state: players, rooms, furniture, groups, and every other realm's persistent data | CI runs against PostgreSQL 17; anything reasonably recent should work |
| **Redis** | Sessions, SSO tickets, and other short-lived shared state | Any Redis-compatible server works |
| **S3-compatible object storage** | User-generated images, such as photos taken with the in-room camera | MinIO is a natural self-hosted choice, but any S3-compatible provider works |
| **Liquibase** | Applying database schema migrations | Run through the official `liquibase/liquibase` Docker image; no local install needed |
| **Go 1.26 or newer** | Building from source | Only needed if you're not using the container image |
| **Docker** | Building or running the container image | Optional if you build and run the binary directly |

Nothing else is required to start the server. Individual realms enable or disable optional behavior through their own configuration, all of it documented with defaults in [`.env.example`](https://github.com/niflaot/pixels/blob/main/.env.example).

## Minimum configuration

Pixels reads configuration from environment variables, and every single one has a working default so the server boots out of the box for local development. Copy `.env.example` to `.env` to get all of them at once. For anything beyond a laptop, though, a short list of them stop being safe to leave at their default:

| Variable | Why it matters |
|---|---|
| `PIXELS_ENV` | Set to `production` outside local development; this also tightens connection security requirements |
| `PIXELS_ACCESS_KEY` | Protects every private HTTP route behind `X-API-Key`; the default is a published development value |
| `PIXELS_HOST` / `PIXELS_PORT` | Where the protocol listener and HTTP server bind |
| `PIXELS_POSTGRES_HOST`, `_PORT`, `_DATABASE`, `_USER`, `_PASSWORD` | Your actual PostgreSQL connection |
| `REDIS_ADDRESS`, `REDIS_PASSWORD` | Your actual Redis connection |
| `SSO_KEY` | HMAC key used to derive SSO ticket storage keys; the default is a published development value |
| `STORAGE_ENDPOINT`, `STORAGE_ACCESS_KEY`, `STORAGE_SECRET_KEY`, `STORAGE_BUCKET`, `STORAGE_PUBLIC_BASE_URL` | Your object storage credentials and bucket; there is no safe default for credentials, so the camera realm won't work correctly until these are set |
| `PIXELS_FIGURE_DATA_PATH` | Points at a figure data file used for avatar validation; the bundled default points at a development fixture and should be overridden with your own file |

Everything else, from chat flood thresholds to subscription payday intervals, has a sensible default and can be left alone until you have a specific reason to change it.

## Preparing the database

Schema and seed data are managed with Liquibase, run through its official Docker image so you don't need a local install. Validate and apply the schema before the first run:

```sh
docker run --rm --network host -v "$PWD:/workspace" -w /workspace liquibase/liquibase:4.31 \
  --defaults-file=database/liquibase.example.properties validate

docker run --rm --network host -v "$PWD:/workspace" -w /workspace liquibase/liquibase:4.31 \
  --defaults-file=database/liquibase.example.properties update
```

Seed data lives in a separate changelog on purpose, so fixtures are never applied to a database by accident. Seeds are tagged with a context, and you opt in explicitly:

```sh
docker run --rm --network host -v "$PWD:/workspace" -w /workspace liquibase/liquibase:4.31 \
  --defaults-file=database/liquibase.seed.example.properties --context-filter=development update
```

The `development` seed gives you a fully working local hotel: four players (`demo`, `alice`, `bob`, `carol`, with IDs 1 through 4 and different permission levels), rooms to walk around in, a stocked catalog, furniture, and fixtures for every implemented realm. Seeds are idempotent, so running the command again after pulling new changes only applies what's new and never duplicates rows. Skip the seed step entirely for a production database; it exists for development and QA.

## Running from source

```sh
go run ./cmd
```

## Your first session

Clients don't log in with a username and password against Pixels directly. Whatever sits in front of the server (a CMS, or you with `curl` during development) creates a one-time SSO ticket for a player, hands that ticket to the client, and the client presents it over the protocol connection. To create a ticket for the seeded `demo` player:

```sh
curl -X POST http://127.0.0.1:3000/api/sso/tickets \
  -H "X-API-Key: pixels-development-access-key-change-me" \
  -H "Content-Type: application/json" \
  -d '{"playerId": 1}'
```

The response contains the ticket and its expiry:

```json
{"ticket": "…", "expiresAt": "2026-01-01T00:05:00Z"}
```

Tickets are single-use and expire after five minutes by default (`ttlSeconds` in the request body overrides that per ticket). Configure your Nitro client's `sso.ticket` with the returned value and point its socket URL at `ws://127.0.0.1:3000/ws`, and you'll land in the hotel as `demo`. The endpoint also accepts an `Idempotency-Key` header, so a retried request returns the original ticket instead of minting a second one.

## Running as a container

Pre-built images are published to the GitHub Container Registry on every tagged release:

```sh
docker pull ghcr.io/niflaot/pixels:latest
```

Or build one locally from the repository root:

```sh
docker build -t pixels .
```

The image is a multi-stage build. The first stage compiles with CGO enabled because Go's native plugin loader requires dynamic linking. The Alpine runtime contains only the resulting binary, its runtime libraries, CA certificates, and the i18n catalog. Currency definitions are supplied through `PIXELS_CURRENCY_TYPES`, so no currency JSON asset is copied into the image.

Run it with your configuration passed in as environment variables:

```sh
docker run --rm -p 3000:3000 --env-file .env ghcr.io/niflaot/pixels:latest
```

The container listens on `0.0.0.0:3000` by default. Database migrations are not applied by the image on startup; run them the same way as in local development, against your target database, before pointing a container at it.

The image being available does not by itself make the current security handshake production ready. [[PRODUCTION-SETUP]] documents the remaining secure channel blocker and the required operational checklist.

## Minimum requirements

There's no published hardware benchmark yet, so treat this as a starting point rather than a guarantee. A single vCPU and around 2 GB of RAM is enough to run the Pixels process itself comfortably for a small instance. On top of that, budget storage for whatever your object storage backend needs to hold uploaded images, such as photos taken with the in-room camera, and let your PostgreSQL storage grow as your player base and its data do. Neither of those scales with Pixels' own footprint; they scale with how much your community actually uses the hotel.

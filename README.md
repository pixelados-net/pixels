# Pixels

[![CI](https://github.com/pixelados-net/pixels/actions/workflows/ci.yml/badge.svg)](https://github.com/pixelados-net/pixels/actions/workflows/ci.yml)
[![Package](https://github.com/pixelados-net/pixels/actions/workflows/package.yml/badge.svg?branch=v0.0.1&event=push)](https://github.com/pixelados-net/pixels/actions/workflows/package.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/pixelados-net/pixels)](go.mod)
[![Version](https://img.shields.io/badge/version-v0.0.2-blue)](https://github.com/pixelados-net/pixels/releases)
[![License](https://img.shields.io/github/license/pixelados-net/pixels)](LICENSE)

Pixels is a server implementation of the [Pixel Protocol](https://pixelados-net.github.io/pixel-protocol), written in Go. The Pixel Protocol is an open specification for the network protocol, room engine, and catalog model spoken by [Nitro](https://github.com/billsonnn/nitro-react) and the wider family of Nitro-based clients: HTML5 clients that render entirely in the browser over native WebSockets, instead of a Flash client with a bridge sitting in front of it.

Pixels is built from scratch against that specification, with every packet checked directly against what a real Nitro client sends and receives on the wire. It's organized by feature realm under `internal/realm/`, backed by PostgreSQL, Redis, and S3-compatible object storage, and wired together with [`go.uber.org/fx`](https://github.com/uber-go/fx) instead of global state.

## Status

Pixels is under active development. Most of the protocol surface the specification defines is implemented realm by realm. Where the shipped Nitro client has no code path left to trigger a specified packet, Pixels still implements the wire contract but performs no behavior behind it, documented as such rather than left silently missing. CI compiles, vets, tests, and enforces coverage on every change.

## Getting started

The [wiki](https://github.com/pixelados-net/pixels/wiki) is the place to start:

- [Getting Started](https://github.com/pixelados-net/pixels/wiki/GETTING-STARTED) covers dependencies, minimum configuration, seeding the database, issuing your first SSO ticket, and running the server locally or as a container.
- [Architecture](https://github.com/pixelados-net/pixels/wiki/ARCHITECTURE) is a tour of the codebase: what a realm is and what each one does.
- The Architecture Internals pages explain every core concept in depth, from packet handlers and commands to events, projections, configuration, and infrastructure, so you can understand the project before reading a single line of code.

The short version for local development:

```sh
cp .env.example .env
go run ./cmd
```

## Layout

```text
pkg/                    reusable infrastructure with no game logic of its own
internal/               emulator-only realm features
networking/codec        pixel-protocol frame and payload coding
networking/connection   transport-agnostic sessions and handlers
networking/crypto       cryptographic contracts and implementations
networking/inbound      client-to-server packet decoders
networking/outbound     server-to-client packet encoders
sdk/                    controlled plugin and bot creation surface
```

## Development

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

Every environment variable the server reads is documented with its default in [`.env.example`](.env.example). Administrative HTTP routes are documented interactively at `GET /docs` when `PIXELS_ENV=development`.

## Contributing

Contributions are welcome. Read [`CONTRIBUTING.md`](CONTRIBUTING.md) for the workflow and expectations, and [`AGENTS.md`](AGENTS.md) for the project's architectural conventions: package layout rules, code style, testing requirements, and the full index of implemented features. In short: keep changes scoped, add tests with every behavioral change, and keep coverage above 80%.

## License

Pixels is licensed under the [GNU Affero General Public License version 3](LICENSE).

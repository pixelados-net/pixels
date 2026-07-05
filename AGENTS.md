# AGENTS.md

This repository contains Pixels, a fast and idiomatic Go emulator for the pixel protocol. Agents working here should keep the codebase simple, reusable, and easy to extend without hiding behavior behind premature abstractions.

## Project Layout

- `pkg/` contains reusable global components such as storage, Redis, WebSocket helpers, logging, and process utilities.
- `internal/` contains pixel-protocol realm features that are private to this emulator.
- `networking/` contains pixel-protocol packet coding, decoding, framing, and transport logic.
- `sdk/` contains controlled reusable implementations for plugin creation and extension points.
- `cmd/` contains executable entry points.

## Networking Layout

- `networking/codec/` contains two-way wire encoding and decoding helpers.
- `networking/connection/` contains transport-agnostic connection sessions, registries, handler routing, state, security policy, and disconnect reasons.
- `networking/crypto/` contains cryptographic contracts and implementations such as Diffie-Hellman.
- `networking/inbound/` contains client-to-server packet definitions.
- `networking/outbound/` contains server-to-client packet definitions.
- Connection sessions may hold transport and security callbacks, but packet handlers must stay transport-agnostic and security-agnostic.
- Connection sessions must unwrap security before dispatching packets, so handlers never know whether a packet came from encrypted or plain traffic.
- Realm packages own their handlers and commands; connection handler registries only register handlers and route decoded plain packets to them.
- Inbound packet packages decode only; expose `Decode(packet codec.Packet) (Payload, error)` and do not expose packet constructors.
- Outbound packet packages encode only; expose `Encode(...) (codec.Packet, error)` and do not expose packet decoders or public payload structs.
- Outbound required protocol fields must be function parameters, while optional protocol fields must use packet-local option functions such as `WithReason(value)`.
- Inbound decoders must validate the packet header before decoding the payload.
- Packet definitions must not create `c2s` or `s2c` package directories.
- Keep exactly one packet per packet package, with `packet.go` and `packet_test.go`.
- Prefer direct functional package names such as `networking/outbound/session/hotel/availability`.
- Avoid one-word ladder paths for packet names, such as `closed/and/opens/at`.
- When packets share a functional concept, classify them by inbound or outbound and use a concise package name.

## Package Rules

- Prefer small, single-name packages with nested paths over long package names.
- Use `networking/session/ping/packet.go` and `networking/session/ping/packet_test.go` instead of names like `networking/session/pingpacket.go`.
- Keep each package focused on one responsibility.
- Keep each file focused on one responsibility.
- Keep every Go source file at or below 250 lines.
- Markdown planning and documentation files are exempt from the 250-line limit, but should still stay structured and easy to scan.
- Keep each package to a maximum of six file pairs, where `hello.go` plus `hello_test.go` counts as one pair.
- If a package needs more tests after six file pairs, create a `tests/` folder inside that package.

## Go Style

- Write Go the Go way: clear data flow, small functions, explicit errors, goroutines where concurrency is natural, and channels only where they clarify ownership.
- Document every package, function, method, struct, interface, type, const, var, field, private field, and test helper in Go doc style.
- Do not add comments inside function bodies.
- Avoid unnecessary interfaces. Introduce an interface only when it decouples a real boundary, supports multiple implementations, or enables focused tests.
- Prefer composition over inheritance-like hierarchies.
- Keep public APIs conservative and stable.
- Keep private APIs readable enough that new contributors can follow them quickly.

## Configuration

- Keep configuration holders close to the component they configure.
- Give each configuration holder its own `config.go`, such as `redis/config.go`, `storage/config.go`, or `logger/config.go`.
- Keep protocol host and port in `pkg/config/app/config.go` unless a concrete transport boundary needs its own settings.
- Do not create long registry-style configuration structs with every setting in the project.
- Compose small configuration structs into application-level configuration only where dependency injection or startup needs a single value.
- Every configuration field must have a default value.
- Keep `.env.example` documented whenever configuration variables are added, removed, or renamed.
- Load local `.env` files for development, but keep environment variables as the source of truth.

## HTTP Routes

- Document every HTTP route in `pkg/http/openapi`.
- Include request headers, request bodies, responses, possible error codes, and response bodies.
- Keep `/status`, `/ws`, and development-only documentation routes public.
- Protect private routes with the configured API key header.

## Testing

- All code must maintain more than 80% test coverage.
- Add tests with every behavioral change.
- Keep tests focused on behavior, not implementation details.
- Prefer table-driven tests when cases share the same setup.
- CI must compile and test the full module before changes are considered ready.

## SDK Rules

- Treat `sdk/` as a controlled extension surface.
- Ask before adding new SDK APIs, exported types, extension hooks, or compatibility promises.
- Keep SDK additions decoupled from realm internals.
- Prefer explicit capability objects and small contracts over broad plugin access.

## Change Discipline

- Keep changes scoped to the requested behavior.
- Do not refactor unrelated code while implementing a feature.
- Split responsibilities before files or packages become hard to scan.
- Preserve the legacy tree unless the task explicitly targets it.

# Contributing to Pixels

Thanks for your interest in contributing. Pixels is a from-scratch Go implementation of the [Pixel Protocol](https://pixelados-net.github.io/pixel-protocol), and it holds itself to a fairly strict set of conventions so the codebase stays readable as it grows. This document covers the workflow; [`AGENTS.md`](AGENTS.md) is the authoritative reference for the conventions themselves.

## Before you start

1. Read the [Getting Started](https://github.com/niflaot/pixels/wiki/GETTING-STARTED) wiki page and get a local instance running against PostgreSQL, Redis, and (if you're touching the camera realm) object storage.
2. Skim the [Architecture](https://github.com/niflaot/pixels/wiki/ARCHITECTURE) wiki pages. They explain the realm pattern, handlers, commands, events, and projections, which is most of what you need to know before opening any file.
3. Check the [issues](https://github.com/niflaot/pixels/issues) for existing discussion. For anything larger than a bug fix, open an issue first so the approach can be agreed on before you invest time in it.

## Ground rules

These are the rules most likely to come up in review. `AGENTS.md` has the full list.

- **One packet per package.** Inbound packets only decode and live under `networking/inbound/`; outbound packets only encode and live under `networking/outbound/`. Each packet package is `packet.go` plus `packet_test.go`.
- **Realms own their behavior.** A packet handler and the domain logic it triggers belong to the realm that owns the data. Realms depend on each other's exported service interfaces, never on internals.
- **Files stay small.** Go source files stay at or below 250 lines, packages stay focused on one responsibility, and a package holds at most six file pairs before tests move to a `tests/` folder.
- **Everything is documented.** Every exported (and private) package, function, type, and field carries a Go doc comment. No comments inside function bodies.
- **Player-facing text goes through i18n.** Anything a player sees is localized through `pkg/i18n` with a stable namespaced key. Packets serialize already-resolved text.
- **Every config value has a default.** New environment variables are documented in `.env.example` in the same change that introduces them.
- **Protocol claims need evidence.** If you're implementing a packet, its wire shape must match what the real Nitro client sends and receives, not just what the specification says. If the client can't trigger a packet at all, it's implemented as a documented no-op, never as invented gameplay.

## Testing

All code must keep coverage above 80%, and CI enforces it. Before opening a pull request, run the same checks CI runs:

```sh
gofmt -l .                                       # must print nothing
go vet ./...
go test -race -covermode=atomic -coverprofile=coverage.out ./...
go tool cover -func=coverage.out                 # total must stay above 80%
```

If your change touches the database schema, validate and apply the Liquibase changelogs from a clean database before pushing (the commands are in the [Getting Started](https://github.com/niflaot/pixels/wiki/GETTING-STARTED) page). Seeds must be idempotent: applying them twice must not duplicate rows.

Add tests with every behavioral change. Prefer table-driven tests when cases share a setup, and test behavior rather than implementation details.

## Pull request workflow

1. Fork the repository and create a topic branch from `main`.
2. Keep the change scoped to one concern. Don't refactor unrelated code in the same pull request, and don't touch the `legacy/` reference tree unless the change is explicitly about it.
3. Write a clear description: what changed, why, and how you verified it. If the change affects a packet, name the header ID and direction and say what you checked it against.
4. Make sure the checklist below passes.
5. Open the pull request against `main`. CI must be green before review.

### Pull request checklist

- [ ] `gofmt`, `go vet`, and the full race-enabled test suite pass locally.
- [ ] Coverage stays above 80%.
- [ ] New behavior has tests; changed behavior has updated tests.
- [ ] New configuration is documented in `.env.example` with a default.
- [ ] New admin routes are documented in `pkg/http/openapi`.
- [ ] Player-facing text is localized through `pkg/i18n`.
- [ ] Database changes come with Liquibase changelogs, rollbacks, and idempotent seeds.
- [ ] The change doesn't leave a file over 250 lines or a package over six file pairs.

## Documentation and the wiki

The wiki is generated from [`docs/wiki/`](docs/wiki) in this repository and synced automatically on every push to `main`. Never edit the wiki through the GitHub UI; those changes get overwritten. If your change alters behavior that the wiki describes, update the corresponding page under `docs/wiki/` in the same pull request.

## Reporting issues

Bug reports are most useful when they include: what you did, what you expected, what happened instead, the server log around the failure, and, for protocol issues, the packet header ID and direction involved. If you can reproduce the problem against a development seed (`demo`, `alice`, `bob`, `carol`), say so; it makes the report reproducible for everyone else.

Security issues that could affect running instances should not be filed as public issues. Reach out to the maintainers privately first.

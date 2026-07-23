# Pixels

A server implementation of the [Pixel Protocol](https://pixelados-net.github.io/pixel-protocol), written in Go.

> **This wiki is generated automatically** from the [`docs/wiki/`](https://github.com/niflaot/pixels/tree/main/docs/wiki) folder of the repository on every push to `main`. Do not edit pages here: your changes will be overwritten. Open a pull request against `docs/wiki/` instead.

## What is this?

The [Pixel Protocol](https://pixelados-net.github.io/pixel-protocol) is an open specification for the network protocol, room engine and catalog model spoken by [Nitro](https://github.com/billsonnn/nitro-react) and the wider family of Nitro-based clients: HTML5 clients that render entirely in the browser. The specification exists to push that ecosystem forward. It encourages clients that are genuinely HTML friendly and treats WebSockets as the connection's native transport, instead of a Flash client with a TCP-to-WebSocket bridge sitting in front of it.

Pixels is a server for that specification, built from scratch. It doesn't emulate a particular historical server or start from an existing codebase. Every packet is implemented by reading the specification and then checking it against what a real Nitro client actually sends and receives on the wire, so when the two disagree, the client wins.

## Who is this for?

- **Operators** who want to run their own instance on something that starts from a clean, modern foundation instead of a decade of accumulated patches.
- **Contributors** who want to read a Go codebase organized by realm (`internal/realm/<name>`) and by packet (`networking/{inbound,outbound}/<realm>/<action>`), where finding the code that handles a given feature is a matter of following a naming convention rather than digging through history.
- **Protocol implementers**, meaning anyone building or maintaining a Nitro-family client, a tool, or another server, who wants a second, independent, carefully cross-referenced implementation to check their own wire behavior against.

## Design goals

Pixels is built the way a project like this would be designed today, not the way it would have been designed when the protocol it speaks was new.

It's written in Go, so it ships as a single static binary with real concurrency primitives and a standard toolchain instead of hand-rolled tooling. State is split across purpose-built infrastructure: PostgreSQL for durable data with Liquibase-managed migrations, Redis for session and cache state, and S3-compatible object storage (MinIO or any equivalent) for user-generated images. Each of those is a well-understood, independently operable piece of infrastructure rather than something bundled and improvised. Every connection is a WebSocket end to end, with no legacy binary-TCP core hiding behind a bridge.

Internally, each feature area (rooms, catalog, moderation, groups, pets, crafting, and so on) is its own self-contained realm with its own persistence, commands and packet handlers, wired together through [`go.uber.org/fx`](https://github.com/uber-go/fx) instead of global state. Documentation is treated as part of the work, not an afterthought: exported packages and functions carry real doc comments, HTTP admin routes are described in OpenAPI, and any feature large enough to need a decision explains that decision in a design document before it explains the code.

And it's free, open-source software, meant to be read and extended by anyone who wants to.

## Protocol coverage

Pixels aims to implement every packet the Pixel Protocol specification defines, but "the specification defines it" and "the shipped Nitro client can actually trigger it" aren't always the same thing. The specification describes the protocol's full historical surface, and browser clients kept evolving after it was written. A handful of specified packets describe UI or gameplay flows, such as certain interstitial ads, older reward and competition widgets, or a phone-verification flow tied to a client feature that never made it into the HTML5 rewrite, that simply have no code path left in the reference client to send or receive them.

For those packets, Pixels still implements the wire contract exactly as specified. It encodes and decodes them correctly, so the format is never a mystery to anyone inspecting traffic, but performs no behavior behind them. That's a deliberate, documented no-op, not a gap. Inventing gameplay for a packet nothing can ever send would be a worse outcome than admitting plainly that the client-side feature it belonged to no longer exists. Every case like this is tracked and justified on its own, alongside the much larger set of packets that are fully and functionally implemented.

## Getting started

New to the project? Start with **[[GETTING-STARTED]]**: dependencies, minimum configuration, seeding the database, issuing your first SSO ticket, and how to build and run the server, including as a container. Before a real deployment, read **[[PRODUCTION-SETUP]]** and the complete **[[ENVIRONMENT-VARIABLES]]** reference.

Once it's running, **[[ARCHITECTURE]]** walks through how the codebase is organized and what each realm is responsible for. When you're ready to go deeper, the Architecture Internals pages explain every core concept before you read a line of code, and they're written to be read in order:

1. **[[REALM-LAYOUT]]**: how a realm's folder tree is shaped, the record/database split, and migrations.
2. **[[HANDLERS-AND-COMMANDS]]**: the life of an inbound packet, the two dispatch styles, and tick versus immediate work.
3. **[[EVENTS]]**: the event bus, the one-package-per-event convention, and publish-after-commit.
4. **[[PROJECTIONS]]**: outbound packets, the projection layer, and the three delivery paths.
5. **[[CONFIGURATION]]**: the two settings patterns, naming, and environment versus database config.
6. **[[INFRASTRUCTURE]]**: fx wiring, zap logging, PostgreSQL scoped transactions, Redis, object storage, i18n, and HTTP.

From there, six feature-focused sections go deep on the systems players actually touch, each split across a few pages so no single one runs long:

- **Authentication** ([[AUTH-CONNECTIONS]], [[AUTH-HANDSHAKE]], [[AUTH-SECURITY]], [[AUTH-SSO]]): the transport independent connection contract, the state machine, the current status of the Diffie Hellman handshake, and how SSO tickets create a live session.
- **Users** ([[USERS-MODEL]], [[USERS-PROFILE]], [[USERS-BADGES-PERMISSIONS]], [[USERS-PERMISSIONS]]): the durable and live player split, profile and figure data, badges, and the hotel permission resolver.
- **Navigator** ([[NAVIGATOR-BROWSING]], [[NAVIGATOR-ROOM-ADS]], [[NAVIGATOR-CREATION-AND-COMPAT]]): the four browsing tabs behind one search packet, room ad promotions, and room creation.
- **Inventory** ([[INVENTORY-FURNITURE]], [[INVENTORY-WALLET]], [[INVENTORY-COLLECTIONS]]): the furniture inventory model, the currency wallet, and every other owned collection.
- **Furniture** ([[FURNITURE-MODEL]], [[FURNITURE-INTERACTIONS]], [[FURNITURE-ADVANCED]]): the furniture definition model, the full interaction-type catalog, and the standalone subsystems (rollers, teleports, mystery boxes, and more) built on top of it.
- **Rooms** ([[ROOMS-RUNTIME]], [[ROOMS-ENTRY]], [[ROOMS-HEIGHTMAP]], [[ROOMS-ENTITIES]]): the per-room goroutine and tick model, doors and doorbells, the 2.5D heightmap and stacking system, and how players, bots, pets, and furniture share and synchronize that one room.
- **Decoration** ([[DECORATION-SEATING]], [[DECORATION-WALL]], [[DECORATION-AMBIENCE]]): the furniture that acts through movement and the surface system rather than clicks: seats, rollers, teleports, wall items, post-its, room paint, and the mood light.
- **Games** ([[GAMES-OVERVIEW]], [[GAMES-AREA]], [[GAMES-TEAM]]): the shared engine behind the four server-authoritative furniture games, and what makes Battle Banzai, Freeze, Football, and Tag each distinct.
- **Plugins** ([[PLUGINS-OVERVIEW]], [[PLUGINS-CREATING]], [[PLUGINS-LISTENERS]], [[PLUGINS-COMMANDS]], [[PLUGINS-SDK]], [[PLUGINS-DEPLOYMENT]]): native Go plugin architecture, creation, listeners, Brigodier commands, permission integration, SDK capabilities, and compatible deployment.

Beyond the wiki:

- `GET /docs` on a running instance (with `PIXELS_ENV=development`) serves the full interactive OpenAPI reference for every administrative route.
- [`AGENTS.md`](https://github.com/niflaot/pixels/blob/main/AGENTS.md) documents the project's architectural conventions in depth, including package layout rules and the full index of implemented features.
- [`CONTRIBUTING.md`](https://github.com/niflaot/pixels/blob/main/CONTRIBUTING.md) covers the contribution workflow and the pull request checklist.

## Pages

| Category | Pages |
|---|---|
| Getting Started | [[Home]] · [[GETTING-STARTED]] · [[PRODUCTION-SETUP]] · [[ENVIRONMENT-VARIABLES]] |
| Architecture | [[ARCHITECTURE]] |
| Architecture Internals | [[REALM-LAYOUT]] · [[HANDLERS-AND-COMMANDS]] · [[EVENTS]] · [[PROJECTIONS]] · [[CONFIGURATION]] · [[INFRASTRUCTURE]] |
| Authentication | [[AUTH-CONNECTIONS]] · [[AUTH-HANDSHAKE]] · [[AUTH-SECURITY]] · [[AUTH-SSO]] |
| Users | [[USERS-MODEL]] · [[USERS-PROFILE]] · [[USERS-BADGES-PERMISSIONS]] · [[USERS-PERMISSIONS]] |
| Navigator | [[NAVIGATOR-BROWSING]] · [[NAVIGATOR-ROOM-ADS]] · [[NAVIGATOR-CREATION-AND-COMPAT]] |
| Inventory | [[INVENTORY-FURNITURE]] · [[INVENTORY-WALLET]] · [[INVENTORY-COLLECTIONS]] |
| Furniture | [[FURNITURE-MODEL]] · [[FURNITURE-INTERACTIONS]] · [[FURNITURE-ADVANCED]] |
| Rooms | [[ROOMS-RUNTIME]] · [[ROOMS-ENTRY]] · [[ROOMS-HEIGHTMAP]] · [[ROOMS-ENTITIES]] |
| Decoration | [[DECORATION-SEATING]] · [[DECORATION-WALL]] · [[DECORATION-AMBIENCE]] |
| Games | [[GAMES-OVERVIEW]] · [[GAMES-AREA]] · [[GAMES-TEAM]] |
| Plugins | [[PLUGINS-OVERVIEW]] · [[PLUGINS-CREATING]] · [[PLUGINS-LISTENERS]] · [[PLUGINS-COMMANDS]] · [[PLUGINS-SDK]] · [[PLUGINS-DEPLOYMENT]] |

## Links

- [Repository](https://github.com/niflaot/pixels)
- [Issues](https://github.com/niflaot/pixels/issues)
- [Pixel Protocol specification](https://pixelados-net.github.io/pixel-protocol)
- [Nitro client](https://github.com/billsonnn/nitro-react)

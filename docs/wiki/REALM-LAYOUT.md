# Realm Layout

This page explains how a realm's folder tree is organized and why, using real realms from the codebase as examples. If you understand this page, you can open any realm and predict where everything is before you look. For what each realm *does*, see [[ARCHITECTURE]]; this page is about how each one is *shaped*.

## The shape of a realm

Take `internal/realm/crafting`, one of the cleanest examples of the pattern:

```text
internal/realm/crafting/
├── config/            runtime tunables with safe defaults
├── policy/            permission node registrations
├── record/            domain model + persistence contract (no SQL here)
│   ├── model.go       plain structs: Recipe, Ingredient, Altar…
│   ├── store.go       the Store interface a database must satisfy
│   └── errors.go      sentinel errors: ErrRecipeNotFound, ErrIngredients…
├── database/          the PostgreSQL implementation of record.Store
│   ├── repository.go  pool wiring + transaction helpers
│   ├── recipe.go      queries for one aggregate area
│   ├── recycler.go    queries for another
│   ├── changelog.xml  this realm's Liquibase master
│   ├── migrations/    0001_recipes.sql, 0002_discovery.sql, …
│   └── seed/          development fixtures, in their own changelog
├── recipe/            feature slice: crafting recipes
│   ├── service.go     the domain service
│   ├── handlers/      packet handlers for this feature
│   └── events/        events this feature publishes
├── recycler/          feature slice: the recycler
├── exchange/          feature slice: credit-furniture exchange
├── projection/        glue that maps service results to outbound packets
└── module.go          the fx.Module that wires it all together
```

Three things to notice.

**Features are sliced vertically, not by technical role.** There's no realm-wide `handlers/` folder holding every handler and no realm-wide `services/` folder. Instead, `recipe/`, `recycler/` and `exchange/` each own their service, their handlers and their events. When you work on the recycler, everything you need sits in one subtree.

**`record/` and `database/` are deliberately separate.** `record/` defines what the domain needs from persistence: plain structs and a `Store` interface, with zero PostgreSQL in sight. `database/` is the pgx implementation of that interface, plus the migrations that create its tables. Domain services import `record`, never `database`. The realm's `module.go` is the only place the two meet:

```go
// NewStore exposes PostgreSQL persistence through the crafting contract.
func NewStore(repository *craftingdb.Repository) craftingrecord.Store { return repository }
```

**Migrations live with the realm that owns the tables.** Each realm has its own `database/changelog.xml` and `database/seed/changelog.xml`, and the top-level `database/realms/changelog.xml` includes every realm's changelog in dependency order. A migration file is Liquibase's formatted SQL with an explicit rollback:

```sql
--liquibase formatted sql
--changeset pixels:pixels-crafting-0001-recipes
create table crafting_altars (
    definition_id bigint primary key references furniture_definitions(id),
    enabled boolean not null default true,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    version bigint not null default 1 check(version > 0)
);
--rollback drop table if exists crafting_altars;
```

Mutable aggregates carry the same standard columns everywhere: `created_at`, `updated_at`, and a `version` counter used for optimistic concurrency in admin APIs. Seeds are separate changesets tagged `context:development`, written to be idempotent (`on conflict do update` or `do nothing`), so re-applying them never duplicates rows.

## Larger realms split by capability first

`crafting` fits in one screen. `room` doesn't, so it introduces one more level: stable capability names above the feature slices.

```text
internal/realm/room/
├── access/      who can enter: entry rules, doorbell, closed-room modes
├── control/     owner-facing management: rights, settings, moderation, floor plan, votes, audit
├── record/      the durable room record: model, service, bundles
├── runtime/     live rooms: the active-room registry, broadcast, projection
├── world/       the simulation: grid, pathing, furniture, units, Wired, games
└── database/    all of the above's persistence
```

The rule, straight from `AGENTS.md`: organize large realms by stable capability *before* technical role, and put that capability's commands, handlers and events inside the same subtree. `pet` follows the same idea with `presence/`, `care/`, `breeding/`, `equipment/`, `catalog/` and a `runtime/` for its room-tick integration.

## Events get one package each

An event is its own leaf package with a `Name` constant and, when needed, a `Payload` struct. Nothing else. This is the entire file for the exchange-redeemed event:

```go
// Package redeemed contains the committed exchange event.
package redeemed

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies one committed exchange.
const Name bus.Name = "exchange.redeemed"

// Payload stores bounded exchange identifiers and amount.
type Payload struct {
	PlayerID int64
	ItemID   int64
	Credits  int64
}
```

The path tells you everything: `internal/realm/crafting/exchange/events/redeemed/event.go`. Payloads carry IDs and amounts, not whole domain objects; a subscriber that needs more loads it through the owning realm's service. See [[EVENTS]] for how these are published and consumed.

## Packets are not part of the realm tree

Wire encoding lives under `networking/inbound/` and `networking/outbound/`, one small package per packet, and realms *reference* those packages from their handlers. The mapping is one-directional: a packet package never imports realm code, and the realm decides which headers it registers. This keeps the entire protocol surface auditable with a single grep:

```sh
grep -rhoE "Header uint16 = [0-9]+" networking/inbound networking/outbound --include="*.go"
```

That exact command is how the project's own protocol-coverage audits are produced.

## File and package discipline

A few hard rules keep every realm navigable, enforced in review rather than by tooling:

- Go source files stay at or below 250 lines. When a file wants to grow past that, the responsibility is split.
- A package holds at most six file pairs (`thing.go` + `thing_test.go` counts as one). More tests than that move into a `tests/` folder inside the package.
- One packet per package, one event per package, one responsibility per file.
- Every package, type, function and field carries a Go doc comment; there are no comments inside function bodies.

The result is that "where is X?" almost always has a mechanical answer: find the realm, find the capability, find the feature slice, and the file you want is one of a handful sitting right there.

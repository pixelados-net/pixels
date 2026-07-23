# Projections

A projection is the step that turns "what the server decided" into "what each client should see." Domain services return results; projections map those results onto outbound packets and deliver them to the right connections. This page covers outbound packet encoding, the projection layer itself, and the delivery paths.

## Outbound packets: encode only

The outbound side mirrors the inbound convention from [[HANDLERS-AND-COMMANDS]], inverted. Every server-to-client packet is its own package under `networking/outbound/` exposing a `Header` constant and an `Encode` function. Outbound packages never expose decoders or public payload structs: required protocol fields are function parameters, and optional fields use packet-local option functions. The crafting result packet shows the shape:

```go
const Header uint16 = 618

// WithProduct attaches the crafted product on success.
func WithProduct(recipeName string, itemName string) Option { … }

func Encode(success bool, options ...Option) (codec.Packet, error) { … }
```

A caller can't forget a required field (the compiler makes it a parameter) and can't misuse an optional one (the option function documents when it applies). Variable-length lists are encoded as an `int32` count followed by that many records, matching what the client's parser reads field by field. Every packet package carries a golden test asserting the exact byte layout against the real Nitro parser's expectations, so wire drift fails CI instead of failing in someone's browser.

## The projection layer

Between a service result and the outbound packets sits a small mapping layer. Sometimes it's a dedicated package; `internal/realm/crafting/projection` is a minimal real example that turns a committed craft into the inventory updates the client needs:

```go
// Inventory sends committed removals, one addition, and a refresh marker.
func Inventory(ctx context.Context, connection netconn.Context, removed []int64,
	granted furnituremodel.Item, definition furnituremodel.Definition) error {
	for _, itemID := range removed {
		packet, err := outremove.Encode(itemID)
		if err != nil { return err }
		if err = connection.Send(ctx, packet); err != nil { return err }
	}
	packet, err := outadd.Encode(itemRecord(granted, definition))
	// … then a refresh marker
}
```

The room runtime has the largest projection surface in the codebase (`internal/realm/room/runtime/projection`): it maps live world state, including units, furniture, heights, and effects, into the packets a client entering or watching a room receives. The pattern is always the same, though: projections read domain results and encode; they never contain business decisions. If a projection is making a choice about *whether* something happens, that choice belongs in the service.

One consequence of this split shows up all over the design documents: **durable state is authoritative, projections are best-effort**. If a client misses a projection because it disconnected mid-send, reconnecting rebuilds its view from durable state. A lost packet may cost a visual refresh; it can never duplicate a purchase or lose an item.

## Delivery paths

Projections deliver through three distinct paths, and picking the right one is part of designing a feature:

**Reply to the requester.** The handler holds the requesting connection (`netconn.Context`) and sends directly on it. This is the path for request/response flows: your wallet, your search results, your craft outcome.

**Push to a specific player.** When the trigger isn't a request from that player (a moderation ticket was closed, a friend leveled an achievement), the sender resolves the target's live connection by player ID and delivers best-effort. Moderation's runtime context is the canonical helper:

```go
// SendTo delivers a packet to one online player, silently skipping offline targets.
handler.SendTo(ctx, issue.ReporterPlayerID, packet)
```

Offline targets are simply skipped; if the information matters durably, it's already in the database and will be projected at next login.

**Broadcast to a room.** State every occupant must see, such as furniture appearing, an avatar's effect changing, or a game tile locking, goes through the room runtime's broadcast, which fans one encoded packet out to every connection currently in the room. Encoding happens once; only delivery is per-connection. Bulk updates prefer batch packets over loops of singles, which is why, for example, a Battle Banzai flood-fill locks an entire area with one object-data batch packet rather than one packet per tile.

## Localization happens before encoding

Anything a player reads, including alerts, bubbles, and error feedback, is resolved through `pkg/i18n` *before* it reaches an encoder. Packets serialize already-localized text; no packet package knows translation keys exist. The generic alert helper used across the codebase takes an `i18n.Key`, resolves it for the player's locale with a fallback chain, and only then encodes:

```go
message := handler.Translations.Default("user.settings.volume_invalid")
packet, err := outalert.Encode(message)
return connection.Send(ctx, packet)
```

Keys are stable and namespaced (`session.bubble.furniture.no_rights`, `moderation.report.received`), which keeps the translation catalog greppable and the packets dumb.

## Compatibility projections

Some specified packets have no live feature behind them, because the shipped Nitro client can no longer trigger the flow they belonged to (see the protocol coverage section on [[Home]]). Those still get real encoders and golden tests, and a dedicated compatibility handler answers their requests with explicit, parseable empty snapshots: a zero-length list, a `false` flag, a valid header with nothing in it. The client renders an empty state instead of hanging, and the no-op is a documented decision in code rather than a missing handler. When you see a handler that "responds with nothing," check its doc comment; it will tell you why.

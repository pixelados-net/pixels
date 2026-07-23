# The Player Model

First page of the Users section. It explains the single most important idea in how Pixels handles people: the split between the durable player and the live player, and the binding registry that ties both to a connection. [[USERS-PROFILE]] covers the durable state families, [[USERS-BADGES-PERMISSIONS]] covers badges, and [[USERS-PERMISSIONS]] covers hotel authorization.

## Durable player versus live player

Pixels keeps two representations of the same person, deliberately separated:

The **durable player** is the database record and everything reachable from it: identity, profile, settings, wardrobe, badges, currencies, club membership. It exists whether or not the player is online, and it is the only authority. Admin routes (`/api/admin/players/…`) operate on this layer, which is why you can create, edit, or sanction a player who has never connected once.

The **live player** (`internal/realm/player/live`) is an in-memory projection created at login and released at disconnect. It carries hot, frequently-read snapshots (current look, motto, club level, client settings, effect state) so packet handlers and the room engine never touch PostgreSQL to answer "what does this avatar look like right now." A room full of thirty avatars re-renders from memory, not from thirty queries.

The rule that keeps the two in sync: **durable state is written first, then the live snapshot is updated from the committed result.** A crash between the two costs a stale projection until reconnect, never a lost write. The settings handler shows the shape literally: persist through the service, then `player.SetClientSettings(…)` with the record that came back.

```go
record, err := handler.Service.SetOldChat(context.Background(), playerID, payload.OldChat)
if err != nil {
	return err
}
applyRecord(player, record) // live snapshot updated from the committed record
```

## The binding registry

`internal/realm/session/binding` is the small but load-bearing piece connecting a person to a socket:

```go
// Binding maps one authenticated player to one connection.
type Binding struct {
	PlayerID       int64
	ConnectionID   connection.ID
	ConnectionKind connection.Kind
	BoundAt        time.Time
}
```

The registry indexes bindings both ways, player to connection and connection to player, behind an `RWMutex`. Every authenticated packet handler starts by resolving the sender through it, and every "push this to player 42" projection resolves the target connection through it. Handlers use a standard helper shape you'll see across realms:

```go
current, found := handler.Bindings.FindByConnection(binding.ConnectionKey{Kind: connection.ConnectionKind, ID: connection.ConnectionID})
player, found := handler.Players.Find(current.PlayerID)
```

Duplicate bindings are rejected, which is also what makes a second login for the same player an explicit, handled event instead of two ghost sessions fighting over one identity.

## Lifecycle: connect and disconnect

At authentication time the sequence is: resolve the ticket to a player, authenticate the session, register the binding, load the live player, publish `player.connected` on the event bus. Disconnect reverses it and publishes `player.disconnected`.

Those two events are the realm-neutral seam everything else hangs off. Badges warm their equipped-badge snapshot on connect and release it on disconnect. Rooms remove the player's unit. Pets flush dirty state. Moderation refreshes sanction projections. None of those realms are known to the connection code; they subscribed to the events (see [[EVENTS]]).

## What this buys you

Three properties fall out of this model, and the rest of the Users section leans on all of them:

1. **Reconnecting always heals.** Nothing session-scoped exists that can't be rebuilt from durable state, so a dropped WebSocket costs a reconnect, never data.
2. **Offline operations are first-class.** Grants, sanctions, purchases, and admin edits work identically whether the target is online (live projection updates too) or offline (it'll be projected at next login).
3. **Hot paths stay in memory.** The simulation reads snapshots; the database sees writes and cold loads, not per-frame reads.

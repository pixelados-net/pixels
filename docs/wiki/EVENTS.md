# Events

Commands are how behavior gets *requested*; events are how the rest of the system finds out it *happened*. This page covers the event bus, the one-package-per-event convention, and the rules around publishing and subscribing that keep realms decoupled.

## The bus

`pkg/bus` is a small, in-process event bus with prioritized subscribers, provided to the fx graph as two narrow interfaces so components only depend on the half they use:

```go
// Publisher publishes local events.
type Publisher interface {
	Publish(context.Context, Event) error
}

// Subscriber subscribes to local events.
type Subscriber interface {
	Subscribe(Name, int, Handler) (*Subscription, error)
	SubscribeOnce(Name, int, Handler) (*Subscription, error)
}

// Handler handles one local event.
type Handler func(context.Context, Event) error
```

An `Event` is a `Name` plus a payload. Names are stable, namespaced strings owned by the publishing realm: `exchange.redeemed`, `player.connected`, `player.disconnected`, `room.floorplan.saved`. Subscribers register with a priority so ordering between listeners is explicit rather than accidental.

## One event, one package

Just like packets, every event is its own leaf package containing a `Name` constant and a `Payload` struct, nothing more:

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

The package path doubles as documentation: `internal/realm/crafting/exchange/events/redeemed` tells you the realm, the feature, and the fact it's an event before you open the file. Payloads deliberately carry only identifiers and bounded values, never live objects or full domain records. If a subscriber needs the whole entity, it loads it through the owning realm's service. That keeps events cheap to publish, safe to log, and impossible to mutate from the outside.

## Publish after commit, never before

Events announce facts. A fact isn't a fact until the transaction that created it commits, so the pattern everywhere in the codebase is: run the transaction, then publish. Here's the real exchange service:

```go
func (service *Service) Redeem(ctx context.Context, playerID int64, itemID int64) (Result, error) {
	result := Result{RemovedItemID: itemID}
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		// validate ownership, pick up the item, delete it, grant credits…
		return err
	})
	if err == nil && service.events != nil {
		_ = service.events.Publish(context.Background(), bus.Event{
			Name:    redeemedevent.Name,
			Payload: redeemedevent.Payload{PlayerID: playerID, ItemID: itemID, Credits: result.Credits},
		})
	}
	return result, err
}
```

Two details worth noticing. The publish error is intentionally discarded: the durable outcome already committed, and a projection or side effect that fails to fan out must never turn a successful operation into a reported failure. And the publisher is optional (`!= nil`), so services stay constructible in tests without a bus.

## What subscribes to events

Events are the seam that lets unrelated realms react to each other without importing each other. Real examples of the pattern in the codebase:

- **Session lifecycle.** `player.connected` and `player.disconnected` are the realm-neutral way to notice a session starting or ending. The badge service warms its equipped-badge snapshot on connect and releases it on disconnect by subscribing to exactly these, without the connection code knowing badges exist.
- **Cross-realm progression.** The progression engine advances achievements and quests by listening to events other realms already publish (a crafted recipe, a completed trade, a placed furniture item). The publishing realm doesn't know progression exists; it just states what happened.
- **In-realm fan-out.** A realm often subscribes to its own events to keep projection concerns out of its services: the service commits and publishes, a subscriber turns the event into outbound packets for whoever should see it.

Subscriptions are registered at startup, usually from a realm's `module.go` or a dedicated wiring file, via an `fx.Invoke` that receives the `bus.Subscriber`. Handlers on the bus follow the same discipline as packet handlers: do bounded work, and if you need heavy lifting, load through a service rather than dragging state through the payload.

## When to add an event, and when not to

Add an event when something durable happened that code outside the current feature could legitimately care about. Don't add one for request/response flows that begin and end inside a single handler, and don't use events as a substitute for calling a service you're allowed to depend on directly. The test is direction: services are for asking another realm to do something; events are for telling anyone who cares that something already happened. If you find yourself publishing an event and then waiting for its effects, you wanted a service call.

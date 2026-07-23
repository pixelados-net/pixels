# Handlers and Commands

This page follows a packet from the moment it arrives on a WebSocket to the moment domain logic runs, and explains the two dispatch styles you'll meet along the way. Read [[REALM-LAYOUT]] first if you haven't; this page assumes you know what a realm looks like.

## The life of an inbound packet

1. The WebSocket transport reads bytes and `networking/codec` decodes them into frames, each one a `codec.Packet`: a `uint16` header plus an opaque payload.
2. The connection's `Session` (in `networking/connection`) unwraps security first, so by the time anything downstream sees the packet, it's plaintext. Handlers never know or care whether the connection was encrypted.
3. The session looks the header up in its `HandlerRegistry`, a plain `map[uint16]Handler` that every realm populated at startup, and invokes the registered handler.
4. The handler decodes the payload into a typed struct, resolves who sent it, and either calls a domain service directly or dispatches a typed command. Domain logic runs, results are projected back out as outbound packets ([[PROJECTIONS]]).

## Decoding: one packet, one package

Every inbound packet is its own package under `networking/inbound/`, exposing exactly three things: a `Header` constant, a `Payload` struct, and a `Decode` function. Decoders validate the header before touching the payload, and they only decode; there are no packet constructors on the inbound side. This is the real craft packet:

```go
package craft

import "github.com/niflaot/pixels/networking/codec"

const Header uint16 = 3591

type Payload struct {
	AltarItemID int64
	RecipeName  string
}

func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.StringField})
	if err != nil {
		return Payload{}, err
	}
	return Payload{AltarItemID: int64(values[0].Int32), RecipeName: values[1].String}, nil
}
```

`codec.Definition` describes the field layout (`Int32Field`, `StringField`, `BooleanField`, and so on), and `DecodePacketExact` fails if any bytes are missing or left over. Malformed wire data is a protocol error; a *valid* packet asking for something impossible is a domain concern and must never disconnect the client.

## Handler style one: direct handlers

Request/response realms with no ongoing simulation, such as crafting or moderation's call-for-help intake, register plain handler functions. A feature's `handlers` package holds a `Handler` struct with its dependencies and a `Register` function that binds each header to a method:

```go
// Register installs user-facing call-for-help adapters.
func Register(registry *netconn.HandlerRegistry, runtime *moderationruntime.Context) {
	handler := Handler{Context: runtime}
	_ = registry.Register(inreport.Header, handler.callForHelp(false))
	_ = registry.Register(inreportim.Header, handler.callForHelp(true))
	_ = registry.Register(incfhpending.Header, handler.pending)
	_ = registry.Register(incfhdelete.Header, handler.deletePending)
}
```

Each method decodes, resolves the acting player, calls the service, and sends a response. The realm's `module.go` invokes these `Register` functions against the shared registry at startup:

```go
// RegisterConnectionHandlers registers every crafting realm packet adapter.
func RegisterConnectionHandlers(handlers *realmconn.Handlers, recipes *recipehandlers.Handler, ...) {
	recipehandlers.Register(handlers.Inbound, recipes)
	recyclerhandlers.Register(handlers.Inbound, recycler)
	exchangehandlers.Register(handlers.Inbound, exchange)
}
```

Handlers stay thin on purpose. Decoding, actor resolution, and translating domain errors into localized responses is all a handler does; everything with consequences lives in the service it calls.

## Handler style two: typed commands

Realms whose behavior interacts with live, mutable room state (room, pet, furniture interactions) don't call services directly from the network goroutine. They wrap the request in a typed **command** and hand it to a dispatcher. The contract lives in `internal/command`:

```go
// Command describes a typed runtime command.
type Command interface {
	CommandName() Name
}

// Envelope wraps a command with runtime metadata.
type Envelope[T Command] struct {
	Command  T
	Metadata Metadata // PlayerID, ConnectionID, CreatedAt
}

// Handler handles one typed command.
type Handler[T Command] interface {
	Handle(context.Context, Envelope[T]) error
}
```

A command is a plain struct naming an intent (`room.enter`, `navigator.search`, a furniture pickup) plus the data needed to execute it. `command.NewDispatcher` wraps a `Handler[T]` with validation, optional middleware, and structured logging of every dispatch: command name, player ID, connection ID, timestamp. The packet handler's whole job becomes: decode, build the command, dispatch.

Why the extra layer? Three reasons. Commands give every state-changing action a uniform audit trail in the logs. Middleware (permission checks, throttles) composes without touching handler bodies. And commands are how work crosses from "the goroutine this connection runs on" into "the context of a live room" safely.

## Immediate work versus tick work

The room world is a simulation: rollers roll, pets wander, Wired effects fire after delays, game timers count down. That simulation advances on a **tick**, driven by the contracts in `internal/tick`:

```go
// Tick describes one runtime tick.
type Tick struct {
	At       time.Time
	Delta    time.Duration
	Sequence uint64
}

// Target handles ticks.
type Target interface {
	Tick(context.Context, Tick) error
}
```

The dividing line is simple. Anything that must observe or mutate the continuous simulation happens *inside* the tick, scheduled through the room's own runtime. Anything that's a self-contained request, like opening a catalog page, reading your wallet, or crafting an item, executes immediately on the connection's goroutine. This is also a hard rule for feature design: realms never spawn their own goroutines or `time.AfterFunc` timers per entity. A freeze ball that explodes in two seconds is a deadline registered on the room's scheduler, not a timer floating around the runtime. That keeps everything that touches a room serialized through one place, which is why the room engine needs no locks around its world state.

## Where errors go

Handlers distinguish three failure classes, and the distinction matters:

- **Protocol errors** (bad header, malformed payload): returned as errors, which terminates the connection. Only broken clients produce these.
- **Expected domain failures** (recipe sold out, no permission, room full): translated into a localized response through `pkg/i18n` and sent as a normal packet. The session always survives these.
- **Infrastructure failures** (database down): logged with context and surfaced as a generic failure to the player, without leaking internals.

A useful invariant to remember when writing a handler: nothing a *well-behaved* client can do should ever disconnect it.

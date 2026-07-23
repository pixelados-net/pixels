# Transport Agnostic Connections

Pixels separates the Pixel Protocol session from the socket that carries it. WebSocket is the only transport adapter currently shipped, but nothing in packet routing, authentication, realm handlers, session state, or the connection registry depends on Fiber WebSocket types.

## The boundary

`networking/connection.Session` owns protocol state. A transport creates it with a `SessionConfig` and supplies three callbacks:

| Callback | Responsibility |
|---|---|
| `Sender` | Accept one outbound `codec.Packet` and deliver it through the transport |
| `Disposer` | Release transport resources for a typed disconnect reason |
| `SecurityActivator` | Install a secure channel only after earlier queued writes have completed |

The config also supplies a connection id, a transport `Kind`, the remote address, shared inbound and outbound handler registries, a security policy, and an optional packet logger.

```go
session, err := connection.NewSession(connection.SessionConfig{
	ID:                connection.ID(id),
	Kind:              connection.Kind("custom-transport"),
	RemoteAddr:        remoteAddress,
	Inbound:           handlers.Inbound,
	Outbound:          handlers.Outbound,
	SecurityPolicy:    policy,
	Sender:            transport.send,
	Disposer:          transport.dispose,
	SecurityActivator: transport.activate,
})
```

The session never reads a WebSocket frame, chooses a TCP close code, or owns a network library. It receives and sends decoded packets. The adapter owns framing, read and write deadlines, queues, heartbeats, backpressure, and physical closure.

## The WebSocket adapter

`pkg/http/websocket` is one implementation of that boundary. It performs this inbound sequence:

```text
WebSocket binary message
  secure channel Open
  Pixel frame decoding
  Session Receive
  inbound interceptors
  handler policy
  realm handler
```

Outbound traffic takes the reverse path:

```text
Realm creates packet
  Session Send
  outbound handlers
  WebSocket queue
  Pixel frame encoding
  secure channel Seal
  WebSocket binary message
```

One writer goroutine serializes packets, security activation barriers, protocol disconnect packets, and the final WebSocket close frame. This prevents concurrent writes and guarantees that an unencrypted key exchange completion can leave the queue before encryption becomes active.

The reader accepts binary messages only. It keeps incomplete frame bytes between messages, decodes every complete packet, and maps transport, policy, framing, and handler errors to typed disconnect reasons. A heartbeat goroutine sends protocol ping packets and closes sessions that exceed the configured pong timeout.

## Adding another transport

A TCP, QUIC, test, or embedded adapter would follow the same contract:

1. Accept a physical peer and assign a unique id plus a distinct `Kind`.
2. Construct `Session` with the shared handler registries.
3. Register it in `networking/connection.Registry` under that kind.
4. Convert transport bytes into `codec.Packet` values before `Receive`.
5. Implement `Sender` by framing packets for that transport.
6. Implement `Disposer` and map `connection.Reason` to native closure semantics.
7. Serialize writes and provide a security activation barrier when the transport supports in protocol encryption.
8. Remove the session, release player bindings, and publish disconnect lifecycle behavior when the read loop ends.

Connection ids are scoped by kind, so two adapters may use the same textual id without colliding. Administration can list one kind or every active connection through the common registry.

## Handler isolation

Realm handlers receive `connection.Context`, which is an immutable session snapshot plus safe helpers such as `Send`, `Disconnect`, `Transition`, and security validation. They never receive the transport object.

Every handler registration declares allowed lifecycle states and whether authentication is required. The default policy accepts only authenticated `Connected` sessions. Handshake and ticket handlers explicitly opt into early states and unauthenticated traffic. A new transport therefore inherits the same protocol policy without duplicating it.

## Why this matters

Transport failures, protocol failures, authentication failures, moderation kicks, bans, timeouts, and shutdowns all become one stable `connection.Reason`. Player bindings store both connection kind and id. Packet handlers and projections call the `Connection` interface. This lets the emulator add a transport without introducing transport conditionals throughout the realms.

# Connection Lifecycle and Handshake

This page covers what happens between a transport opening and the moment a session is allowed to present credentials: the connection state machine and the handshake packets. Read [[AUTH-CONNECTIONS]] for the transport boundary, continue with [[AUTH-SECURITY]] for byte wrapping, and use [[AUTH-SSO]] for the login itself.

## The connection state machine

Every session (`networking/connection`) moves through an explicit lifecycle, and every packet handler declares which states it accepts. A packet arriving in the wrong state is a protocol error, not a judgment call:

```text
Created → Handshaking → (Securing ⇄ Handshaking) → Authenticating → Authenticated → Connected
```

Handlers register with state guards, so the registration itself is the security model:

```go
_ = registry.Register(inrelease.Header, Release,
	netconn.AllowStates(netconn.StateCreated, netconn.StateHandshaking), netconn.AllowUnauthenticated())
_ = registry.Register(inticket.Header, Ticket(authenticator),
	netconn.AllowStates(netconn.StateHandshaking), netconn.AllowUnauthenticated())
```

Only a handful of packets carry `AllowUnauthenticated`; everything else implicitly requires an authenticated session. That's why no realm handler ever re-checks "is this user logged in": a packet that reaches a realm handler could not have arrived otherwise.

Transitions are validated too: the state machine only permits legal moves (`Created → Handshaking`, `Handshaking → Securing`, back, and forward into authentication), and an illegal transition disconnects with a typed reason instead of leaving the session in an undefined phase.

## The handshake packets

A Nitro client currently opens the WebSocket (`GET /ws`) and sends a short metadata sequence before authenticating. The shared session would accept the same packet sequence from any future transport adapter. Pixels accepts, in the early states:

| Packet | Purpose | What Pixels does |
|---|---|---|
| Release version | Client build identification | Decoded and accepted |
| Client variables | Asset/config URLs the client is using | Decoded and accepted |
| Policy probe | Legacy cross-domain policy request | Decoded and accepted |
| Machine ID | A persistent per-device identifier | Validated; a missing or malformed ID gets a server-issued replacement |
| Diffie init / complete | In-protocol key exchange | See [[AUTH-SECURITY]] |
| SSO ticket | The actual login | See [[AUTH-SSO]] |

The machine ID exchange deserves a note: when the client presents nothing usable, Pixels generates a random hex identifier and sends it back, so every device ends up with a stable ID without any of it being *trusted* for authentication. It's device telemetry, not identity.

## Where handlers for this live

The handshake handlers are deliberately outside any gameplay realm, in `internal/realm/connection/handlers/handshake` and `internal/realm/connection/handlers/security`. They are the only handlers in the codebase allowed to run before authentication. `handshake` owns client metadata and key exchange packets. `security` owns machine identity and the SSO ticket. Neither package imports the WebSocket adapter.

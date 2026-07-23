# Security and Encryption

This page covers the secure channel wrapper, the current state of the Diffie Hellman contract, the per environment security policy, and the configuration values you must never ship with defaults. Read [[AUTH-CONNECTIONS]] for the transport boundary and [[AUTH-HANDSHAKE]] for the states these rules apply in.

## The secure channel wrapper

Security is a session capability rather than a special connection type. `SecureChannel` owns five operations:

```go
type SecureChannel interface {
	State() SecurityState
	Begin(context.Context) error
	Open([]byte) ([]byte, error)
	Seal([]byte) ([]byte, error)
	Close(Reason) error
}
```

When no ready channel is attached, `Session.Open` and `Session.Seal` return the original bytes. Once a channel reports `SecurityReady`, inbound transport bytes pass through `Open` before Pixel frame decoding, and complete outbound frames pass through `Seal` immediately before transport writing. Realm handlers always receive the same decoded packets.

Security activation needs an ordering barrier. `CompleteSecurity` first sends the plaintext completion packet, then asks the transport's `SecurityActivator` to queue channel activation. The WebSocket adapter uses its single writer queue for both operations, so encryption cannot begin before the completion packet is physically written.

## Diffie Hellman current status

The wire protocol defines an in-band Diffie-Hellman negotiation (RSA-signed prime and generator, an encrypted public key exchange), and `networking/crypto/diffie` defines exactly those contracts:

```go
// Provider prepares protocol Diffie-Hellman values.
type Provider interface {
	// Begin returns encrypted prime and generator values.
	Begin(context.Context) (Parameters, error)
	// Complete consumes a client public key and returns server completion values.
	Complete(context.Context, PublicKey) (Result, error)
}
```

What exists today is the contract and the packet decoding, not the negotiation. A client that sends Diffie init or complete is disconnected with a clear protocol error (`diffie unavailable`) rather than strung along with a half-working exchange:

```go
// DiffieInit handles Diffie start packets.
func DiffieInit(context netconn.Context, packet codec.Packet) error {
	if _, err := indiffieinit.Decode(packet); err != nil {
		return err
	}
	return disconnectMissingDiffie(context)
}
```

Nitro clients handle this fine in development because they can be configured to skip the in-protocol handshake entirely and go straight to the SSO ticket.

## The security policy

Enforcement doesn't depend on Diffie being implemented. Each session carries a security policy derived from the environment at connection time:

```go
func SecurityPolicyForEnvironment(environment string) SecurityPolicy {
	if strings.EqualFold(environment, "production") {
		return SecurityPolicy{Mode: SecurityRequired}
	}
	return DefaultSecurityPolicy() // SecurityOptional
}
```

| Mode | Authentication rule |
|---|---|
| `SecurityOptional` | A plaintext session may present an SSO ticket |
| `SecurityRequired` | The session must have a `SecurityReady` channel before the SSO ticket is accepted |

Handlers never see any of this. The session unwraps security before dispatch, so packet handlers are byte for byte identical on plain and secured connections. That separation is a hard architectural rule.

TLS termination in front of the server is still required for a public deployment because it protects the HTTP and WebSocket transport. It does not currently mark `SecureChannel` ready. Since the in protocol Diffie implementation is not complete, `PIXELS_ENV=production` currently rejects the normal Nitro SSO flow even behind a TLS proxy. [[PRODUCTION-SETUP]] calls this out as a release blocker instead of presenting TLS as a substitute for the session policy.

## Configuration you must not leave at defaults

Everything boots with defaults for development convenience, and three of those defaults are published in this repository, which makes them exactly as secret as a sticky note on the monitor:

| Variable | Default | Why you must change it |
|---|---|---|
| `PIXELS_ACCESS_KEY` | `pixels-development-access-key-change-me` | Guards every private HTTP route, **including SSO ticket creation**. Anyone holding it can mint a login ticket for any player. |
| `SSO_KEY` | `pixels-development-sso-key-change-me` | HMAC key deriving the Redis storage keys for tickets. |
| `PIXELS_ENV` | `development` | Leaves connection security optional and exposes `/docs`. Set `production`. |

Related knobs with sane defaults you may still want to tune: `SSO_DEFAULT_TTL` (five minutes; tickets are consumed within seconds in a healthy flow, so shorter is fine), `SSO_PREFIX` (Redis namespacing when sharing an instance), and the `PIXELS_WS_*` family, outbound queue size, read/write timeouts, ping cadence, that governs the WebSocket layer everything above rides on.

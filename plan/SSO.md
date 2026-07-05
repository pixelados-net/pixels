# SSO Ticket Plan

This plan documents the Redis-backed single sign-on ticket flow used by Pixels. The goal is one-time SSO tickets without SQL storage, while keeping the current user identity field as a temporary TODO until the real user/account model exists.

## Goals

- Create opaque SSO tickets through a private Fiber API route.
- Store tickets in Redis with a TTL.
- Consume tickets exactly once.
- Optionally bind a ticket to an IP address.
- Avoid storing raw ticket values in Redis.
- Keep the feature inside `internal/auth/sso` until the wider auth model is stable.

## Current Shape

`internal/auth/sso` owns:

- `Config` in `config.go`.
- `Service` in `service.go`.
- Ticket request and response domain types in `ticket.go`.
- Fx wiring in `module.go`.

`pkg/http` owns:

- The private `POST /api/sso/tickets` route.
- HTTP request and response DTOs.
- OpenAPI documentation for the route.

`pkg/redis` owns:

- Basic Redis operations.
- Atomic `Take` via Redis `GETDEL`, used for one-time ticket consumption.

## Configuration

The SSO config holder has its own `config.go` and is composed into `pkg/config.AppConfig`.

Environment variables:

| Variable | Default | Meaning |
| --- | --- | --- |
| `SSO_DEFAULT_TTL` | `5m` | Default ticket lifetime when a request does not override TTL. |
| `SSO_KEY` | `pixels-development-sso-key-change-me` | HMAC key used to derive Redis storage keys. |
| `SSO_PREFIX` | `pixels:sso` | Redis key prefix for SSO records. |

Production deployments must override `SSO_KEY`.

## Ticket Storage

Ticket creation generates 32 random bytes and returns them as base64url text.

The raw ticket is never used directly as the Redis key. Instead:

1. Generate opaque ticket value.
2. Compute `HMAC-SHA256(ticket, SSO_KEY)`.
3. Hex-encode the digest.
4. Store the record at `SSO_PREFIX:<digest>`.

The Redis payload contains:

- `userId`: temporary TODO user identifier.
- `ip`: optional IP address.
- `expiresAt`: expiration timestamp for API response and auditing.

Redis TTL is the source of truth for expiry.

## One-Time Usage

Ticket consumption computes the same HMAC-derived Redis key and uses Redis `GETDEL`.

That means:

- The first valid consume receives the record and deletes it.
- Any later consume receives missing-ticket behavior.
- Concurrent consumers race on Redis atomics instead of process memory.

## IP Binding

The IP field is optional.

- Empty stored IP means any consuming IP is accepted.
- Non-empty stored IP must match the consume request IP.
- A mismatch returns an SSO IP mismatch error after consuming the ticket.

Consuming on mismatch is intentional for now: a ticket observed from the wrong IP should not remain reusable.

## HTTP API

Private route:

```text
POST /api/sso/tickets
X-API-Key: <PIXELS_ACCESS_KEY>
Content-Type: application/json
```

Request:

```json
{
  "userId": "todo-user-id",
  "ip": "127.0.0.1",
  "ttlSeconds": 60
}
```

Response:

```json
{
  "ticket": "opaque-ticket",
  "expiresAt": "2026-07-05T12:00:00Z"
}
```

The route creates tickets only. Consumption is used by the future authentication realm flow when the client presents `SECURITY_TICKET`.

## Future Work

- Replace TODO `userId` with the real account identifier once the account model exists.
- Add authentication-realm handlers that consume SSO tickets during `SECURITY_TICKET`.
- Decide whether failed IP validation should produce an auth-failed disconnect reason or policy violation.
- Add production secret validation that rejects the development default `SSO_KEY`.
- Consider audit logging for ticket creation and consume failures without logging raw tickets.

# SSO Tickets and Login

This page covers the credential itself: how tickets are created, why they are single use by construction, what happens when one is consumed, and the bootstrap the client receives. See [[AUTH-CONNECTIONS]], [[AUTH-HANDSHAKE]], and [[AUTH-SECURITY]] for everything leading up to this moment.

## Why tickets

Pixels doesn't store passwords and has no login form. Authentication is delegated to whatever sits in front of the server (a CMS, or you with `curl` during development), which creates a **one-time ticket** bound to a player and hands it to the client:

```text
POST /api/sso/tickets   (X-API-Key required)
{"playerId": 1, "ip": "…optional…", "ttlSeconds": 300}
→ {"ticket": "…", "expiresAt": "…"}
```

This keeps the trust boundary clean: the front-end owns "who is this human" (passwords, OAuth, whatever it likes), and Pixels owns "this ticket equals this player for the next few minutes."

## The ticket lifecycle

`internal/auth/sso` builds the lifecycle on Redis primitives:

- **Creation** stores a JSON record (player ID, optional IP, expiry) under a key derived from the ticket value with an HMAC-SHA256 of `SSO_KEY`, with the TTL as the Redis expiry. Ticket values themselves are cryptographically random.
- **Consumption** uses an atomic get-and-delete (the Redis client's `Take`), which is what makes tickets single-use *by construction*: two connections racing the same ticket can't both win, and no locking is involved.
- If the ticket was created bound to an IP, consuming it from a different address fails with a distinct error.
- Expiry is enforced twice, by the Redis TTL and by checking the recorded `ExpiresAt` at consumption, so clock skew or a misconfigured Redis can't stretch a ticket's life.

A used, expired, or simply unknown ticket produces the same rejection at the protocol level; the client never learns which of the three it was.

The creation endpoint also honors an `Idempotency-Key` header: a CMS retrying a timed-out request gets the original ticket back instead of minting a second valid credential. Replays are marked with a response header so the caller can tell.

## From ticket to session

When the ticket packet arrives (in the `Handshaking` state, after security validation), the authenticator runs a strict sequence:

1. **Resolve** the ticket to a player record; failure disconnects with an authentication-failed reason.
2. **Transition** the session through `Authenticating` and mark it `Authenticated` with a timestamp.
3. **Bind** the player to the connection in the session binding registry (see [[USERS-MODEL]]), the point where "a WebSocket" becomes "player 1's connection."
4. **Bootstrap** the client (below).
5. **Announce** the login by publishing `player.connected` on the event bus, which is how every other realm (badges, rooms, pets, moderation) warms its per-player state without the connection layer knowing any of them exist.

Any failure between steps rejects the authentication and disconnects with a typed reason; there is no half-logged-in state.

## The bootstrap

Immediately after binding, the server streams everything the client needs before it can draw the hotel: identity and user info, client settings and home room, chat settings, the currency wallet, permission-derived flags, club status, and hotel availability. From the client's perspective this is "logging in"; from the server's it's a pure projection of durable state, which is why reconnecting always heals a stale client. Nothing session-scoped exists that can't be rebuilt from the database.

## Creating tickets in development

The full walkthrough, including the seeded players, lives in [[GETTING-STARTED]]. The short version:

```sh
curl -X POST http://127.0.0.1:3000/api/sso/tickets \
  -H "X-API-Key: pixels-development-access-key-change-me" \
  -H "Content-Type: application/json" \
  -d '{"playerId": 1}'
```

Point your Nitro client's `sso.ticket` at the returned value, its socket at `ws://127.0.0.1:3000/ws`, and you're in as `demo`.

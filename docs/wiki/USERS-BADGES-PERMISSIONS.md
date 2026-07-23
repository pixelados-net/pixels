# Badges and Development Seeds

Third page of the Users section, covering the badge shelf and the development players every guide assumes. Hotel authorization and its permission groups are documented separately in [[USERS-PERMISSIONS]].

## Badge inventory

`player/achievement` owns badge *inventory*: which badge codes a player has and which of the **five equip slots** each worn badge occupies. The name of the package is historical. The engine that decides when you *earn* an achievement badge lives in the progression realm; this package owns the shelf badges sit on.

The operations are small and deliberate:

- **Granting is idempotent** (`INSERT … ON CONFLICT DO NOTHING`), so double-granting from a retried admin call or a replayed reward is harmless by construction.
- **Equipping** validates ownership, slot range (1-5), and duplicates, then persists slot assignments; the live service keeps an equipped-badge snapshot per online player, warmed on `player.connected` and released on `player.disconnected`.
- **`Wearing(playerID, code)`** answers "is this player wearing badge X" from that snapshot without allocations. It exists because Wired badge conditions ask it on hot room paths.
- **Achievement badges upgrade in place.** When progression levels you from `ACH_RoomEntry1` to `ACH_RoomEntry2`, the existing row's code is replaced, preserving its slot if equipped. You wear the highest level, never a collection of superseded ones.

Grants also push incremental packets (badge added, refreshed lists) so the client's badge tab updates without relogging, the projection pattern from [[PROJECTIONS]].

## The development players

The `development` seed context creates four players with stable IDs, and the whole documentation set leans on them:

| Player | ID | Shape |
|---|---|---|
| `demo` | 1 | Administrator group; the protagonist of every QA guide |
| `alice` | 2 | Moderator, with a deliberately narrower node set |
| `bob` | 3 | Regular member |
| `carol` | 4 | Regular member, mostly used for edge-case fixtures |

The trio of permission tiers is the point: any authorization behavior can be tested immediately as "works for `demo`, partially for `alice`, denied for `bob`" without creating accounts. Seeded IDs are stable on purpose. QA documents, harness scripts, and the SSO example in [[AUTH-SSO]] all assume `demo` is player 1.

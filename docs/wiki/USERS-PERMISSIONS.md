# Permission Groups and Resolution

Pixels has two unrelated concepts called groups. Social groups are player communities. Permission groups are operator controlled authorization roles. This page covers permission groups and the exact algorithm used to decide one dotted capability.

## Permission groups are not Habbo groups

The `internal/realm/group` realm implements Habbo style social groups. They have a name, badge, owner, members, administrators, a home room, furniture rights, and a forum. Joining one is player facing social behavior.

`internal/permission` implements hotel wide authorization. Its groups resemble roles such as `member`, `moderator`, and `admin`. They have a numeric weight, optional inheritance, node grants, a client security level projection, and an optional room effect. Players never browse or join these roles through the Habbo group interface.

| Concept | Social group | Permission group |
|---|---|---|
| Owner | `internal/realm/group` | `internal/permission` |
| Purpose | Community identity and collaboration | Server authorization and staff capabilities |
| Typical names | Fan club, builders, event team | Member, moderator, admin |
| Membership | Managed through social group flows | Managed through protected permission administration |
| Badge and forum | Yes | No |
| Dotted capability nodes | No | Yes |
| Numeric weight | No | Yes |

A player may belong to many groups of either kind. Membership in a social group never grants a permission node unless a separate feature explicitly consults that social group.

## Dotted permission nodes

A node names one concrete capability:

```text
room.doorbell.answer.any
crafting.altar.manage.any
moderation.guide.duty
plugin.hello-plugin.hello.use
```

Segments use lowercase ASCII letters, digits, underscores, or hyphens. Dots express a namespace, not an automatic inheritance tree. A stored wildcard creates prefix coverage only when `*` is the final complete segment.

| Stored grant | Query | Matches | Specificity |
|---|---|---:|---:|
| `*` | `room.doorbell.answer.any` | Yes | 0 |
| `room.*` | `room.doorbell.answer.any` | Yes | 1 |
| `room.doorbell.*` | `room.doorbell.answer.any` | Yes | 2 |
| `room.doorbell.answer.any` | `room.doorbell.answer.any` | Yes | 4 |
| `room.doorbell.*` | `room.doorbell` | No | Not applicable |
| `room.*.answer` | Any query | No | Invalid syntax |

Specificity is the number of fixed segments. Exact nodes therefore beat broader wildcards inside the same resolution scope.

Realms register concrete nodes in code at startup. This produces a catalog with the declaring package, optional Nitro perk name, and plugin descriptions. Persistence may store wildcards, but checks always ask for a concrete registered capability.

## Grants and denies

Both permission groups and individual players may store a node with `allowed=true` or `allowed=false`. A false grant is an explicit deny. Removing a grant is different from denying it: removal lets the resolver continue to another source, while a deny is a decision.

The complete order is:

1. Resolve direct player grants.
2. If any direct grant matches, use the most specific direct match and stop.
3. Load the player's active permission groups by descending weight, then ascending group id for equal weights.
4. Resolve the first group whose inheritance chain contains any matching grant.
5. Ignore every lower weight group after that first group decision.
6. Deny when no source contains a matching grant.

This makes a direct player override absolute. A direct `room.* = false` wins even if an admin group grants `* = true`. It also means a high weight group with a matching deny wins over all lower weight groups.

## Resolution inside one group

Each permission group may have one parent. The resolver walks the selected group, then its parent, then the next parent until the chain ends. It detects cycles and rejects a broken chain.

Candidates inside that chain are compared in this order:

1. More fixed node segments win.
2. When specificity ties, the grant nearest the selected child group wins.
3. When specificity and inheritance depth both tie, deny wins.

Specificity comes before inheritance distance. For example, a parent exact grant beats a child wildcard because the exact node describes the requested capability more precisely. A child exact deny beats a parent exact allow because both have the same specificity and the child is nearer.

## Worked examples

Assume a player belongs to `moderator` at weight 50 and `member` at weight 0. `moderator` inherits from `member`.

| Grants | Query | Result | Reason |
|---|---|---|---|
| Moderator has `room.* = true` | `room.doorbell.answer.any` | Allow | First matching group and prefix wildcard |
| Moderator has `room.* = false`, member has exact allow | `room.doorbell.answer.any` | Deny | Moderator already produced a decision, so member membership is not considered |
| Moderator child has `room.* = false`, inherited member has exact allow | `room.doorbell.answer.any` | Allow | Both are in one inheritance chain and the parent exact grant is more specific |
| Moderator child has exact deny, inherited member has exact allow | `room.doorbell.answer.any` | Deny | Same specificity, nearer child wins |
| Direct player exact allow, admin group has `* = false` | Same exact node | Allow | Direct player decisions are resolved before groups |
| No matching direct or group grant | Any concrete node | Deny | Permissions are closed by default |

## Weight and primary group

Weight determines which membership is considered first and which group is exposed as the player's primary permission group. It is not added together, and lower groups do not contribute after a higher group has made a matching decision.

The primary group is simply the active membership with greatest weight. Pixels uses it for client security projection and synthetic group effects. Authorization still follows the complete node algorithm, so being primary does not imply every capability.

The development seeds make `demo` an admin at weight 100, `alice` a moderator at weight 50, and `bob` plus `carol` members at weight 0. Every created player receives the default `member` membership atomically.

## Perks and live projection

A registered node may map to a Nitro perk name. `EffectivePerks` resolves every mapped node and sends only allowed perks through `USER_PERKS`. The client view is therefore derived from the server permission engine rather than maintained as a second authorization list.

Permission records use a local cache for hot checks and Redis for shared cache fragments. Mutations invalidate the affected player, membership, group, and node fragments. Online players receive refreshed permissions and perks after the database commit. The warmed local resolution path is designed to remain allocation free.

## Administration

The protected permission routes expose the registered catalog, permission groups, memberships, direct player grants, effective nodes, and individual checks. `GET /docs` in development documents the exact request bodies and responses.

The API key authenticates the HTTP caller to the private API boundary. Permission nodes then authorize the acting player for domain operations. These are separate checks: possession of `X-API-Key` should not be treated as an unlimited staff rank inside gameplay workflows.

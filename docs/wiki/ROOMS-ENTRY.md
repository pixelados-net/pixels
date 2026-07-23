# Doors, Passwords, and Doorbells

Second page of the Rooms section. Getting into a room is a chain of decisions: is it open, are you banned, do you already have rights, is it locked behind a password or an owner's approval. All of it is resolved by one authorization service before the room realm ever joins a player in. This page walks that chain, then the two protected modes that can stop it: passwords and doorbells.

## The four door modes

```go
type DoorMode int16

const (
	DoorModeOpen      DoorMode = 0 // direct entry
	DoorModeDoorbell  DoorMode = 1 // requires owner approval
	DoorModePassword  DoorMode = 2 // requires a password
	DoorModeInvisible DoorMode = 3 // hides the room from normal access
)
```

`Invisible` doesn't add a fourth authorization rule of its own. It's enforced earlier, by keeping the room out of normal navigator listings and direct-entry paths (see [[NAVIGATOR-BROWSING]]); once someone does have a route to the room id, invisible behaves like closed-by-default the same way doorbell does.

## The authorization chain

`entry.Service.Authorize` runs a fixed sequence of checks, each one a potential early exit, before it ever looks at the door mode itself:

```go
func (service *Service) Authorize(ctx context.Context, request Request) (Result, error) {
	enterAny, err := service.checkBan(ctx, request)     // banned + no override → ErrBanned
	if request.Room.OwnerPlayerID == request.PlayerID ||
		request.Trusted || enterAny {
		return Result{}, nil                              // owner, trusted, or staff bypass
	}
	if hasRights, _ := service.hasRights(ctx, ...); hasRights {
		return Result{}, nil                              // room-scoped build/manage rights
	}
	if request.Room.DoorMode == roommodel.DoorModeOpen {
		return Result{}, nil
	}
	if service.trust.Consume(request.PlayerID, request.Room.ID, service.now()) {
		return Result{}, nil                              // one-shot server-granted bypass
	}
	if enterAny, _ = service.hasPermission(ctx, request.PlayerID, service.nodes.EnterAny); enterAny {
		return Result{}, nil                              // global staff override permission
	}
	switch request.Room.DoorMode {
	case roommodel.DoorModePassword:
		return service.authorizePassword(ctx, request)
	case roommodel.DoorModeDoorbell:
		return Result{}, ErrDoorbellRequired
	default:
		return Result{}, ErrAccessDenied
	}
}
```

Reading top to bottom: a ban (unless the player holds the global `EnterAny` node) always loses first, regardless of ownership. Ownership, an explicit `Trusted` entry (used for server-controlled direct placements, like being sent into a game room), and room-scoped rights all bypass door mode entirely. Only once none of those apply does the actual `DoorMode` get consulted, and even then, a short-lived server-granted trust token or the global override permission can still let someone through a closed door. `GrantTrusted` is how another part of the codebase issues that one-shot bypass ahead of time (a moderator "send player to room" action, for instance) without needing to know or forge a password.

The room enter command (`internal/realm/room/access/commands/enter`) is what actually calls `Authorize` and turns its result into protocol behavior:

```go
result, err := handler.authorize(ctx, roomentry.Request{Room: room, PlayerID: player.ID(), ...})
if errors.Is(err, roomentry.ErrDoorbellRequired) {
	return handler.requestDoorbell(ctx, player, envelope.Command.Handler, room)
}
if err != nil {
	if result.Alert != "" {
		_ = handler.sendAlert(ctx, envelope.Command.Handler, result.Alert)
	}
	return handler.sendEntryError(ctx, envelope.Command.Handler, err)
}
```

`ErrDoorbellRequired` is the one error that doesn't map to a rejection. It redirects into the doorbell flow below. Every other error (`ErrBanned`, `ErrAccessDenied`, `ErrWrongPassword`, `ErrEntryLocked`) maps to the client's entry-error packet, optionally preceded by a one-time alert.

## Passwords

Room passwords are never stored in plaintext. `HashPassword` bcrypt-hashes them at set time, with the configured cost clamped into bcrypt's valid range:

```go
func HashPassword(password string, cost int) (string, error) {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	...
}
```

Checking a submitted password is a constant-shape bcrypt compare, and a correct match clears any accumulated failed-attempt counter:

```go
func (service *Service) authorizePassword(ctx context.Context, request Request) (Result, error) {
	if locked, _ := service.locked(ctx, request.Room.ID, request.PlayerID); locked {
		return Result{}, ErrEntryLocked
	}
	if passwordMatches(request.Room.PasswordHash, request.Password) {
		return Result{}, service.redis.Delete(ctx, attemptKey(request.Room.ID, request.PlayerID))
	}
	return service.failedPassword(ctx, request.Room.ID, request.PlayerID)
}
```

Failed attempts are tracked per player per room in Redis with a sliding window (`AttemptWindow`); once the count reaches `MaxPasswordAttempts`, the player is locked out of that specific room for `LockoutDuration`, and the lockout itself, not just the wrong-password error, is what the client sees on any further try during the window, including a localized one-time alert. Lockouts are per player and per room by construction (the Redis key is `room:entry:lockout:{roomID}:{playerID}`), so a wrong guess against one room's password never affects another room, and one player's lockout never affects anyone else trying the same room.

## Doorbells

A doorbell request doesn't reject the visitor. It parks them:

```go
func (handler Handler) requestDoorbell(ctx context.Context, player *playerlive.Player, connection netconn.Context, room roommodel.Room) error {
	active, found := handler.Runtime.Find(room.ID)
	approvers, err := handler.doorbellApprovers(ctx, active, room)
	entry := roomdoorbell.Entry{PlayerID: player.ID(), Username: player.Username(), Handler: connection, RequestedAt: time.Now()}
	if !active.RequestDoorbell(entry, len(approvers) > 0) {
		return handler.sendDoorbellDenied(ctx, connection)
	}
	...
}
```

The waiting request is denied outright if the room currently has nobody eligible to answer it. There's no point queuing a knock nobody can hear. The queue itself (`internal/realm/room/access/doorbell`) is an in-memory, mutex-protected map keyed by player id, scoped to one active room instance, not a durable table. A doorbell request that outlives the room (everyone leaves before answering) simply stops existing along with the room's live state, consistent with [[ROOMS-RUNTIME]]'s point that active-room state is disposable.

Three things remove a waiting entry: an explicit resolution by username (`Resolve`, when someone answers), the room's own tick sweeping anything past `doorbellTimeout` (`Sweep`, `ExpiredTimeout`), or the room being torn down or losing every possible responder (`Drain`, `ExpiredNoRightsHolder` / `ExpiredRoomClosed`). All three reasons are distinct `ExpireReason` values specifically so the client-facing message can say *why* the knock went unanswered instead of a generic timeout.

Answering is its own command, `room.doorbell.respond`, gated by the same rights model as room management:

```go
func (service *Service) CanAnswerDoorbell(ctx context.Context, roomID int64, ownerPlayerID int64, playerID int64) (bool, error) {
	if playerID == ownerPlayerID {
		return true, nil
	}
	if hasRights, _ := service.hasRights(ctx, roomID, playerID); hasRights {
		return true, nil
	}
	return service.hasPermission(ctx, playerID, service.nodes.AnswerAnyDoorbell)
}
```

The owner can always answer; a rights-holder can answer without owning the room; and a global `AnswerAnyDoorbell` permission node lets staff answer requests in any room, the same escalation pattern documented in [[USERS-PERMISSIONS]]. Accepting a doorbell request re-enters the visitor through the normal enter command with `Trusted: true`, so a doorbell approval doesn't bypass anything else in the chain above it. It just supplies the one bypass that was missing.

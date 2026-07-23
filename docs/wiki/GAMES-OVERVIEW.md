# The Games Engine

First page of the Games section. Four furniture-based games are fully server-authoritative: Battle Banzai, Freeze, Football, and Tag (in its IceTag, Rollerskate, and Bunnyrun variants). This page covers what all four share: one package, one dispatch chain, one room-tick cycle, one team/timer/scoring model, before [[GAMES-AREA]] and [[GAMES-TEAM]] cover what makes each game distinct. As [[ARCHITECTURE]] notes, these are the real playable games; `gamecenter` is a separate, genuinely thin external-lobby realm.

## One package, one state struct per room

All four games live in `internal/realm/room/world/games` (`roomgames.Service`), reached from the furniture-interaction boundary through a thin adapter that avoids an import cycle between the furniture and room-world realms:

```go
// Package game adapts room games to the furniture interaction boundary.
type Service struct {
	games *roomgames.Service
}

func (service *Service) UseFurniture(ctx context.Context, request essential.Request) (bool, error) {
	return service.games.UseFurniture(ctx, roomgames.UseRequest{
		PlayerID: request.PlayerID, Room: request.Room, Item: request.Item, State: request.State,
	})
}
```

Registered against `essential.Service` as an `External` handler (see [[FURNITURE-INTERACTIONS]]), so an unrecognized click falls through the generic interaction registry until this adapter claims it. Every active room that hosts game furniture gets exactly one `roomState` struct holding **all four** games' live state at once: a board for Banzai, three arenas for Tag's variants, in-flight balls for Football, live players for Freeze, keyed by room id in `service.states`. A room can physically host furniture from more than one game, though which one it's officially "playing" is resolved by a fixed-priority scan:

```go
func roomGameKind(active *roomlive.Room) string {
	interactions := []string{"battlebanzai_tile", "freeze_tile", "football", "icetag_field", "rollerskate_field", "bunnyrun_field"}
	kinds := []string{"banzai", "freeze", "football", "tag", "tag", "tag"}
	for index, interaction := range interactions {
		if len(active.FurnitureByInteraction(interaction)) > 0 {
			return kinds[index]
		}
	}
	return "wired"
}
```

Every game furniture item is recognized purely by its `InteractionType` string. There's no separate "is this a game item" flag on the definition, the same convention [[FURNITURE-INTERACTIONS]] uses everywhere else.

## Two ways a game reacts: clicks and movement

Some game furniture responds to the generic furniture-use click path described in [[FURNITURE-INTERACTIONS]]:

```go
func (service *Service) UseFurniture(ctx context.Context, request UseRequest) (bool, error) {
	kind := request.Item.Definition.InteractionType
	switch {
	case kind == "game_timer":
		return true, service.toggleTimer(ctx, request) // or increaseTimer on a second click mode
	case strings.HasPrefix(kind, "football_counter_"):
		return true, service.footballCounterClick(ctx, request)
	case kind == "freeze_block" || kind == "freeze_tile":
		return true, service.throwFreeze(ctx, request)
	case kind == "football":
		return true, service.kickFootball(request)
	case strings.HasSuffix(kind, "_pole"):
		return true, nil // poles are click-inert; server-driven only
	default:
		return false, nil
	}
}
```

But Battle Banzai's tile stepping, team-gate joining (both Banzai and Freeze), Tag's field join/leave, and Football's walk-into kick are **not** clicks at all. They're driven by the same `furniture.walkedon`/`walkedoff` and unit-movement bus events that drive pressure plates and effect tiles elsewhere in the codebase (see [[FURNITURE-INTERACTIONS]] and [[ROOMS-ENTITIES]]). A game reacting to a player simply walking somewhere isn't a special case; it's the same event plumbing every other occupancy-driven furniture interaction already uses.

## Team gates: one shared join mechanism

Battle Banzai and Freeze both assign players to one of four teams by walking onto a colored gate. There's no team-selection dialog; the gate *is* the selector:

```go
if team, gate := gateTeam(kind); gate {
	if snapshot, exists := service.wired.Snapshot(payload.RoomID); exists && snapshot.Running {
		return nil // no switching teams mid-match
	}
	if joined && current == team {
		changed = service.wired.LeaveTeam(payload.RoomID, payload.PlayerID)
	} else {
		changed = service.wired.JoinTeam(payload.RoomID, payload.PlayerID, team)
	}
}

func gateTeam(kind string) (int32, bool) {
	colors := []string{"_r", "_g", "_b", "_y"}
	if !strings.Contains(kind, "_gate_") {
		return 0, false
	}
	for index, suffix := range colors {
		if strings.HasSuffix(kind, suffix) {
			return int32(index + 1), true
		}
	}
	return 0, false
}
```

`wired` here is `internal/realm/room/world/wired/game`, the same ephemeral team-and-score store WIRED's own "give score" and "join team" trigger actions use outside of a furniture-game context. The games engine doesn't maintain a separate team system; it's a consumer of WIRED's. Football has no team gates at all (its own gate is a clothing swap, covered in [[GAMES-TEAM]]); Tag has no gates either. Participation is auto-join by stepping onto the field, covered in [[GAMES-TEAM]].

## The shared timer and match boundary

A single `game_timer` furniture item, when present, is what starts and stops a match. Clicking it with furniture-management rights toggles start/pause, and because there's exactly one timer per room state, every game type co-located in that room shares the same match boundary. Its step ladder (`30s, 60s, 120s, 180s, 300s, 600s` by default, Arcturus-compatible) is parsed from the placed timer's `CustomParams`, the same free-form configuration field described in [[FURNITURE-MODEL]].

## The room-tick cycle, not a separate goroutine

Exactly like rollers, teleports, pets, and bots (see [[ROOMS-ENTITIES]]), the games engine has no timer or goroutine of its own. It hooks the room's existing owner-loop tick:

```go
rooms.AddCyclePublisher(service.Cycle)
rooms.AddClosePublisher(service.Close)

func (service *Service) Cycle(ctx context.Context, active *roomlive.Room, now time.Time) error {
	if err := service.cycleFreeze(ctx, active, now); err != nil {
		return err
	}
	if err := service.cycleFootball(ctx, active); err != nil {
		return err
	}
	service.cycleTag(ctx, active, now)
	// ...match timer tick, finish() when it hits zero
	return nil
}
```

One `Cycle` call per room per tick drives all four games' time-based work in a fixed, cheap order: Freeze's scheduled explosions, Football's incremental ball movement, Tag's per-minute progression credit, and the shared timer countdown. A room with none of these active pays almost nothing, since each sub-cycle exits immediately when it finds no due work.

## Ending a match and scoring

A match ends either when the shared timer reaches zero, or when a game-specific condition fires early (Banzai's board fully locked, Freeze's last-team-standing check) and calls `finish` directly:

```go
func (service *Service) finish(ctx context.Context, active *roomlive.Room, startedAt time.Time) error {
	_ = service.coordinator.End(ctx, active.ID())
	kind := roomGameKind(active)
	for playerID, team := range snapshot.Teams {
		entries = append(entries, Score{RoomID: active.ID(), Kind: kind, StartedAt: startedAt,
			PlayerID: playerID, Team: team, Score: snapshot.Scores[playerID] - snapshot.WiredScores[playerID],
			TeamScore: snapshot.TeamScores[team]})
		service.progress(ctx, playerID, "game.played", 1)
		service.progress(ctx, playerID, playedKey(kind), 1)
	}
	service.progressWinners(ctx, snapshot, kind)
	if service.scores != nil && len(entries) > 0 {
		return service.scores.Save(ctx, entries)
	}
	return nil
}
```

`finish` only produces a durable score row (`room_game_scores`, one row per participant per match) for players who are in WIRED's `Teams` map, meaning only games that actually call `JoinTeam` leave a persistent record. Battle Banzai and Freeze do; Football and Tag don't, so their matches are functionally ephemeral beyond the achievement progress events each one fires directly (Football's per-goal progress, Tag's per-minute progress). Worth knowing before assuming all four games are symmetric in what survives a match.

## Shared broadcast helpers

All four reuse the same small set of projection helpers rather than each defining their own packets:

- **`projectState`**: writes one furniture item's multi-state `ExtraData` and broadcasts the furniture state packet (header `2376`). Used for the timer countdown, Banzai's leader-indicator sphere, Football's scoreboards, Tag's pole flashes.
- **`projectStatesBatch`**: the same, batched across several tiles in one packet (header `1453`). Used for Banzai's captured region and Freeze's blast wave.
- **`sendPlaying`**: toggles the client's `PLAYING_GAME` flag (header `448`) on gate or field join/leave.
- **`coordinator.NotifyScore`** and **`coordinator.AddScore`**: fire WIRED's score-related triggers and, for team games, mutate the shared team score.

Because every game routes through these same helpers, adding a fifth game means reusing this vocabulary rather than inventing new packet plumbing. The pattern established here is deliberately the extension point.

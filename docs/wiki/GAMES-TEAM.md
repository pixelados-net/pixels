# Football and Tag

Third page of the Games section. Neither game uses the team-gate or scoring model [[GAMES-OVERVIEW]] and [[GAMES-AREA]] describe. Football is a physics simulation with no teams at all, and Tag is a continuous chase with no scoring at all. Both are still fully server-authoritative; they're just built on different mechanics than Banzai and Freeze.

## Football: kick, roll, rebound, score

A football has no team gate; there's no `football_gate_*` color family. What's called a "gate" for football is actually a wardrobe swap (below), a different feature entirely.

Kicking works two ways: walking directly into the ball, or clicking it while adjacent. Either queues a movement with a fixed step budget:

```go
func (service *Service) queueFootball(active *roomlive.Room, itemID int64, direction uint8, kickerID int64) {
	if ball := state.footballs[itemID]; ball != nil && ball.resetting {
		return // still returning home after a goal, not kickable yet
	}
	state.footballs[itemID] = &footballBall{direction: direction, remaining: 6, kickerID: kickerID}
}
```

The ball then advances one tile per room tick until its budget runs out, trying the straight-line target first and falling back to one of three rebound directions if that tile is blocked:

```go
func Rebounds(rotation uint8) [3]uint8 {
	switch rotation % 8 {
	case 1: return [3]uint8{7, 3, 5}
	case 3: return [3]uint8{5, 1, 7}
	case 5: return [3]uint8{3, 7, 1}
	case 7: return [3]uint8{1, 5, 3}
	default:
		reverse := (rotation + 4) % 8
		return [3]uint8{reverse, reverse, reverse}
	}
}
```

A tile counts as blocked if it's occupied, holds non-stackable furniture, requires too large a height step (matching the `MaxStepUp` rule from [[ROOMS-HEIGHTMAP]]), or is a goal not facing the ball's direction of travel. That last condition is also how scoring itself is decided: the goal only counts a ball entering roughly through its front, not a ball that clips its back or side:

```go
func GoalScores(movementRotation uint8, goalRotation uint8) bool {
	scoringDirection := (goalRotation + 4) % 8
	difference := (movementRotation + 8 - scoringDirection) % 8
	return difference == 0 || difference == 1 || difference == 7 // within one 45° step of straight-in
}
```

A scored goal increments the matching color's `football_counter_*` furniture, wrapping at 100, and notifies WIRED's score trigger, but does **not** touch WIRED's team-score state the way Banzai and Freeze do:

```go
next := (previous + 1) % 100
service.coordinator.NotifyScore(active.ID(), kickerID, int64(previous), int64(next))
```

That's why [[GAMES-OVERVIEW]] flags Football as producing no durable `room_game_scores` rows: without a call to `JoinTeam`, nobody is ever a tracked participant when a match ends. The scoreboard counter itself, and the achievement progress fired on every goal, are the only lasting record of a football match. After a goal the ball freezes and returns to its kickoff tile three-quarters of a second later, rather than staying wherever it crossed the line.

The actual `football_gate` interaction is a dressing-room kit swap, not a team join. Walking onto one merges a preset clothing kit into the player's figure, replacing only the clothing-relevant parts (chest, chest accessory, coat, cap, legs, waist, shoes) and preserving everything else, then restores the player's original figure when they leave:

```go
var footballClothing = map[string]struct{}{"ch": {}, "ca": {}, "cc": {}, "cp": {}, "lg": {}, "wa": {}, "sh": {}}
```

## Tag: proximity catch, no scoring, three variants sharing one mechanism

IceTag, Rollerskate, and Bunnyrun are the same mechanism with different avatar effects. A room can host all three arenas simultaneously, each tracked independently:

```go
tags: map[tag.Variant]*tag.Game{
	tag.IceTag: tag.New(tag.IceTag), tag.Rollerskate: tag.New(tag.Rollerskate), tag.Bunnyrun: tag.New(tag.Bunnyrun),
}
```

Joining is automatic: stepping onto any `*_field` furniture joins that variant's arena, no gate, no timer requirement:

```go
func (game *Game) Join(playerID int64) bool {
	if _, found := game.players[playerID]; found { return false }
	game.players[playerID] = struct{}{}
	if game.tagger == 0 {
		game.tagger = playerID // the first entrant starts as tagger
	}
	return true
}
```

The tag itself transfers on ordinary movement, not a click. Every player movement in a room hosting an active Tag arena checks whether the mover just became adjacent to the current tagger (in either direction: the tagger approaching a target, or a target approaching the tagger):

```go
func (game *Game) Transfer(sourceID int64, targetID int64, adjacent bool) bool {
	if !adjacent || sourceID != game.tagger || sourceID == targetID {
		return false
	}
	game.tagger = targetID
	return true
}
```

This is checked from the same `roommoved` bus event [[ROOMS-ENTITIES]] describes as the general unit-movement notification. Tag doesn't add its own movement tracking; it subscribes to the one that already exists. If the current tagger leaves the arena, the tag passes deterministically to the lowest-id remaining player rather than being left unassigned:

```go
func (game *Game) Leave(playerID int64) bool {
	delete(game.players, playerID)
	if game.tagger != playerID {
		return true
	}
	game.tagger = 0
	if players := game.Players(); len(players) > 0 {
		game.tagger = players[0]
	}
	return true
}
```

The three variants differ only in their avatar effect id, and only IceTag and Rollerskate distinguish the tagger visually from everyone else. Bunnyrun's non-tagger players get no special effect at all, just the tagger:

```go
func Effect(variant Variant, female bool, tagger bool) int32 {
	switch variant {
	case IceTag:
		base := int32(38); if female { base = 39 }; if tagger { base += 7 }; return base
	case Rollerskate:
		base := int32(55); if female { base = 56 }; if tagger { base += 2 }; return base
	case Bunnyrun:
		if tagger { return 68 }
	}
	return 0
}
```

Tag has no scoring and no end condition of its own. No team assignment means no `room_game_scores` rows and no winner, matching Football's asymmetry with Banzai and Freeze. What Tag does track is continuous presence credit: once a minute, every player currently in an IceTag or Rollerskate arena (Bunnyrun has no equivalent) earns achievement progress just for participating, independent of whether any shared timer is even running. If a `game_timer` happens to be present and expires, `finish` still runs, but since Tag participants never entered WIRED's team map, that path produces no durable record for them. Tag's entire footprint is the in-memory arena state and the progress events it fires directly.

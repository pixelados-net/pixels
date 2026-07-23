# Battle Banzai and Freeze

Second page of the Games section. Both games are about claiming territory: Banzai by locking tiles into a flood-filled region, Freeze by eliminating the opposing team from a shared field. [[GAMES-OVERVIEW]] covers the dispatch, team-gate, and timer mechanics both share; this page covers what makes each distinct.

## Battle Banzai: step, lock, flood-fill capture

Setup scans every `battlebanzai_tile` in the room, computes their bounding rectangle, and builds a board sized to it. The arena is however the tiles are actually laid out, not a fixed size. Each tile is a compact two-field state:

```go
type Tile struct {
	Team     uint8 // color one through four, or zero when unowned
	Progress uint8 // zero through two; two is permanently locked
}

func (tile Tile) Locked() bool { return tile.Team != 0 && tile.Progress == 2 }

func (tile *Tile) Step(team uint8, stealPoints int, lockPoints int) (int, bool) {
	if team < 1 || team > 4 || tile.Locked() {
		return 0, false
	}
	if tile.Team != team {
		stolen := tile.Team != 0
		tile.Team, tile.Progress = team, 0
		if stolen {
			return stealPoints, false
		}
		return 0, false
	}
	if tile.Progress < 2 {
		tile.Progress++
	}
	if tile.Progress == 2 {
		return lockPoints, true
	}
	return 0, false
}
```

Stepping on a neutral or enemy tile claims it at progress zero (an enemy tile being claimed pays a small steal bonus); stepping on your own team's tile advances it, and the third step locks it permanently. A locked tile can never be stolen back. The only way to affect it further is through the capture mechanic below.

Locking a tile immediately attempts a **flood-fill capture** of the largest fully-enclosed region belonging to your team: a breadth-first search over every unlocked-or-enemy tile, discarding any candidate region that touches the arena's outer edge:

```go
func (board *Board) CaptureLargest(team uint8) []int {
	for start, tile := range board.Tiles {
		if tile.Team == team || board.visited[start] == board.generation {
			continue
		}
		// BFS from start, 4-directional, tracking whether the region touches an edge
		...
		if !touchesEdge && len(board.candidate) > len(board.best) {
			board.best = append(board.best[:0], board.candidate...)
		}
	}
	for _, index := range board.best {
		board.Tiles[index] = Tile{Team: team, Progress: 2}
	}
	return board.best
}
```

An edge-touching region is never captured, because it isn't actually enclosed. This is the rule that makes surrounding an opponent's tiles meaningful rather than every step-3 lock instantly capturing the entire remaining board. Every tile in the captured region is force-set to `{Team: team, Progress: 2}`. Captured tiles are locked, not merely re-teamed, so a large capture is worth defending immediately once it happens.

The orchestration ties scoring to both effects of one step:

```go
points, locked := state.board.Tiles[index].Step(uint8(team), config.Banzai.PointsSteal, config.Banzai.PointsLock)
if locked {
	captured = state.board.CaptureLargest(uint8(team))
}
totalPoints := points + len(captured)*config.Banzai.PointsFill
if totalPoints != 0 {
	service.coordinator.AddScore(ctx, active.ID(), playerID, int64(totalPoints))
}
```

Battle Banzai also reuses Football's ball-physics engine for its `battlebanzai_puck` furniture: kicking the puck moves it exactly like a football, but landing on a tile calls the tile-step logic above instead of checking for a goal, one concrete example of the cross-game reuse [[GAMES-OVERVIEW]] encourages. A `battlebanzai_random_teleport` tile is the arena's other movement twist: walking onto one briefly freezes the player, then teleports them to a different teleport pad in the room, chosen deterministically from the player and source tile ids rather than by true randomness.

## Freeze: throw, delayed blast, elimination

Freeze participants are exactly whoever is on a team when the match starts. Walking onto a `freeze_gate_*` after the match has begun doesn't join it. Each starts with a fixed life count, one snowball, and blast radius one:

```go
for playerID, team := range snapshot.Teams {
	state.freezePlayers[playerID] = &freeze.Player{
		ID: playerID, Team: uint8(team), Lives: config.Freeze.MaxLives, Snowballs: 1, Radius: 1,
	}
}
```

Throwing a snowball, from an owned `freeze_tile` you're adjacent to, or an adjacent `freeze_block`, doesn't resolve instantly. It's scheduled with a two-second fuse and resolved on a later room tick, the only one of the four games with this real-time delay:

```go
state.freezeBalls = append(state.freezeBalls, freeze.Throw{
	OwnerID: request.PlayerID, ItemID: request.Item.ID, BlockID: blockID, Center: request.Item.Point,
	Deadline: time.Now().Add(2 * time.Second), Radius: player.Radius, Diagonal: player.Diagonal, Massive: player.Massive,
})
```

When the fuse expires, the blast generates a ray of points from its center (orthogonal by default, diagonal or all eight directions ("massive") if the thrower picked up those boosts), stopping at the first invalid or blocked tile along each ray:

```go
func Explosion(center grid.Point, radius int, diagonal bool, massive bool, valid func(grid.Point) bool) []grid.Point {
	directions := [][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}
	if diagonal { directions = [][2]int{{1, -1}, {1, 1}, {-1, 1}, {-1, -1}} }
	if massive  { directions = [][2]int{{0, -1}, {1, -1}, {1, 0}, {1, 1}, {0, 1}, {-1, 1}, {-1, 0}, {-1, -1}} }
	...
}
```

Every player standing on a hit point (other than the thrower) takes a hit, unless shielded or already frozen:

```go
func (player *Player) Hit(now time.Time, frozen time.Duration, lostSnowballs int, lostBoost int) bool {
	if !player.Alive() || player.ProtectedUntil.After(now) || player.FrozenUntil.After(now) {
		return false
	}
	player.Lives--
	player.FrozenUntil = now.Add(frozen)
	player.Snowballs, player.Radius = max(1, player.Snowballs-lostSnowballs), max(1, player.Radius-lostBoost)
	player.Diagonal, player.Massive = false, false
	return true
}
```

A defeated player is server-teleported to a `freeze_exit` tile. Hitting your own team scores negative points rather than positive: friendly fire is actively discouraged, not merely unrewarded:

```go
func FreezePoints(ownerTeam uint8, targetTeam uint8, points int) int {
	if ownerTeam != 0 && ownerTeam == targetTeam {
		return -points
	}
	return points
}
```

`freeze_block` tiles caught in a blast break and may drop a power-up, chosen by hashing the block and player ids together rather than a true random draw. The result is deterministic per pairing, which is what makes the mechanic reproducible in tests:

```go
func Drop(blockID int64, playerID int64, chance int) (PowerUp, bool) {
	value := uint64(blockID)*0x9e3779b185ebca87 ^ uint64(playerID)*0xc2b2ae3d27d4eb4f
	value ^= value >> 33
	value *= 0xff51afd7ed558ccd
	value ^= value >> 33
	if chance < 100 && int(value%100) >= chance {
		return 0, false
	}
	return PowerUp(2 + (value/100)%6), true
}
```

Power-ups extend range, add snowballs, unlock diagonal or massive blasts, add a life, or grant a temporary shield. Walking onto a broken block (not clicking it) collects whichever one it revealed.

A match ends the moment at most one team still has a living member, checked after every hit:

```go
func freezeMatchOver(players map[int64]*freeze.Player) bool {
	// tracks which teams are represented and which still have a living player
	return teamCount > 1 && aliveCount <= 1
}
```

When that happens, Freeze calls `finish` directly rather than waiting for the shared timer. Freeze is the one game on this page with its own real win condition independent of the timer running out.

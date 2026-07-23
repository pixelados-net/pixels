// Package freeze implements deterministic server-authored Freeze mechanics.
package freeze

import (
	"sort"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
)

// PowerUp identifies a Freeze block reward.
type PowerUp uint8

const (
	// RangeUp increases blast radius.
	RangeUp PowerUp = 2
	// BallUp increases ammunition.
	BallUp PowerUp = 3
	// Diagonal changes the next throw to diagonal rays.
	Diagonal PowerUp = 4
	// Massive changes the next throw to all eight rays.
	Massive PowerUp = 5
	// LifeUp grants one life up to the configured maximum.
	LifeUp PowerUp = 6
	// Shield grants temporary protection.
	Shield PowerUp = 7
)

// Throw stores one scheduled Freeze projectile and its captured boosts.
type Throw struct {
	// OwnerID identifies the throwing player.
	OwnerID int64
	// ItemID identifies the launching tile or ice block.
	ItemID int64
	// BlockID identifies the clicked block, when present.
	BlockID int64
	// Center stores the blast origin.
	Center grid.Point
	// Deadline stores the room tick due time.
	Deadline time.Time
	// Radius stores the captured throw range.
	Radius int
	// Diagonal stores the captured diagonal boost.
	Diagonal bool
	// Massive stores the captured massive boost.
	Massive bool
}

// Player stores authoritative Freeze state.
type Player struct {
	// ID identifies the player.
	ID int64
	// Team stores color one through four.
	Team uint8
	// Lives stores remaining lives.
	Lives int
	// Snowballs stores available ammunition.
	Snowballs int
	// Radius stores blast range.
	Radius int
	// Diagonal reports whether the next throw is diagonal.
	Diagonal bool
	// Massive reports whether the next throw uses all eight rays.
	Massive bool
	// FrozenUntil stores immobilization expiry.
	FrozenUntil time.Time
	// ProtectedUntil stores shield expiry.
	ProtectedUntil time.Time
	// Score stores non-Wired game score.
	Score int64
}

// Alive reports whether a player remains in the match.
func (player Player) Alive() bool { return player.Lives > 0 }

// ApplyPowerUp mutates bounded player state.
func (player *Player) ApplyPowerUp(power PowerUp, maxSnowballs int, maxLives int, now time.Time, protection time.Duration, stack bool) {
	switch power {
	case RangeUp:
		player.Radius++
	case BallUp:
		if player.Snowballs < maxSnowballs {
			player.Snowballs++
		}
	case Diagonal:
		player.Diagonal = true
	case Massive:
		player.Massive = true
	case LifeUp:
		if player.Lives < maxLives {
			player.Lives++
		}
	case Shield:
		start := now
		if stack && player.ProtectedUntil.After(now) {
			start = player.ProtectedUntil
		}
		player.ProtectedUntil = start.Add(protection)
	}
}

// Hit applies one blast and reports whether a life was lost.
func (player *Player) Hit(now time.Time, frozen time.Duration, lostSnowballs int, lostBoost int) bool {
	if !player.Alive() || player.ProtectedUntil.After(now) || player.FrozenUntil.After(now) {
		return false
	}
	player.Lives--
	player.FrozenUntil = now.Add(frozen)
	player.Snowballs -= lostSnowballs
	if player.Snowballs < 1 {
		player.Snowballs = 1
	}
	player.Radius -= lostBoost
	if player.Radius < 1 {
		player.Radius = 1
	}
	player.Diagonal, player.Massive = false, false
	return true
}

// FreezePoints returns a reward for an opponent or an equal friendly-fire penalty.
func FreezePoints(ownerTeam uint8, targetTeam uint8, points int) int {
	if ownerTeam != 0 && ownerTeam == targetTeam {
		return -points
	}
	return points
}

// ArmedState returns Nitro's native armed Freeze furniture state.
func ArmedState(radius int) int {
	if radius < 1 {
		radius = 1
	}
	return (radius + 1) * 1000
}

// ResetState returns Nitro's native delayed reset for one blast distance.
func ResetState(distance int) int {
	if distance < 0 {
		distance = 0
	}
	return 11 + distance*100
}

// Distance returns the Chebyshev blast distance between two tiles.
func Distance(center grid.Point, point grid.Point) int {
	dx, dy := int(center.X)-int(point.X), int(center.Y)-int(point.Y)
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	if dy > dx {
		return dy
	}
	return dx
}

// ApproachPoints returns valid neighboring tiles nearest to one player.
func ApproachPoints(center grid.Point, player grid.Point) []grid.Point {
	offsets := [8][2]int{{0, -1}, {1, -1}, {1, 0}, {1, 1}, {0, 1}, {-1, 1}, {-1, 0}, {-1, -1}}
	points := make([]grid.Point, 0, len(offsets))
	for _, offset := range offsets {
		if point, ok := grid.NewPoint(int(center.X)+offset[0], int(center.Y)+offset[1]); ok {
			points = append(points, point)
		}
	}
	sort.SliceStable(points, func(left int, right int) bool {
		return Distance(player, points[left]) < Distance(player, points[right])
	})
	return points
}

// WinningTeams returns the sole surviving team or the highest-scoring team.
func WinningTeams(players []Player) []uint8 {
	alive := make(map[uint8]struct{})
	scores := make(map[uint8]int64)
	for _, player := range players {
		if player.Alive() {
			alive[player.Team] = struct{}{}
		}
		scores[player.Team] += player.Score
	}
	if len(alive) == 1 {
		for team := range alive {
			return []uint8{team}
		}
	}
	var best int64
	first := true
	for _, score := range scores {
		if first || score > best {
			best, first = score, false
		}
	}
	winners := make([]uint8, 0, 4)
	for team, score := range scores {
		if score == best {
			winners = append(winners, team)
		}
	}
	sort.Slice(winners, func(left int, right int) bool { return winners[left] < winners[right] })
	return winners
}

// Explosion returns the center and bounded ray points in deterministic order.
func Explosion(center grid.Point, radius int, diagonal bool, massive bool, valid func(grid.Point) bool) []grid.Point {
	if radius < 1 {
		radius = 1
	}
	directions := [][2]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}}
	if diagonal {
		directions = [][2]int{{1, -1}, {1, 1}, {-1, 1}, {-1, -1}}
	}
	if massive {
		directions = [][2]int{{0, -1}, {1, -1}, {1, 0}, {1, 1}, {0, 1}, {-1, 1}, {-1, 0}, {-1, -1}}
	}
	points := make([]grid.Point, 0, 1+len(directions)*radius)
	if valid == nil || valid(center) {
		points = append(points, center)
	}
	for _, direction := range directions {
		for distance := 1; distance <= radius; distance++ {
			point, ok := grid.NewPoint(int(center.X)+direction[0]*distance, int(center.Y)+direction[1]*distance)
			if !ok || (valid != nil && !valid(point)) {
				break
			}
			points = append(points, point)
		}
	}
	return points
}

// Package tag implements IceTag, Rollerskate, and Bunnyrun state.
package tag

import "sort"

// Variant identifies one Tag arena.
type Variant uint8

const (
	// IceTag identifies the ice field.
	IceTag Variant = iota + 1
	// Rollerskate identifies the rollerskate field.
	Rollerskate
	// Bunnyrun identifies the bunny field.
	Bunnyrun
)

// Game stores one room's Tag participants and current tagger.
type Game struct {
	// Variant stores arena behavior.
	Variant Variant
	// players stores active player ids.
	players map[int64]struct{}
	// tagger stores the player who is currently it.
	tagger int64
}

// New creates one Tag arena.
func New(variant Variant) *Game { return &Game{Variant: variant, players: make(map[int64]struct{})} }

// Join adds one player and assigns the first entrant as tagger.
func (game *Game) Join(playerID int64) bool {
	if playerID <= 0 {
		return false
	}
	if _, found := game.players[playerID]; found {
		return false
	}
	game.players[playerID] = struct{}{}
	if game.tagger == 0 {
		game.tagger = playerID
	}
	return true
}

// Leave removes one player and deterministically transfers the tag when needed.
func (game *Game) Leave(playerID int64) bool {
	if _, found := game.players[playerID]; !found {
		return false
	}
	delete(game.players, playerID)
	if game.tagger != playerID {
		return true
	}
	game.tagger = 0
	players := game.Players()
	if len(players) > 0 {
		game.tagger = players[0]
	}
	return true
}

// Transfer passes the tag from the current tagger to an active adjacent target.
func (game *Game) Transfer(sourceID int64, targetID int64, adjacent bool) bool {
	if !adjacent || sourceID != game.tagger || sourceID == targetID {
		return false
	}
	if _, found := game.players[targetID]; !found {
		return false
	}
	game.tagger = targetID
	return true
}

// Tagger returns the current tagger.
func (game *Game) Tagger() int64 { return game.tagger }

// Players returns stable participant ids.
func (game *Game) Players() []int64 {
	players := make([]int64, 0, len(game.players))
	for playerID := range game.players {
		players = append(players, playerID)
	}
	sort.Slice(players, func(left int, right int) bool { return players[left] < players[right] })
	return players
}

// Effect returns the Nitro avatar effect for one variant, gender, and tag state.
func Effect(variant Variant, female bool, tagger bool) int32 {
	switch variant {
	case IceTag:
		base := int32(38)
		if female {
			base = 39
		}
		if tagger {
			base += 7
		}
		return base
	case Rollerskate:
		base := int32(55)
		if female {
			base = 56
		}
		if tagger {
			base += 2
		}
		return base
	case Bunnyrun:
		if tagger {
			return 68
		}
	}
	return 0
}

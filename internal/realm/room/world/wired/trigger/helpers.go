package trigger

import (
	"strings"
	"unicode/utf8"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
)

// containsFold reports whether text contains a Unicode case-insensitive keyword without allocation.
func containsFold(value string, keyword string) bool {
	if keyword == "" {
		return true
	}
	if strings.Contains(value, keyword) {
		return true
	}
	for start := 0; start < len(value); {
		for end := start; end < len(value); {
			_, size := utf8.DecodeRuneInString(value[end:])
			end += size
			if strings.EqualFold(value[start:end], keyword) {
				return true
			}
		}
		_, size := utf8.DecodeRuneInString(value[start:])
		start += size
	}
	return false
}

// actorAllowed reports whether event actor kind satisfies the descriptor.
func actorAllowed(node *configuration.Node, actor ActorKind) bool {
	switch node.Descriptor.Actor {
	case registry.ActorPlayer:
		return actor == ActorPlayer
	case registry.ActorUnit:
		return actor == ActorPlayer || actor == ActorBot || actor == ActorPet
	case registry.ActorBot:
		return actor == ActorBot
	default:
		return true
	}
}

// botNameMatches compares an optional compiled bot name.
func botNameMatches(node *configuration.Node, event Event) bool {
	return node.Parameters.Name == "" || strings.EqualFold(node.Parameters.Name, event.Username)
}

// targetMatches applies Nitro's explicit ID, type, and context policies without allocation.
func targetMatches(node *configuration.Node, itemID int64, spriteID int32) bool {
	if node.SelectionMode == 0 {
		return false
	}
	for _, target := range node.Targets {
		if target.ItemID == itemID {
			return true
		}
		if node.SelectionMode >= 2 && spriteID > 0 && spriteID == target.SpriteID {
			return true
		}
	}
	return false
}

// kindFor maps canonical trigger keys to typed room events.
func kindFor(key string) Kind {
	switch key {
	case "wf_trg_enter_room":
		return EnterRoom
	case "wf_trg_says_something":
		return Say
	case "wf_trg_walks_on_furni":
		return WalkOn
	case "wf_trg_walks_off_furni":
		return WalkOff
	case "wf_trg_state_changed":
		return StateChanged
	case "wf_trg_collision":
		return Collision
	case "wf_trg_periodically":
		return Periodic
	case "wf_trg_period_long":
		return PeriodicLong
	case "wf_trg_at_given_time":
		return AtTime
	case "wf_trg_at_time_long":
		return AtTimeLong
	case "wf_trg_game_starts":
		return GameStarted
	case "wf_trg_game_ends":
		return GameEnded
	case "wf_trg_score_achieved":
		return ScoreAchieved
	case "wf_trg_bot_reached_stf":
		return BotReachedFurniture
	case "wf_trg_bot_reached_avtr":
		return BotReachedAvatar
	case "wf_trg_game_team_win":
		return TeamWon
	case "wf_trg_game_team_lose":
		return TeamLost
	default:
		return 0
	}
}

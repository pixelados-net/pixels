package runtime

import "github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"

// eventKind maps canonical trigger keys to the runtime index.
func eventKind(key string) trigger.Kind {
	switch key {
	case "wf_trg_enter_room":
		return trigger.EnterRoom
	case "wf_trg_says_something":
		return trigger.Say
	case "wf_trg_walks_on_furni":
		return trigger.WalkOn
	case "wf_trg_walks_off_furni":
		return trigger.WalkOff
	case "wf_trg_state_changed":
		return trigger.StateChanged
	case "wf_trg_collision":
		return trigger.Collision
	case "wf_trg_game_starts":
		return trigger.GameStarted
	case "wf_trg_game_ends":
		return trigger.GameEnded
	case "wf_trg_score_achieved":
		return trigger.ScoreAchieved
	case "wf_trg_bot_reached_stf":
		return trigger.BotReachedFurniture
	case "wf_trg_bot_reached_avtr":
		return trigger.BotReachedAvatar
	case "wf_trg_game_team_win":
		return trigger.TeamWon
	case "wf_trg_game_team_lose":
		return trigger.TeamLost
	default:
		kind, _ := timerKind(key)
		return kind
	}
}

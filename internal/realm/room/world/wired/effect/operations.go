package effect

// furnitureOperation maps furniture effect descriptors to domain operations.
func furnitureOperation(key string) (FurnitureOperation, bool) {
	switch key {
	case "wf_act_toggle_state":
		return ToggleState, true
	case "wf_act_match_to_sshot":
		return MatchSnapshot, true
	case "wf_act_move_rotate":
		return MoveRotate, true
	case "wf_act_chase":
		return ChaseActor, true
	case "wf_act_flee":
		return FleeActor, true
	case "wf_act_move_to_dir":
		return MoveDirection, true
	case "wf_act_toggle_to_rnd":
		return ToggleRandomState, true
	case "wf_act_move_furni_to":
		return MoveFurnitureTo, true
	default:
		return 0, false
	}
}

// avatarOperation maps player-facing descriptors to domain operations.
func avatarOperation(key string) (AvatarOperation, bool) {
	switch key {
	case "wf_act_show_message":
		return ShowMessage, true
	case "wf_act_teleport_to":
		return TeleportAvatar, true
	case "wf_act_kick_user":
		return KickAvatar, true
	case "wf_act_mute_triggerer":
		return MuteAvatar, true
	case "wf_act_give_respect":
		return GiveRespect, true
	case "wf_act_alert":
		return AlertAvatar, true
	case "wf_act_give_handitem":
		return GiveHanditem, true
	case "wf_act_give_effect":
		return GiveEffect, true
	default:
		return 0, false
	}
}

// botOperation maps bot descriptors to domain operations.
func botOperation(key string) (BotOperation, bool) {
	switch key {
	case "wf_act_bot_teleport":
		return BotTeleport, true
	case "wf_act_bot_move":
		return BotMove, true
	case "wf_act_bot_talk":
		return BotTalk, true
	case "wf_act_bot_give_handitem":
		return BotGiveHanditem, true
	case "wf_act_bot_follow_avatar":
		return BotFollowAvatar, true
	case "wf_act_bot_clothes":
		return BotClothes, true
	case "wf_act_bot_talk_to_avatar":
		return BotTalkToAvatar, true
	default:
		return 0, false
	}
}

// gameOperation maps game descriptors to domain operations.
func gameOperation(key string) (GameOperation, bool) {
	switch key {
	case "wf_act_give_score":
		return GiveScore, true
	case "wf_act_join_team":
		return JoinTeam, true
	case "wf_act_leave_team":
		return LeaveTeam, true
	case "wf_act_give_score_tm":
		return GiveTeamScore, true
	case "wf_act_reset_highscore":
		return ResetHighscore, true
	default:
		return 0, false
	}
}

// progressionOperation maps progression descriptors to domain operations.
func progressionOperation(key string) (ProgressionOperation, bool) {
	switch key {
	case "wf_act_progress_achievement":
		return ProgressAchievement, true
	case "wf_act_progress_quest":
		return ProgressQuest, true
	case "wf_act_start_quest":
		return StartQuest, true
	default:
		return 0, false
	}
}

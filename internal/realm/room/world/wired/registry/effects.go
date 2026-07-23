package registry

// canonicalEffects returns the thirty audited Arcturus effect behaviors.
func canonicalEffects() []Descriptor {
	return []Descriptor{
		{Key: "wf_act_toggle_state", Family: FamilyEffect, ClientCode: 0, Selection: SelectionRequired, Editor: true},
		{Key: "wf_act_reset_timers", Family: FamilyEffect, ClientCode: 1, Editor: true},
		{Key: "wf_act_match_to_sshot", Family: FamilyEffect, ClientCode: 3, Selection: SelectionRequired, Editor: true, Aliases: []string{"wf_act_match_to_sshot_height", "wf_act_plus_match_furni_state"}},
		{Key: "wf_act_move_rotate", Family: FamilyEffect, ClientCode: 4, Selection: SelectionRequired, Editor: true},
		{Key: "wf_act_give_score", Family: FamilyEffect, ClientCode: 6, Actor: ActorPlayer, Editor: true},
		{Key: "wf_act_show_message", Family: FamilyEffect, ClientCode: 7, Actor: ActorPlayer, Editor: true},
		{Key: "wf_act_teleport_to", Family: FamilyEffect, ClientCode: 8, Selection: SelectionRequired, Actor: ActorUnit, Editor: true},
		{Key: "wf_act_join_team", Family: FamilyEffect, ClientCode: 9, Actor: ActorPlayer, Editor: true},
		{Key: "wf_act_leave_team", Family: FamilyEffect, ClientCode: 10, Actor: ActorPlayer, Editor: true},
		{Key: "wf_act_chase", Family: FamilyEffect, ClientCode: 11, Selection: SelectionRequired, Actor: ActorUnit, Editor: true},
		{Key: "wf_act_flee", Family: FamilyEffect, ClientCode: 12, Selection: SelectionRequired, Actor: ActorUnit, Editor: true},
		{Key: "wf_act_move_to_dir", Family: FamilyEffect, ClientCode: 13, Selection: SelectionRequired, Editor: true},
		{Key: "wf_act_give_score_tm", Family: FamilyEffect, ClientCode: 14, Actor: ActorPlayer, Editor: true},
		{Key: "wf_act_toggle_to_rnd", Family: FamilyEffect, ClientCode: 15, Selection: SelectionRequired, Aliases: []string{"wf_act_toggle_state_random"}},
		{Key: "wf_act_move_furni_to", Family: FamilyEffect, ClientCode: 16, Selection: SelectionRequired, Editor: true},
		{Key: "wf_act_give_reward", Family: FamilyEffect, ClientCode: 17, Actor: ActorPlayer, Editor: true},
		{Key: "wf_act_call_stacks", Family: FamilyEffect, ClientCode: 18, Selection: SelectionRequired, Editor: true},
		{Key: "wf_act_kick_user", Family: FamilyEffect, ClientCode: 19, Actor: ActorPlayer, Editor: true},
		{Key: "wf_act_mute_triggerer", Family: FamilyEffect, ClientCode: 20, Actor: ActorPlayer, Editor: true},
		{Key: "wf_act_bot_teleport", Family: FamilyEffect, ClientCode: 21, Selection: SelectionRequired, Editor: true},
		{Key: "wf_act_bot_move", Family: FamilyEffect, ClientCode: 22, Selection: SelectionRequired, Editor: true},
		{Key: "wf_act_bot_talk", Family: FamilyEffect, ClientCode: 23, Editor: true},
		{Key: "wf_act_bot_give_handitem", Family: FamilyEffect, ClientCode: 24, Actor: ActorPlayer, Editor: true},
		{Key: "wf_act_bot_follow_avatar", Family: FamilyEffect, ClientCode: 25, Actor: ActorPlayer, Editor: true},
		{Key: "wf_act_bot_clothes", Family: FamilyEffect, ClientCode: 26, Editor: true},
		{Key: "wf_act_bot_talk_to_avatar", Family: FamilyEffect, ClientCode: 27, Actor: ActorPlayer, Editor: true},
		{Key: "wf_act_give_respect", Family: FamilyEffect, ClientCode: 7, Actor: ActorPlayer},
		{Key: "wf_act_alert", Family: FamilyEffect, ClientCode: 7, Actor: ActorPlayer, Aliases: []string{"wf_act_alert_habbo"}},
		{Key: "wf_act_give_handitem", Family: FamilyEffect, ClientCode: 7, Actor: ActorPlayer},
		{Key: "wf_act_give_effect", Family: FamilyEffect, ClientCode: 7, Actor: ActorPlayer},
	}
}

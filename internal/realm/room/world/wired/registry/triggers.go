package registry

// canonicalTriggers returns the seventeen audited Arcturus trigger behaviors.
func canonicalTriggers() []Descriptor {
	return []Descriptor{
		{Key: "wf_trg_says_something", Family: FamilyTrigger, ClientCode: 0, Actor: ActorUnit, Editor: true},
		{Key: "wf_trg_walks_on_furni", Family: FamilyTrigger, ClientCode: 1, Selection: SelectionRequired, Actor: ActorUnit, Editor: true},
		{Key: "wf_trg_walks_off_furni", Family: FamilyTrigger, ClientCode: 2, Selection: SelectionRequired, Actor: ActorUnit, Editor: true},
		{Key: "wf_trg_at_given_time", Family: FamilyTrigger, ClientCode: 3, Actor: ActorOptional, Editor: true},
		{Key: "wf_trg_at_time_long", Family: FamilyTrigger, ClientCode: 3, Actor: ActorOptional, Editor: true},
		{Key: "wf_trg_state_changed", Family: FamilyTrigger, ClientCode: 4, Selection: SelectionRequired, Actor: ActorUnit, Editor: true},
		{Key: "wf_trg_periodically", Family: FamilyTrigger, ClientCode: 6, Actor: ActorOptional, Editor: true},
		{Key: "wf_trg_enter_room", Family: FamilyTrigger, ClientCode: 7, Actor: ActorPlayer, Editor: true},
		{Key: "wf_trg_game_starts", Family: FamilyTrigger, ClientCode: 8, Actor: ActorOptional, Editor: true},
		{Key: "wf_trg_game_ends", Family: FamilyTrigger, ClientCode: 9, Actor: ActorOptional, Editor: true},
		{Key: "wf_trg_score_achieved", Family: FamilyTrigger, ClientCode: 10, Actor: ActorOptional, Editor: true},
		{Key: "wf_trg_collision", Family: FamilyTrigger, ClientCode: 11, Actor: ActorUnit, Editor: true},
		{Key: "wf_trg_period_long", Family: FamilyTrigger, ClientCode: 12, Actor: ActorOptional, Editor: true},
		{Key: "wf_trg_bot_reached_stf", Family: FamilyTrigger, ClientCode: 13, Selection: SelectionRequired, Actor: ActorBot, Editor: true},
		{Key: "wf_trg_bot_reached_avtr", Family: FamilyTrigger, ClientCode: 14, Actor: ActorBot, Editor: true},
		{Key: "wf_trg_game_team_win", Family: FamilyTrigger, ClientCode: 13, Actor: ActorPlayer, Editor: true},
		{Key: "wf_trg_game_team_lose", Family: FamilyTrigger, ClientCode: 13, Actor: ActorPlayer, Editor: true},
	}
}

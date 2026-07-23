package registry

// canonicalConditions returns the twenty-four audited Arcturus conditions.
func canonicalConditions() []Descriptor {
	return []Descriptor{
		{Key: "wf_cnd_match_snapshot", Family: FamilyCondition, ClientCode: 0, Selection: SelectionRequired, Editor: true},
		{Key: "wf_cnd_furnis_hv_avtrs", Family: FamilyCondition, ClientCode: 1, Selection: SelectionRequired, Actor: ActorOptional, Editor: true},
		{Key: "wf_cnd_trggrer_on_frn", Family: FamilyCondition, ClientCode: 2, Selection: SelectionRequired, Actor: ActorUnit, Editor: true},
		{Key: "wf_cnd_time_more_than", Family: FamilyCondition, ClientCode: 3, Editor: true},
		{Key: "wf_cnd_time_less_than", Family: FamilyCondition, ClientCode: 4, Editor: true},
		{Key: "wf_cnd_user_count_in", Family: FamilyCondition, ClientCode: 5, Editor: true},
		{Key: "wf_cnd_actor_in_team", Family: FamilyCondition, ClientCode: 6, Actor: ActorPlayer, Editor: true},
		{Key: "wf_cnd_has_furni_on", Family: FamilyCondition, ClientCode: 7, Selection: SelectionRequired, Editor: true},
		{Key: "wf_cnd_stuff_is", Family: FamilyCondition, ClientCode: 8, Selection: SelectionRequired, Editor: true},
		{Key: "wf_cnd_actor_in_group", Family: FamilyCondition, ClientCode: 10, Actor: ActorPlayer, Editor: true},
		{Key: "wf_cnd_wearing_badge", Family: FamilyCondition, ClientCode: 11, Actor: ActorPlayer, Editor: true, Aliases: []string{"wf_cnd_habbo_owns_badge"}},
		{Key: "wf_cnd_wearing_effect", Family: FamilyCondition, ClientCode: 12, Actor: ActorPlayer, Editor: true},
		{Key: "wf_cnd_not_match_snap", Family: FamilyCondition, ClientCode: 13, Selection: SelectionRequired, Editor: true},
		{Key: "wf_cnd_not_hv_avtrs", Family: FamilyCondition, ClientCode: 14, Selection: SelectionRequired, Editor: true},
		{Key: "wf_cnd_not_trggrer_on", Family: FamilyCondition, ClientCode: 15, Selection: SelectionRequired, Actor: ActorUnit, Editor: true},
		{Key: "wf_cnd_not_user_count", Family: FamilyCondition, ClientCode: 16, Editor: true},
		{Key: "wf_cnd_not_in_team", Family: FamilyCondition, ClientCode: 17, Actor: ActorPlayer, Editor: true},
		{Key: "wf_cnd_not_furni_on", Family: FamilyCondition, ClientCode: 18, Selection: SelectionRequired, Editor: true},
		{Key: "wf_cnd_not_stuff_is", Family: FamilyCondition, ClientCode: 19, Selection: SelectionRequired, Editor: true},
		{Key: "wf_cnd_not_in_group", Family: FamilyCondition, ClientCode: 21, Actor: ActorPlayer, Editor: true},
		{Key: "wf_cnd_not_wearing_b", Family: FamilyCondition, ClientCode: 22, Actor: ActorPlayer, Editor: true, Aliases: []string{"wf_cnd_not_habbo_owns_badge"}},
		{Key: "wf_cnd_not_wearing_fx", Family: FamilyCondition, ClientCode: 23, Actor: ActorPlayer, Editor: true},
		{Key: "wf_cnd_date_rng_active", Family: FamilyCondition, ClientCode: 24, Editor: true},
		{Key: "wf_cnd_has_handitem", Family: FamilyCondition, ClientCode: 25, Actor: ActorPlayer, Editor: true, Aliases: []string{"wf_cnd_wears_handitem"}},
	}
}

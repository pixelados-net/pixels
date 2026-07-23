package registry

// canonicalExtras returns selectors, game pickup, and highscore behavior.
func canonicalExtras() []Descriptor {
	return []Descriptor{
		{Key: "wf_xtra_random", Family: FamilyExtra, ClientCode: -1},
		{Key: "wf_xtra_unseen", Family: FamilyExtra, ClientCode: -1},
		{Key: "wf_xtra_or_eval", Family: FamilyExtra, ClientCode: -1},
		{Key: "wf_blob", Family: FamilyExtra, ClientCode: -1, Actor: ActorPlayer},
		{Key: "wf_highscore", Family: FamilyHighscore, ClientCode: -1},
	}
}

// CanonicalManifest returns the audited seventy-six behavior descriptors.
func CanonicalManifest() []Descriptor {
	manifest := make([]Descriptor, 0, 76)
	manifest = append(manifest, canonicalTriggers()...)
	manifest = append(manifest, canonicalEffects()...)
	manifest = append(manifest, canonicalConditions()...)
	manifest = append(manifest, canonicalExtras()...)
	return manifest
}

// CompatibilityManifest returns explicitly designed non-master extensions.
func CompatibilityManifest() []Descriptor {
	return []Descriptor{
		{Key: "wf_cnd_valid_moves", Family: FamilyCondition, ClientCode: -1, Actor: ActorOptional},
		{Key: "wf_act_reset_highscore", Family: FamilyEffect, ClientCode: 0, Selection: SelectionRequired, Editor: true, Aliases: []string{"wf_cstm_reset_highscore"}},
		{Key: "wf_act_progress_achievement", Family: FamilyEffect, ClientCode: 7, Actor: ActorPlayer, Editor: true, Aliases: []string{"wf_cstm_achievement"}},
		{Key: "wf_act_progress_quest", Family: FamilyEffect, ClientCode: 7, Actor: ActorPlayer, Editor: true, Aliases: []string{"wf_cstm_progress_quest"}},
		{Key: "wf_act_start_quest", Family: FamilyEffect, ClientCode: 7, Actor: ActorPlayer, Editor: true, Aliases: []string{"wf_cstm_start_quest"}},
	}
}

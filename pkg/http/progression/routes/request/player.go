package request

import progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"

// Progress adds achievement progress through the real engine.
type Progress struct {
	Audit
	// Amount stores the positive progress delta.
	Amount int64 `json:"amount"`
}

// ForceLevel forces one exact achievement level.
type ForceLevel struct {
	Audit
	// Level stores the exact level, including zero.
	Level int32 `json:"level"`
	// PayRewards controls payment of newly crossed levels.
	PayRewards bool `json:"payRewards"`
}

// Badge grants one arbitrary badge.
type Badge struct {
	Audit
	// Badge stores the durable badge code.
	Badge string `json:"badge"`
}

// TalentLevel defines one talent track level.
type TalentLevel struct {
	Audit
	// Requirements stores achievement prerequisites.
	Requirements []progressionrecord.TalentRequirement `json:"requirements"`
	// Items stores rewarded furniture definition identifiers.
	Items []int64 `json:"items"`
	// Perks stores rewarded Nitro perk names.
	Perks []string `json:"perks"`
	// Badges stores rewarded badge codes.
	Badges []string `json:"badges"`
}

// TalentForce forces one exact player talent level.
type TalentForce struct {
	Audit
	// Level stores the exact paid track level.
	Level int32 `json:"level"`
}

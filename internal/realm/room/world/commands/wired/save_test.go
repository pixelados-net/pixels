package wired

import "testing"

// TestPrivilegedProgression verifies only the three custom progression effects require superwired.
func TestPrivilegedProgression(t *testing.T) {
	tests := map[string]bool{"wf_act_progress_achievement": true, "wf_act_progress_quest": true, "wf_act_start_quest": true, "wf_act_give_score": false, "wf_act_reset_highscore": false}
	for key, expected := range tests {
		if actual := privilegedProgression(key); actual != expected {
			t.Errorf("key=%s actual=%v expected=%v", key, actual, expected)
		}
	}
}

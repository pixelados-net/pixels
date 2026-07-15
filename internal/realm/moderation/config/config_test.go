package config

import (
	"testing"
	"time"
)

// TestLoadDefaults verifies moderation remains usable without environment overrides.
func TestLoadDefaults(t *testing.T) {
	for _, key := range []string{"PIXELS_MODERATION_ENABLED", "PIXELS_MODERATION_REPORT_LIMIT", "PIXELS_MODERATION_REPORT_WINDOW", "PIXELS_MODERATION_CONTEXT_WINDOW", "PIXELS_GUARDIAN_COUNT", "PIXELS_GUARDIAN_VOTE_WINDOW", "PIXELS_GUARDIAN_IGNORE_LIMIT", "PIXELS_GUARDIAN_EXCLUSION"} {
		t.Setenv(key, "")
	}
	value := Load()
	if !value.Enabled || value.ReportLimit != 3 || value.ReportWindow != 10*time.Minute || value.ContextWindow != 50 || value.GuardianCount != 3 || value.GuardianVoteWindow != time.Minute || value.GuardianIgnoreLimit != 3 || value.GuardianExclusion != 30*time.Minute {
		t.Fatalf("defaults=%+v", value)
	}
}

// TestLoadOverrides verifies valid environment policy is parsed.
func TestLoadOverrides(t *testing.T) {
	t.Setenv("PIXELS_MODERATION_ENABLED", "false")
	t.Setenv("PIXELS_MODERATION_REPORT_LIMIT", "5")
	t.Setenv("PIXELS_MODERATION_REPORT_WINDOW", "2m")
	t.Setenv("PIXELS_MODERATION_CONTEXT_WINDOW", "25")
	t.Setenv("PIXELS_GUARDIAN_COUNT", "5")
	t.Setenv("PIXELS_GUARDIAN_VOTE_WINDOW", "45s")
	t.Setenv("PIXELS_GUARDIAN_IGNORE_LIMIT", "2")
	t.Setenv("PIXELS_GUARDIAN_EXCLUSION", "1h")
	value := Load()
	if value.Enabled || value.ReportLimit != 5 || value.ReportWindow != 2*time.Minute || value.ContextWindow != 25 || value.GuardianCount != 5 || value.GuardianVoteWindow != 45*time.Second || value.GuardianIgnoreLimit != 2 || value.GuardianExclusion != time.Hour {
		t.Fatalf("overrides=%+v", value)
	}
}

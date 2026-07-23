package session

import (
	"testing"

	insettings "github.com/niflaot/pixels/networking/inbound/navigator/session/settings"
)

// TestSettingsBounds verifies configured multi-monitor and window constraints.
func TestSettingsBounds(t *testing.T) {
	handler := SettingsHandler{PositionLimit: 1000, MinimumWidth: 320, MaximumWidth: 800, MinimumHeight: 240, MaximumHeight: 700}
	tests := []struct {
		payload insettings.Payload
		valid   bool
	}{
		{payload: insettings.Payload{WindowX: -1000, WindowY: 1000, WindowWidth: 320, WindowHeight: 240, ResultsMode: 2}, valid: true},
		{payload: insettings.Payload{WindowX: -1001, WindowWidth: 320, WindowHeight: 240}},
		{payload: insettings.Payload{WindowWidth: 319, WindowHeight: 240}},
		{payload: insettings.Payload{WindowWidth: 320, WindowHeight: 240, ResultsMode: 3}},
	}
	for _, test := range tests {
		if actual := handler.valid(test.payload); actual != test.valid {
			t.Fatalf("payload=%#v valid=%v expected %v", test.payload, actual, test.valid)
		}
	}
}

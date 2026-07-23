package routes

import (
	"testing"

	gamecenterrecord "github.com/niflaot/pixels/internal/realm/gamecenter/record"
	outcontents "github.com/niflaot/pixels/networking/outbound/progression/poll/contents"
)

// TestCenterGameValidatesAndNormalizes verifies the administration boundary.
func TestCenterGameValidatesAndNormalizes(t *testing.T) {
	game, err := centerGame(CenterRequest{Name: "  Arcade ", BackgroundColor: "AABBCC", TextColor: "0011FF", AssetURL: " art ", LaunchURL: " game ", LaunchKind: gamecenterrecord.LaunchURL, Enabled: true, Version: 3}, 7)
	if err != nil {
		t.Fatal(err)
	}
	if game.ID != 7 || game.Name != "Arcade" || game.BackgroundColor != "aabbcc" || game.TextColor != "0011ff" || game.AssetURL != "art" || game.LaunchURL != "game" || !game.Enabled || game.Version != 3 {
		t.Fatalf("game=%+v", game)
	}
	invalid := []CenterRequest{
		{Name: "", BackgroundColor: "000000", TextColor: "ffffff", LaunchKind: gamecenterrecord.LaunchURL},
		{Name: "x", BackgroundColor: "bad", TextColor: "ffffff", LaunchKind: gamecenterrecord.LaunchURL},
		{Name: "x", BackgroundColor: "000000", TextColor: "ffffff", LaunchKind: "flash"},
	}
	for _, request := range invalid {
		if _, validationErr := centerGame(request, 1); validationErr == nil {
			t.Fatalf("accepted invalid request %+v", request)
		}
	}
}

// TestPollDefinitionValidatesOrderingAndQuestionShape verifies nested poll input.
func TestPollDefinitionValidatesOrderingAndQuestionShape(t *testing.T) {
	request := PollRequest{Title: " Poll ", Headline: " Head ", Questions: []outcontents.Question{{ID: 1, SortOrder: 0, Type: 1, Text: " Ready? ", Choices: []outcontents.Choice{{Value: "yes", Text: "Yes"}}}}, Enabled: true, Version: 4}
	definition, err := pollDefinition(request, 8)
	if err != nil {
		t.Fatal(err)
	}
	if definition.ID != 8 || definition.Title != "Poll" || definition.Headline != "Head" || len(definition.Questions) != 1 || !definition.Enabled || definition.Version != 4 {
		t.Fatalf("definition=%+v", definition)
	}
	invalid := []PollRequest{
		{Title: "missing questions"},
		{Title: "negative order", Questions: []outcontents.Question{{SortOrder: -1, Text: "x"}}},
		{Title: "duplicate", Questions: []outcontents.Question{{SortOrder: 1, Text: "x"}, {SortOrder: 1, Text: "y"}}},
		{Title: "type", Questions: []outcontents.Question{{SortOrder: 1, Type: 3, Text: "x"}}},
		{Title: "choices", Questions: []outcontents.Question{{SortOrder: 1, Type: 1, Text: "x"}}},
		{Title: "duplicate choices", Questions: []outcontents.Question{{SortOrder: 1, Type: 2, Text: "x", Choices: []outcontents.Choice{{Value: "a", Text: "A"}, {Value: "a", Text: "Again"}}}}},
	}
	for _, candidate := range invalid {
		if _, validationErr := pollDefinition(candidate, 1); validationErr == nil {
			t.Fatalf("accepted invalid poll %+v", candidate)
		}
	}
}

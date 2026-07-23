package settings

import (
	"errors"
	"strings"
	"testing"
	"unicode/utf8"

	botrecord "github.com/niflaot/pixels/internal/realm/bot/record"
)

// TestApplyChatClampsDelayAndTotalLength verifies native chat limits.
func TestApplyChatClampsDelayAndTotalLength(t *testing.T) {
	service := &Service{}
	bot := &botrecord.Bot{}
	data := strings.Repeat("á", 100) + "\r" + strings.Repeat("b", 100) + ";#;true;#;1;#;true"
	if err := service.applyChat(bot, data); err != nil {
		t.Fatalf("apply chat: %v", err)
	}
	if bot.ChatDelaySeconds != minChatDelay || !bot.ChatAuto || !bot.ChatRandom {
		t.Fatalf("unexpected settings %#v", bot)
	}
	total := 0
	for _, line := range bot.ChatLines {
		total += utf8.RuneCountInString(line)
	}
	if total != maxChatLength {
		t.Fatalf("expected %d runes, got %d", maxChatLength, total)
	}
}

// TestIdentityValidationRejectsMarkupAndBounds verifies name and motto policy.
func TestIdentityValidationRejectsMarkupAndBounds(t *testing.T) {
	service := &Service{}
	bot := &botrecord.Bot{}
	if err := service.applyName(bot, "<admin>"); !errors.Is(err, botrecord.ErrInvalidSkill) {
		t.Fatalf("name error=%v", err)
	}
	if err := service.applyName(bot, "Connie"); err != nil || bot.Name != "Connie" {
		t.Fatalf("name=%q err=%v", bot.Name, err)
	}
	if err := service.applyMotto(bot, strings.Repeat("x", maxMottoLength+1)); !errors.Is(err, botrecord.ErrInvalidSkill) {
		t.Fatalf("motto error=%v", err)
	}
}

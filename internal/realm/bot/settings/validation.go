package settings

import (
	"strconv"
	"strings"
	"unicode/utf8"

	botrecord "github.com/niflaot/pixels/internal/realm/bot/record"
)

// applyChat parses Nitro's lines, auto, delay, and random data format.
func (service *Service) applyChat(bot *botrecord.Bot, data string) error {
	if len(data) > maxChatInput {
		return botrecord.ErrInvalidSkill
	}
	parts := strings.Split(data, ";#;")
	if len(parts) < 4 {
		return botrecord.ErrInvalidSkill
	}
	delay, err := strconv.Atoi(parts[len(parts)-2])
	if err != nil {
		delay = minChatDelay
	}
	if delay < minChatDelay {
		delay = minChatDelay
	}
	if delay > maxChatDelay {
		delay = maxChatDelay
	}
	lines, total := make([]string, 0), 0
	for _, raw := range strings.Split(strings.Join(parts[:len(parts)-3], ";#;"), "\r") {
		line, converged := sanitizeFixedPoint(raw)
		if !converged {
			lines = nil
			break
		}
		if service.globalFilter != nil {
			line, _ = service.globalFilter.Censor(line)
		}
		remaining := maxChatLength - total
		if line == "" || remaining <= 0 {
			continue
		}
		line = truncateRunes(line, remaining)
		lines = append(lines, line)
		total += utf8.RuneCountInString(line)
	}
	bot.ChatLines, bot.ChatAuto, bot.ChatRandom, bot.ChatDelaySeconds = lines, parts[len(parts)-3] == "true", parts[len(parts)-1] == "true", delay
	return nil
}

// applyName validates visible bot identity against the global filter.
func (service *Service) applyName(bot *botrecord.Bot, value string) error {
	value = strings.TrimSpace(value)
	if value == "" || utf8.RuneCountInString(value) > maxNameLength || strings.ContainsAny(value, "<>") {
		return botrecord.ErrInvalidSkill
	}
	if service.globalFilter != nil {
		filtered, changed := service.globalFilter.Censor(value)
		if changed || !strings.EqualFold(filtered, value) {
			return botrecord.ErrInvalidSkill
		}
	}
	bot.Name = value
	return nil
}

// applyMotto validates and filters the visible bot motto.
func (service *Service) applyMotto(bot *botrecord.Bot, value string) error {
	value, converged := sanitizeFixedPoint(strings.TrimSpace(value))
	if !converged || utf8.RuneCountInString(value) > maxMottoLength {
		return botrecord.ErrInvalidSkill
	}
	if service.globalFilter != nil {
		value, _ = service.globalFilter.Censor(value)
	}
	bot.Motto = value
	return nil
}

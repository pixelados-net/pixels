package core

import (
	"context"
	"strings"
	"time"
	"unicode"

	botrecord "github.com/niflaot/pixels/internal/realm/bot/record"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
	sdkbot "github.com/niflaot/pixels/sdk/bot"
)

// OnPlaced invokes behavior placement and one configured localized greeting.
func (service *Service) OnPlaced(ctx context.Context, bot botrecord.Bot) {
	view := service.StaticView(bot)
	if behavior := service.behaviors.For(bot.BehaviorType); behavior != nil {
		if active, found := service.activeByID(view.RoomID, view.ID); found && active.behavior != nil {
			behavior = active.behavior
		}
		service.dispatch(func() {
			_ = behavior.OnPlace(context.WithoutCancel(ctx), view, service)
		})
	}
	keys := service.config.PlacementKeys()
	if len(keys) == 0 {
		return
	}
	key := keys[int(service.source.Uint64()%uint64(len(keys)))]
	message := key
	if service.translations != nil {
		message = service.translations.Default(i18n.Key(key), nil)
	}
	_ = service.Talk(ctx, view, message, sdkbot.ScopeTalk, 0)
}

// OnPickup invokes behavior cleanup asynchronously.
func (service *Service) OnPickup(ctx context.Context, bot botrecord.Bot) {
	view := service.StaticView(bot)
	behavior := service.behaviors.For(bot.BehaviorType)
	if active, found := service.activeByID(view.RoomID, view.ID); found && active.behavior != nil {
		behavior = active.behavior
	}
	if behavior == nil {
		return
	}
	service.dispatch(func() {
		_ = behavior.OnPickup(context.WithoutCancel(ctx), view, service)
	})
}

// StartFollowing assigns one player target while both entities share a room.
func (service *Service) StartFollowing(roomID int64, botID int64, playerID int64) error {
	bot, found := service.activeByID(roomID, botID)
	if !found {
		return botrecord.ErrBotNotFound
	}
	bot.mutex.Lock()
	bot.followingPlayerID = playerID
	bot.mutex.Unlock()
	return nil
}

// StopFollowing clears any bot following the supplied player.
func (service *Service) StopFollowing(roomID int64, playerID int64) {
	for _, bot := range service.roomBots(roomID) {
		bot.mutex.Lock()
		if bot.followingPlayerID == playerID {
			bot.followingPlayerID = 0
		}
		bot.mutex.Unlock()
	}
}

// HandleUserEnter records a room visit, syncs bots, and dispatches entry hooks.
func (service *Service) HandleUserEnter(ctx context.Context, roomID int64, playerID int64) error {
	if err := service.store.RecordVisit(ctx, roomID, playerID); err != nil {
		return err
	}
	if err := service.SyncPlayer(ctx, roomID, playerID); err != nil {
		return err
	}
	for _, bot := range service.roomBots(roomID) {
		bot.mutex.Lock()
		if bot.record.BehaviorType == botrecord.BehaviorVisitorLog && bot.record.OwnerPlayerID == playerID {
			if bot.visitorPrompted {
				bot.mutex.Unlock()
				continue
			}
			bot.visitorPrompted = true
		}
		view := service.StaticView(bot.record)
		behavior := bot.behavior
		bot.mutex.Unlock()
		if behavior != nil {
			service.dispatch(func() {
				_ = behavior.OnUserEnter(context.WithoutCancel(ctx), view, playerID, service)
			})
		}
	}
	return nil
}

// HandleUserSay dispatches filtered delivered player speech to every room bot.
func (service *Service) HandleUserSay(ctx context.Context, roomID int64, playerID int64, message string) {
	for _, bot := range service.roomBots(roomID) {
		bot.mutex.Lock()
		if bot.record.BehaviorType == botrecord.BehaviorVisitorLog && bot.visitorShown && playerID == bot.record.OwnerPlayerID {
			bot.mutex.Unlock()
			continue
		}
		view := service.StaticView(bot.record)
		behavior := bot.behavior
		if bot.record.BehaviorType == botrecord.BehaviorVisitorLog && playerID == bot.record.OwnerPlayerID && affirmativeMessage(message) {
			bot.visitorShown = true
		}
		bot.mutex.Unlock()
		if behavior != nil {
			input := sdkbot.Message{PlayerID: playerID, Text: message}
			service.dispatch(func() {
				_ = behavior.OnUserSay(context.WithoutCancel(ctx), view, input, service)
			})
		}
	}
}

// StaticView maps a durable placed bot into behavior identity.
func (service *Service) StaticView(bot botrecord.Bot) sdkbot.Bot {
	roomID := int64(0)
	if bot.RoomID != nil {
		roomID = *bot.RoomID
	}
	return sdkbot.Bot{ID: bot.ID, OwnerPlayerID: bot.OwnerPlayerID, RoomID: roomID, Name: bot.Name, BehaviorType: bot.BehaviorType, CanWalk: bot.CanWalk}
}

// affirmativeMessage recognizes baseline localized affirmative input.
func affirmativeMessage(value string) bool {
	return value == "yes" || value == "Yes" || value == "si" || value == "Si" || value == "sí" || value == "Sí"
}

// Publish emits one optional bot event.
func (service *Service) Publish(ctx context.Context, name bus.Name, payload any) {
	if service.events != nil {
		_ = service.events.Publish(ctx, bus.Event{Name: name, Payload: payload})
	}
}

// servingItems loads the small admin mapping cache once.
func (service *Service) servingItems(ctx context.Context) ([]botrecord.ServeItem, error) {
	service.mutex.RLock()
	if service.serveLoaded {
		items := service.serveItems
		service.mutex.RUnlock()
		return items, nil
	}
	service.mutex.RUnlock()
	items, err := service.store.ListServeItems(ctx)
	if err != nil {
		return nil, err
	}
	service.mutex.Lock()
	service.serveItems, service.serveLoaded = items, true
	service.mutex.Unlock()
	return items, nil
}

// InvalidateServeItems clears the immutable bartender mapping snapshot.
func (service *Service) InvalidateServeItems() {
	service.mutex.Lock()
	service.serveItems, service.serveLoaded = nil, false
	service.mutex.Unlock()
}

// containsWholeWord matches Unicode word boundaries without regular expressions.
func containsWholeWord(text string, keyword string) bool {
	textRunes, keywordRunes := []rune(strings.ToLower(text)), []rune(strings.ToLower(keyword))
	if len(keywordRunes) == 0 || len(keywordRunes) > len(textRunes) {
		return false
	}
	for start := 0; start+len(keywordRunes) <= len(textRunes); start++ {
		matched := true
		for offset := range len(keywordRunes) {
			if textRunes[start+offset] != keywordRunes[offset] {
				matched = false
				break
			}
		}
		if !matched {
			continue
		}
		left := start == 0 || !wordRune(textRunes[start-1])
		rightIndex := start + len(keywordRunes)
		right := rightIndex == len(textRunes) || !wordRune(textRunes[rightIndex])
		if left && right {
			return true
		}
	}
	return false
}

// wordRune reports Unicode letters, digits, and underscore as word content.
func wordRune(value rune) bool {
	return unicode.IsLetter(value) || unicode.IsDigit(value) || value == '_'
}

// Visits implements bounded SDK visitor history reads.
func (service *Service) Visits(ctx context.Context, view sdkbot.Bot, excludedPlayerID int64) ([]sdkbot.Visit, error) {
	since := time.Time{}
	if service.playerRecords != nil {
		owner, found, err := service.playerRecords.FindByID(ctx, view.OwnerPlayerID)
		if err != nil {
			return nil, err
		}
		if found && owner.Player.LastLogoutAt != nil {
			since = *owner.Player.LastLogoutAt
		}
	}
	visits, err := service.store.VisitsSince(ctx, view.RoomID, excludedPlayerID, since, 20)
	if err != nil {
		return nil, err
	}
	result := make([]sdkbot.Visit, len(visits))
	for index, visit := range visits {
		result[index] = sdkbot.Visit{PlayerID: visit.PlayerID, PlayerName: visit.PlayerName, EnteredAt: visit.EnteredAt}
	}
	return result, nil
}

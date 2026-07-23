package core

import (
	"context"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	botserveitem "github.com/niflaot/pixels/internal/realm/bot/events/serveitemrequested"
	botshouted "github.com/niflaot/pixels/internal/realm/bot/events/shouted"
	bottalked "github.com/niflaot/pixels/internal/realm/bot/events/talked"
	botwhispered "github.com/niflaot/pixels/internal/realm/bot/events/whispered"
	botrecord "github.com/niflaot/pixels/internal/realm/bot/record"
	chatconfig "github.com/niflaot/pixels/internal/realm/chat/config"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/networking/codec"
	outshout "github.com/niflaot/pixels/networking/outbound/chat/shout"
	outtalk "github.com/niflaot/pixels/networking/outbound/chat/talk"
	outwhisper "github.com/niflaot/pixels/networking/outbound/chat/whisper"
	outexpression "github.com/niflaot/pixels/networking/outbound/room/entities/expression"
	outhanditem "github.com/niflaot/pixels/networking/outbound/room/entities/handitem"
	"github.com/niflaot/pixels/pkg/i18n"
	sdkbot "github.com/niflaot/pixels/sdk/bot"
)

// Talk implements filtered SDK bot chat delivery.
func (service *Service) Talk(ctx context.Context, view sdkbot.Bot, message string, scope sdkbot.Scope, targetPlayerID int64) error {
	active, found := service.rooms.Find(view.RoomID)
	if !found {
		return botrecord.ErrRoomNotFound
	}
	unit, found := active.Unit(EntityKey(view.ID))
	if !found {
		return botrecord.ErrBotNotFound
	}
	message = service.resolveBehaviorMessage(message)
	message, consumed, err := service.speechInterceptor.Intercept(ctx, view, message, scope, targetPlayerID)
	if err != nil || consumed {
		return err
	}
	if service.globalFilter != nil {
		message, _ = service.globalFilter.Censor(message)
	}
	if service.roomFilter != nil {
		filtered, _, err := service.roomFilter.Censor(ctx, view.RoomID, message)
		if err == nil {
			message = filtered
		}
	}
	if message == "" {
		return nil
	}
	style := int32(0)
	if bot, found := service.activeByID(view.RoomID, view.ID); found {
		bot.mutex.Lock()
		style = bot.record.BubbleStyle
		bot.mutex.Unlock()
	}
	packet, err := botChatPacket(scope, unit.UnitID, message, style)
	if err != nil {
		return err
	}
	if scope == sdkbot.ScopeWhisper {
		service.sendPlayer(ctx, targetPlayerID, packet)
	} else if scope == sdkbot.ScopeTalk {
		service.sendTalkAudience(ctx, active, unit, packet)
	} else {
		_ = broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
	}
	if message == "o/" || message == "_o/" {
		if wave, encodeErr := outexpression.Encode(unit.UnitID, 1); encodeErr == nil {
			_ = broadcast.RoomPacket(ctx, service.connections, active, wave, 0)
		}
	}
	name := bottalked.Name
	if scope == sdkbot.ScopeShout {
		name = botshouted.Name
	} else if scope == sdkbot.ScopeWhisper {
		name = botwhispered.Name
		service.Publish(ctx, name, botwhispered.Payload{BotID: view.ID, RoomID: view.RoomID, Message: message, TargetPlayerID: targetPlayerID})
		return nil
	}
	if scope == sdkbot.ScopeShout {
		service.Publish(ctx, name, botshouted.Payload{BotID: view.ID, RoomID: view.RoomID, Message: message})
	} else {
		service.Publish(ctx, name, bottalked.Payload{BotID: view.ID, RoomID: view.RoomID, Message: message})
	}
	return nil
}

// sendTalkAudience delivers ordinary bot speech within the room chat radius.
func (service *Service) sendTalkAudience(ctx context.Context, active *roomlive.Room, speaker roomlive.UnitSnapshot, packet codec.Packet) {
	if service.connections == nil {
		return
	}
	distance := chatconfig.AudienceDistance(active.Snapshot().ChatDistance)
	maximum := distance * distance
	for _, presence := range active.Presences() {
		dx := int64(presence.Unit.Position.Point.X) - int64(speaker.Position.Point.X)
		dy := int64(presence.Unit.Position.Point.Y) - int64(speaker.Position.Point.Y)
		if dx*dx+dy*dy > maximum {
			continue
		}
		connection, found := service.connections.Get(presence.Occupant.ConnectionKind, presence.Occupant.ConnectionID)
		if found {
			_ = connection.Send(ctx, packet)
		}
	}
}

// botChatPacket creates the requested Nitro bot chat packet.
func botChatPacket(scope sdkbot.Scope, unitID int64, message string, style int32) (codec.Packet, error) {
	length := int32(utf8.RuneCountInString(message))
	switch scope {
	case sdkbot.ScopeShout:
		return outshout.Encode(int32(unitID), message, 0, style, length)
	case sdkbot.ScopeWhisper:
		return outwhisper.Encode(int32(unitID), message, 0, style, length)
	default:
		return outtalk.Encode(int32(unitID), message, 0, style, length)
	}
}

// resolveBehaviorMessage localizes built-in visitor behavior tokens.
func (service *Service) resolveBehaviorMessage(message string) string {
	if message == "bots.visitor.no_visits" && service.translations != nil {
		return service.translations.Default(i18n.Key(message), nil)
	}
	const prefix = "bots.visitor.visits:"
	if strings.HasPrefix(message, prefix) && service.translations != nil {
		return service.translations.Default("bots.visitor.visits", i18n.Params{"count": strings.TrimPrefix(message, prefix)})
	}
	return message
}

// ServeKeyword implements whole-word bartender matching and hand-item transfer.
func (service *Service) ServeKeyword(ctx context.Context, view sdkbot.Bot, message sdkbot.Message) (bool, error) {
	active, found := service.rooms.Find(view.RoomID)
	if !found {
		return false, botrecord.ErrRoomNotFound
	}
	botUnit, botFound := active.Unit(EntityKey(view.ID))
	playerUnit, playerFound := active.Unit(message.PlayerID)
	if !botFound || !playerFound || squaredDistance(botUnit.Position.Point, playerUnit.Position.Point) > service.config.BartenderCommandDistance*service.config.BartenderCommandDistance {
		return false, nil
	}
	items, err := service.servingItems(ctx)
	if err != nil {
		return false, err
	}
	for _, item := range items {
		if !containsWholeWord(message.Text, item.Keyword) || item.DefinitionID <= 0 || item.DefinitionID > math.MaxInt32 {
			continue
		}
		_, _ = active.FaceTo(EntityKey(view.ID), playerUnit.Position.Point)
		service.beginServe(active, view, message.PlayerID, int32(item.DefinitionID), item.Keyword)
		service.Publish(ctx, botserveitem.Name, botserveitem.Payload{BotID: view.ID, PlayerID: message.PlayerID, Keyword: item.Keyword, DefinitionID: item.DefinitionID})
		return true, nil
	}
	return false, nil
}

// beginServe gives the bot an item, approaches when necessary, and schedules transfer.
func (service *Service) beginServe(active *roomlive.Room, view sdkbot.Bot, playerID int64, itemID int32, keyword string) {
	botUnit, err := active.SetHandItem(EntityKey(view.ID), itemID)
	if err != nil {
		return
	}
	service.broadcastHandItem(context.Background(), active, botUnit, itemID)
	playerUnit, targetFound := active.Unit(playerID)
	if targetFound && squaredDistance(botUnit.Position.Point, playerUnit.Position.Point) > service.config.BartenderReachDistance*service.config.BartenderReachDistance {
		if goal, valid := rotatedNeighbor(playerUnit.Position.Point, (int(playerUnit.BodyRotation)+4)%8); valid {
			_, _ = active.MoveTo(EntityKey(view.ID), goal)
		}
	}
	service.awaitServe(active, view, playerID, itemID, keyword, 20)
}

// awaitServe polls from the room-owned scheduler until movement settles or expires.
func (service *Service) awaitServe(active *roomlive.Room, view sdkbot.Bot, playerID int64, itemID int32, keyword string, remaining int) {
	active.Schedule(roomlive.DefaultTickInterval, func(time.Time) {
		botUnit, botFound := active.Unit(EntityKey(view.ID))
		playerUnit, playerFound := active.Unit(playerID)
		if !botFound || !playerFound || remaining <= 0 {
			if botFound {
				service.clearBotHand(active, botUnit)
			}
			return
		}
		if botUnit.Moving || squaredDistance(botUnit.Position.Point, playerUnit.Position.Point) > service.config.BartenderReachDistance*service.config.BartenderReachDistance {
			service.awaitServe(active, view, playerID, itemID, keyword, remaining-1)
			return
		}
		received, err := active.SetHandItem(playerID, itemID)
		if err == nil {
			service.broadcastHandItem(context.Background(), active, received, itemID)
		}
		service.clearBotHand(active, botUnit)
		message := "Here is your " + keyword + "."
		if service.translations != nil {
			message = service.translations.Default("bots.bartender.given", i18n.Params{"item": keyword})
		}
		_ = service.Talk(context.Background(), view, message, sdkbot.ScopeTalk, 0)
	})
}

// clearBotHand clears and broadcasts the bot hand item.
func (service *Service) clearBotHand(active *roomlive.Room, unit roomlive.UnitSnapshot) {
	cleared, err := active.SetHandItem(unit.EntityKey, 0)
	if err == nil {
		service.broadcastHandItem(context.Background(), active, cleared, 0)
	}
}

// broadcastHandItem broadcasts one unit hand item update.
func (service *Service) broadcastHandItem(ctx context.Context, active *roomlive.Room, unit roomlive.UnitSnapshot, itemID int32) {
	packet, err := outhanditem.Encode(unit.UnitID, itemID)
	if err == nil {
		_ = broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
	}
}

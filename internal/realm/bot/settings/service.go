// Package settings owns owner-configurable bot skill workflows.
package settings

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	botbehavior "github.com/niflaot/pixels/internal/realm/bot/behavior"
	botcore "github.com/niflaot/pixels/internal/realm/bot/core"
	botchatsaved "github.com/niflaot/pixels/internal/realm/bot/events/chatsaved"
	botlooksaved "github.com/niflaot/pixels/internal/realm/bot/events/looksaved"
	botnamesaved "github.com/niflaot/pixels/internal/realm/bot/events/namesaved"
	botpolicy "github.com/niflaot/pixels/internal/realm/bot/policy"
	botrecord "github.com/niflaot/pixels/internal/realm/bot/record"
	chatfilter "github.com/niflaot/pixels/internal/realm/chat/filter"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	sdkbot "github.com/niflaot/pixels/sdk/bot"
)

// Service coordinates validated bot settings persistence.
type Service struct {
	// store persists bots.
	store botrecord.Store
	// rooms resolves active room projection.
	rooms *roomlive.Registry
	// players resolves the configuring player's current look.
	players *playerlive.Registry
	// permissions resolves management bypasses.
	permissions permissionservice.Checker
	// behaviors resolves custom skill handlers.
	behaviors *botbehavior.Registry
	// globalFilter moderates configured player text.
	globalFilter *chatfilter.Service
	// runtime updates active bot generations and protocol projection.
	runtime *botcore.Service
}

// New creates bot settings behavior.
func New(store botrecord.Store, rooms *roomlive.Registry, players *playerlive.Registry, permissions permissionservice.Checker, behaviors *botbehavior.Registry, globalFilter *chatfilter.Service, runtime *botcore.Service) *Service {
	return &Service{store: store, rooms: rooms, players: players, permissions: permissions, behaviors: behaviors, globalFilter: globalFilter, runtime: runtime}
}

const (
	// SkillGeneric identifies the generic context menu entry.
	SkillGeneric int32 = iota
	// SkillLook copies the configuring player's look.
	SkillLook
	// SkillChat replaces automatic chat configuration.
	SkillChat
	// SkillWalk toggles random walking.
	SkillWalk
	// SkillDance cycles the bot dance.
	SkillDance
	// SkillName replaces the visible name.
	SkillName
	// SkillServe identifies bartender serving.
	SkillServe
	// SkillLink identifies an in-client link.
	SkillLink
	// SkillProceed identifies new-user progression.
	SkillProceed
	// SkillMotto replaces the visible motto.
	SkillMotto
)

const (
	// maxChatInput bounds the encoded Nitro configuration payload.
	maxChatInput = 5112
	// maxChatLength bounds all configured lines combined.
	maxChatLength = 120
	// maxNameLength bounds visible bot names.
	maxNameLength = 15
	// maxMottoLength bounds visible bot mottos.
	maxMottoLength = 38
	// minChatDelay stores the minimum automatic delay.
	minChatDelay = 7
	// maxChatDelay stores the maximum automatic delay.
	maxChatDelay = 604800
)

// SaveSkill validates and persists one native bot configuration command.
func (service *Service) SaveSkill(ctx context.Context, roomID int64, botID int64, actorID int64, skillID int32, data string) (botrecord.Bot, error) {
	bot, found, err := service.store.Find(ctx, botID)
	if err != nil || !found || bot.RoomID == nil || *bot.RoomID != roomID {
		return botrecord.Bot{}, firstError(err, botrecord.ErrBotNotFound)
	}
	allowed := bot.OwnerPlayerID == actorID
	if !allowed {
		allowed, err = service.has(ctx, actorID, botpolicy.AnyRoomOwner)
	}
	if err != nil || !allowed {
		return botrecord.Bot{}, firstError(err, botrecord.ErrNoRights)
	}
	if err = service.applySkill(ctx, &bot, actorID, skillID, data); err != nil {
		return botrecord.Bot{}, err
	}
	saved, found, err := service.store.Save(ctx, bot)
	if err != nil || !found {
		return botrecord.Bot{}, firstError(err, botrecord.ErrConflict)
	}
	service.runtime.ReplacePlaced(saved)
	if active, activeFound := service.rooms.Find(roomID); activeFound {
		service.runtime.ProjectSpawn(ctx, active, saved)
	}
	service.publishSkillSaved(ctx, saved, actorID, skillID)
	return saved, nil
}

// publishSkillSaved emits the three plugin-compatible settings events.
func (service *Service) publishSkillSaved(ctx context.Context, bot botrecord.Bot, actorID int64, skillID int32) {
	switch skillID {
	case SkillLook:
		service.runtime.Publish(ctx, botlooksaved.Name, botlooksaved.Payload{BotID: bot.ID, PlayerID: actorID})
	case SkillChat:
		service.runtime.Publish(ctx, botchatsaved.Name, botchatsaved.Payload{BotID: bot.ID, PlayerID: actorID, LineCount: len(bot.ChatLines)})
	case SkillName:
		service.runtime.Publish(ctx, botnamesaved.Name, botnamesaved.Payload{BotID: bot.ID, PlayerID: actorID, Name: bot.Name})
	}
}

// applySkill mutates one validated durable snapshot.
func (service *Service) applySkill(ctx context.Context, bot *botrecord.Bot, actorID int64, skillID int32, data string) error {
	switch skillID {
	case SkillLook:
		player, found := service.players.Find(actorID)
		if !found {
			return botrecord.ErrInvalidSkill
		}
		snapshot := player.Snapshot()
		bot.Figure, bot.Gender = snapshot.Look, string(snapshot.Gender)
		bot.EffectID = snapshot.ActiveEffectID
	case SkillChat:
		return service.applyChat(bot, data)
	case SkillWalk:
		bot.CanWalk = !bot.CanWalk
	case SkillDance:
		bot.DanceType = (bot.DanceType + 1) % 5
	case SkillName:
		return service.applyName(bot, data)
	case SkillMotto:
		return service.applyMotto(bot, data)
	default:
		behavior := service.behaviors.For(bot.BehaviorType)
		if behavior == nil {
			return botrecord.ErrInvalidSkill
		}
		err := behavior.SaveCustomSkill(ctx, service.runtime.StaticView(*bot), skillID, data)
		if errors.Is(err, sdkbot.ErrUnsupportedSkill) {
			return botrecord.ErrInvalidSkill
		}
		return err
	}
	return nil
}

// ConfigurationData returns Nitro data for one requested skill.
func (service *Service) ConfigurationData(ctx context.Context, roomID int64, botID int64, actorID int64, skillID int32) (string, error) {
	bot, found, err := service.store.Find(ctx, botID)
	if err != nil || !found || bot.RoomID == nil || *bot.RoomID != roomID {
		return "", firstError(err, botrecord.ErrBotNotFound)
	}
	if bot.OwnerPlayerID != actorID {
		allowed, checkErr := service.has(ctx, actorID, botpolicy.AnyRoomOwner)
		if checkErr != nil || !allowed {
			return "", firstError(checkErr, botrecord.ErrNoRights)
		}
	}
	switch skillID {
	case SkillChat:
		return strings.Join(bot.ChatLines, "\r") + ";#;" + strconv.FormatBool(bot.ChatAuto) + ";#;" + strconv.Itoa(bot.ChatDelaySeconds) + ";#;" + strconv.FormatBool(bot.ChatRandom), nil
	case SkillName:
		return bot.Name, nil
	case SkillMotto:
		return bot.Motto, nil
	default:
		return "", nil
	}
}

// has resolves one optional permission checker.
func (service *Service) has(ctx context.Context, playerID int64, node permission.Node) (bool, error) {
	if service.permissions == nil {
		return false, nil
	}
	return service.permissions.HasPermission(ctx, playerID, node)
}

// firstError chooses infrastructure failures before domain failures.
func firstError(err error, fallback error) error {
	if err != nil {
		return err
	}
	return fallback
}

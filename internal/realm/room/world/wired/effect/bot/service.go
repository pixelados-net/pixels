// Package bot executes WIRED bot effects through the existing bot runtime.
package bot

import (
	"context"

	botcore "github.com/niflaot/pixels/internal/realm/bot/core"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
	sdkbot "github.com/niflaot/pixels/sdk/bot"
)

// Service adapts canonical bot operations to the room bot realm.
type Service struct {
	// rooms resolves active room worlds.
	rooms *roomlive.Registry
	// bots resolves and controls active bots.
	bots *botcore.Service
}

// New creates a WIRED bot effect service.
func New(rooms *roomlive.Registry, bots *botcore.Service) *Service {
	return &Service{rooms: rooms, bots: bots}
}

// ExecuteBot executes one validated bot operation.
func (service *Service) ExecuteBot(ctx context.Context, operation effect.BotOperation, node *configuration.Node, event trigger.Event) (effect.Result, error) {
	if service.bots == nil {
		return effect.Result{Status: effect.Blocked}, nil
	}
	active, found := service.rooms.Find(event.RoomID)
	if !found {
		return effect.Result{Status: effect.Skipped}, nil
	}
	view, found := service.bots.ResolveByName(event.RoomID, node.Parameters.Name)
	if !found {
		return effect.Result{Status: effect.Skipped}, nil
	}
	switch operation {
	case effect.BotTalk:
		return applied(service.bots.Talk(ctx, view, node.Parameters.Message, sdkbot.ScopeTalk, 0))
	case effect.BotTalkToAvatar:
		if event.PlayerID <= 0 {
			return effect.Result{Status: effect.Skipped}, nil
		}
		return applied(service.bots.Talk(ctx, view, node.Parameters.Message, sdkbot.ScopeWhisper, event.PlayerID))
	case effect.BotFollowAvatar:
		if event.PlayerID <= 0 {
			return effect.Result{Status: effect.Skipped}, nil
		}
		return applied(service.bots.StartFollowing(event.RoomID, view.ID, event.PlayerID))
	case effect.BotGiveHanditem:
		if event.PlayerID <= 0 || len(node.Parameters.Values) == 0 {
			return effect.Result{Status: effect.Skipped}, nil
		}
		_, err := active.SetHandItem(event.PlayerID, node.Parameters.Values[0])
		return applied(err)
	case effect.BotMove:
		return service.move(active, view, node, false)
	case effect.BotTeleport:
		return service.move(active, view, node, true)
	case effect.BotClothes:
		if node.Parameters.Message == "" {
			return effect.Result{Status: effect.Blocked}, nil
		}
		return applied(service.bots.ChangeFigure(ctx, event.RoomID, view.ID, node.Parameters.Message))
	default:
		return effect.Result{Status: effect.Blocked}, nil
	}
}

// move moves or teleports a bot onto the first valid target.
func (service *Service) move(active *roomlive.Room, view sdkbot.Bot, node *configuration.Node, teleport bool) (effect.Result, error) {
	for _, target := range node.Targets {
		item, found := active.FurnitureItem(target.ItemID)
		if !found {
			continue
		}
		key := botcore.EntityKey(view.ID)
		if teleport {
			unit, found := active.UnitMotion(key)
			if !found {
				return effect.Result{Status: effect.Skipped}, nil
			}
			_, err := active.TeleportUnit(key, item.Point, unit.BodyRotation, false, roomlive.TeleportNear)
			if err != nil {
				return effect.Result{Status: effect.Blocked}, err
			}
			return effect.Result{Status: effect.Applied, Derived: []trigger.Event{{
				Kind: trigger.BotReachedFurniture, RoomID: view.RoomID, ActorKind: trigger.ActorBot,
				ActorID: key, Username: view.Name, SourceItem: item.ID, SourceSprite: int32(item.Definition.SpriteID),
			}}}, nil
		}
		_, err := active.MoveControlled(key, item.Point, worldunit.ControlFurnitureInteraction)
		return applied(err)
	}
	return effect.Result{Status: effect.Skipped}, nil
}

// applied maps an operation error to an effect result.
func applied(err error) (effect.Result, error) {
	if err != nil {
		return effect.Result{Status: effect.Blocked}, err
	}
	return effect.Result{Status: effect.Applied}, nil
}

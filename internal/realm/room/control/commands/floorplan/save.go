package floorplan

import (
	"context"
	"strconv"
	"time"

	"github.com/niflaot/pixels/internal/command"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/control/commands/resolve"
	domain "github.com/niflaot/pixels/internal/realm/room/control/floorplan"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	roomitems "github.com/niflaot/pixels/internal/realm/room/world/items"
	roomlayout "github.com/niflaot/pixels/internal/realm/room/world/layout"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/zap"
)

const (
	// SaveName identifies the floor plan save command.
	SaveName command.Name = "room.floorplan.save"
	// cooldownPrefix namespaces floor plan save throttles.
	cooldownPrefix = "floorplan:cooldown:"
)

// CooldownStore atomically acquires and releases floor plan cooldowns.
type CooldownStore interface {
	// SetIfAbsent writes a cooldown only when none exists.
	SetIfAbsent(context.Context, string, []byte, time.Duration) (bool, error)
	// Delete removes a cooldown key.
	Delete(context.Context, string) error
}

// SaveCommand requests a custom floor plan mutation.
type SaveCommand struct {
	// Handler stores the source connection handler.
	Handler netconn.Context
	// Params stores editable floor plan input.
	Params domain.SaveParams
}

// SaveHandler handles custom floor plan saves.
type SaveHandler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores connection bindings.
	Bindings *binding.Registry
	// Rooms reads room records.
	Rooms RoomFinder
	// Layouts resolves and saves room-owned layouts.
	Layouts roomlayout.RoomManager
	// Furniture manages persistent furniture.
	Furniture furnitureservice.Manager
	// Runtime stores active rooms.
	Runtime *roomlive.Registry
	// Connections stores active connections.
	Connections *netconn.Registry
	// Authorize resolves floor plan capability.
	Authorize *domain.Authorizer
	// Cooldowns stores distributed save throttles.
	Cooldowns CooldownStore
	// Config stores floor plan limits.
	Config domain.Config
	// Events publishes committed floor plan events.
	Events bus.Publisher
	// Translations resolves end-user messages.
	Translations i18n.Translator
	// Log records rejected and committed saves.
	Log *zap.Logger
}

// CommandName returns the stable command name.
func (SaveCommand) CommandName() command.Name { return SaveName }

// Handle validates, commits, and projects a floor plan save.
func (handler SaveHandler) Handle(ctx context.Context, envelope command.Envelope[SaveCommand]) error {
	input := envelope.Command
	player, roomID, err := control.Actor(input.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	room, found, err := handler.Rooms.FindByID(ctx, roomID)
	if err != nil || !found {
		return err
	}
	if err = handler.Authorize.Authorize(ctx, room, player.ID()); err != nil {
		return sendError(ctx, input.Handler, err, handler.Translations)
	}
	release, acquired, err := handler.acquireCooldown(ctx, player.ID())
	if err != nil {
		return err
	}
	if !acquired {
		return sendError(ctx, input.Handler, domain.ValidationErrors{Codes: []domain.ErrorCode{domain.CodeCooldown}}, handler.Translations)
	}
	committed := false
	defer func() {
		if !committed {
			release()
		}
	}()
	validated, err := domain.Validate(handler.Config, input.Params)
	if err != nil {
		return sendError(ctx, input.Handler, err, handler.Translations)
	}
	previous, err := handler.Layouts.ResolveForRoom(ctx, room.ID, room.ModelName)
	if err != nil {
		return err
	}
	oldGrid, err := previous.Grid()
	if err != nil {
		return err
	}
	active, activeFound := handler.Runtime.Find(room.ID)
	blocked := blockedFurniture(active, oldGrid, validated.Grid)
	if len(blocked) > 0 && !validated.Params.AutoPickup {
		return sendError(ctx, input.Handler, domain.ValidationErrors{Codes: []domain.ErrorCode{domain.CodeBlockedFurniture}}, handler.Translations)
	}
	if len(blocked) > domain.MaxAutoPickupItems {
		return sendError(ctx, input.Handler, domain.ValidationErrors{Codes: []domain.ErrorCode{domain.CodeTooManyPickups}}, handler.Translations)
	}
	saved, picked, err := handler.commit(ctx, room.ID, validated, blocked)
	if err != nil {
		return err
	}
	committed = true
	if activeFound {
		if err = handler.reload(ctx, room, active, saved); err != nil {
			return err
		}
	}
	handler.notifyPicked(ctx, picked)
	if handler.Log != nil {
		handler.Log.Debug("room floor plan saved", zap.Int64("room_id", room.ID), zap.Int64("player_id", player.ID()), zap.Int("picked_items", len(picked)))
	}

	return handler.publish(ctx, room.ID, player.ID())
}

// acquireCooldown atomically reserves one player's save interval.
func (handler SaveHandler) acquireCooldown(ctx context.Context, playerID int64) (func(), bool, error) {
	if handler.Cooldowns == nil {
		return func() {}, true, nil
	}
	key := cooldownPrefix + strconv.FormatInt(playerID, 10)
	acquired, err := handler.Cooldowns.SetIfAbsent(ctx, key, []byte{1}, handler.Config.Normalize().SaveCooldown)

	return func() { _ = handler.Cooldowns.Delete(context.Background(), key) }, acquired, err
}

// blockedFurniture resolves runtime furniture affected by the requested geometry.
func blockedFurniture(active *roomlive.Room, previous grid.Grid, next grid.Grid) []domain.BlockedItem {
	if active == nil {
		return nil
	}

	return domain.BlockedItems(previous, next, active.FurnitureItems())
}

// commit persists optional pickups and geometry in one transaction scope.
func (handler SaveHandler) commit(ctx context.Context, roomID int64, validated domain.Validated, blocked []domain.BlockedItem) (roomlayout.Layout, []furnituremodel.Item, error) {
	var saved roomlayout.Layout
	var picked []furnituremodel.Item
	err := handler.Layouts.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error
		if validated.Params.AutoPickup && len(blocked) > 0 {
			picked, err = handler.pickupBlocked(txCtx, roomID, blocked)
			if err != nil {
				return err
			}
		}
		point, _ := grid.NewPoint(validated.Params.DoorX, validated.Params.DoorY)
		doorZ, _ := validated.Grid.HeightAt(point)
		saved, err = handler.Layouts.SaveCustom(txCtx, roomlayout.CustomSaveParams{
			RoomID: roomID, Heightmap: validated.Params.Heightmap,
			DoorX: validated.Params.DoorX, DoorY: validated.Params.DoorY, DoorZ: int(doorZ),
			DoorDirection: validated.Params.DoorDirection,
			WallThickness: validated.Params.WallThickness, FloorThickness: validated.Params.FloorThickness,
			WallHeight: validated.Params.WallHeight,
		})

		return err
	})

	return saved, picked, err
}

// pickupBlocked returns affected furniture to each item's actual owner.
func (handler SaveHandler) pickupBlocked(ctx context.Context, roomID int64, blocked []domain.BlockedItem) ([]furnituremodel.Item, error) {
	records, err := handler.Furniture.ListRoomItems(ctx, roomID)
	if err != nil {
		return nil, err
	}
	blockedIDs := make(map[int64]struct{}, len(blocked))
	for _, item := range blocked {
		blockedIDs[item.Item.ID] = struct{}{}
	}
	picked := make([]furnituremodel.Item, 0, len(blockedIDs))
	for _, item := range records {
		if _, found := blockedIDs[item.ID]; !found {
			continue
		}
		value, err := handler.Furniture.Pickup(ctx, furnitureservice.PickupParams{ItemID: item.ID, ActorPlayerID: item.OwnerPlayerID})
		if err != nil {
			return nil, err
		}
		picked = append(picked, value)
	}

	return picked, nil
}

// loadWorldFurniture reads current furniture after transactional pickups.
func (handler SaveHandler) loadWorldFurniture(ctx context.Context, roomID int64) ([]worldfurniture.Item, error) {
	return roomitems.LoadRoomFurniture(ctx, handler.Furniture, roomID)
}

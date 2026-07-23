package games

import (
	"context"
	"strconv"
	"strings"

	furnituremoved "github.com/niflaot/pixels/internal/realm/furniture/events/moved"
	furniturewalkedoff "github.com/niflaot/pixels/internal/realm/furniture/events/walkedoff"
	furniturewalkedon "github.com/niflaot/pixels/internal/realm/furniture/events/walkedon"
	roomleft "github.com/niflaot/pixels/internal/realm/room/access/events/left"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roommoved "github.com/niflaot/pixels/internal/realm/room/world/events/moved"
	"github.com/niflaot/pixels/internal/realm/room/world/games/banzai"
	gameprogressed "github.com/niflaot/pixels/internal/realm/room/world/games/events/progressed"
	"github.com/niflaot/pixels/internal/realm/room/world/games/tag"
	"github.com/niflaot/pixels/networking/outbound/furniture/stuffdata"
	outobjectbatch "github.com/niflaot/pixels/networking/outbound/room/furniture/objectdata/batch"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
)

// stateUpdate stores one projected furniture multi-state value.
type stateUpdate struct {
	// id identifies the projected furniture item.
	id int64
	// value stores the authoritative state.
	value int
}

// scheduleBanzaiTeleport delegates one room-owned random teleport.
func (service *Service) scheduleBanzaiTeleport(active *roomlive.Room, playerID int64, sourceID int64) error {
	return banzai.ScheduleTeleport(service.connections, active, playerID, sourceID)
}

// Register attaches furniture events and room-owned lifecycle callbacks.
func Register(lifecycle fx.Lifecycle, subscriber bus.Subscriber, rooms *roomlive.Registry, service *Service) error {
	registrations := []struct {
		// name identifies one subscribed event.
		name bus.Name
		// handler processes the event on its owning room runtime.
		handler bus.Handler
	}{{furniturewalkedon.Name, service.WalkedOn}, {furniturewalkedoff.Name, service.WalkedOff}, {furnituremoved.Name, service.FurnitureMoved}, {roommoved.Name, service.UnitMoved}, {roomleft.Name, service.PlayerLeft}}
	subscriptions := make([]*bus.Subscription, 0, len(registrations))
	for _, registration := range registrations {
		subscription, err := subscriber.Subscribe(registration.name, bus.PriorityHigh, registration.handler)
		if err != nil {
			for _, active := range subscriptions {
				active.Unsubscribe()
			}
			return err
		}
		subscriptions = append(subscriptions, subscription)
	}
	rooms.AddCyclePublisher(service.Cycle)
	rooms.AddClosePublisher(service.Close)
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
		for _, subscription := range subscriptions {
			subscription.Unsubscribe()
		}
		return nil
	}})
	return nil
}

// WalkedOn handles team gates, Banzai tiles, and Tag fields.
func (service *Service) WalkedOn(ctx context.Context, event bus.Event) error {
	payload, ok := event.Payload.(furniturewalkedon.Payload)
	if !ok || !service.config.Enabled {
		return nil
	}
	active, found := service.rooms.Find(payload.RoomID)
	if !found {
		return nil
	}
	item, found := active.FurnitureItem(payload.ItemID)
	if !found {
		return nil
	}
	kind := item.Definition.InteractionType
	if kind == "football_gate" {
		return service.toggleFootballKit(ctx, active, payload.PlayerID, item.Definition.CustomParams)
	}
	if kind == "battlebanzai_random_teleport" {
		return service.scheduleBanzaiTeleport(active, payload.PlayerID, item.ID)
	}
	if kind == "freeze_block" {
		return service.collectFreezePowerUp(ctx, active, payload.PlayerID, item.ID)
	}
	if team, gate := gateTeam(kind); gate {
		if snapshot, exists := service.wired.Snapshot(payload.RoomID); exists && snapshot.Running {
			return nil
		}
		current, joined := service.wired.Team(payload.RoomID, payload.PlayerID)
		changed := false
		if joined && current == team {
			changed = service.wired.LeaveTeam(payload.RoomID, payload.PlayerID)
		} else {
			changed = service.wired.JoinTeam(payload.RoomID, payload.PlayerID, team)
		}
		if changed {
			return service.sendPlaying(ctx, active, payload.PlayerID, !(joined && current == team))
		}
		return nil
	}
	if strings.HasSuffix(kind, "_field") {
		variant := tagVariant(kind)
		service.mutex.Lock()
		state := service.stateLocked(active)
		joined := state.tags[variant].Join(payload.PlayerID)
		service.mutex.Unlock()
		if joined {
			service.projectTag(active, variant)
			key := "game.tag.placed"
			if variant == tag.Rollerskate {
				key = "game.rollerskate.placed"
			}
			service.progress(ctx, payload.PlayerID, key, 1)
			return service.sendPlaying(ctx, active, payload.PlayerID, true)
		}
		return nil
	}
	if kind == "battlebanzai_tile" {
		return service.stepBanzai(ctx, active, payload.PlayerID, payload.ItemID)
	}
	return nil
}

// stepBanzai advances one tile and captures enclosed areas.
func (service *Service) stepBanzai(ctx context.Context, active *roomlive.Room, playerID int64, itemID int64) error {
	team, found := service.wired.Team(active.ID(), playerID)
	if !found {
		return nil
	}
	snapshot, found := service.wired.Snapshot(active.ID())
	if !found || !snapshot.Running {
		return nil
	}
	service.mutex.Lock()
	state := service.stateLocked(active)
	if state.board == nil {
		service.mutex.Unlock()
		return nil
	}
	index := -1
	for candidate, id := range state.tileItems {
		if id == itemID {
			index = candidate
			break
		}
	}
	if index < 0 {
		service.mutex.Unlock()
		return nil
	}
	points, locked := state.board.Tiles[index].Step(uint8(team), service.config.Banzai.PointsSteal, service.config.Banzai.PointsLock)
	visual := state.board.Tiles[index].State()
	captured := []int(nil)
	if locked {
		captured = append(captured, state.board.CaptureLargest(uint8(team))...)
		service.metrics.tilesLocked.Add(uint64(1 + len(captured)))
	}
	complete := state.board.Complete()
	updates := make([]stateUpdate, 0, len(captured)+1)
	updates = append(updates, stateUpdate{itemID, visual})
	for _, capturedIndex := range captured {
		if id := state.tileItems[capturedIndex]; id > 0 {
			updates = append(updates, stateUpdate{id, state.board.Tiles[capturedIndex].State()})
		}
	}
	service.mutex.Unlock()
	totalPoints := points + len(captured)*service.config.Banzai.PointsFill
	if totalPoints != 0 {
		service.coordinator.AddScore(ctx, active.ID(), playerID, int64(totalPoints))
	}
	if locked {
		service.progress(ctx, playerID, "game.banzai.tile.locked", int64(1+len(captured)))
	}
	if score, exists := service.wired.Snapshot(active.ID()); exists {
		leader := winningTeam(score)
		for _, sphere := range active.FurnitureByInteraction("battlebanzai_sphere") {
			if err := service.projectState(ctx, active, sphere.ID, int(leader)); err != nil {
				return err
			}
		}
	}
	if len(updates) == 1 {
		if err := service.projectState(ctx, active, updates[0].id, updates[0].value); err != nil {
			return err
		}
	} else if err := service.projectStatesBatch(ctx, active, updates); err != nil {
		return err
	}
	if complete {
		return service.finish(ctx, active, state.startedAt)
	}
	return nil
}

// projectStatesBatch updates authoritative state and broadcasts objectdata batch 1453.
func (service *Service) projectStatesBatch(ctx context.Context, active *roomlive.Room, updates []stateUpdate) error {
	ids := make([]int64, 0, len(updates))
	data := make([]*stuffdata.Data, 0, len(updates))
	for _, update := range updates {
		value := strconv.Itoa(update.value)
		active.SetFurnitureExtraData(update.id, value)
		ids, data = append(ids, update.id), append(data, stuffdata.Legacy(value))
	}
	packet, err := outobjectbatch.Encode(ids, data)
	if err != nil {
		return err
	}
	return broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
}

// progress publishes one committed room game delta.
func (service *Service) progress(ctx context.Context, playerID int64, key string, amount int64) {
	if service.events == nil || playerID <= 0 || key == "" || amount <= 0 {
		return
	}
	_ = service.events.Publish(ctx, bus.Event{Name: gameprogressed.Name, Payload: gameprogressed.Payload{PlayerID: playerID, Key: key, Amount: amount}})
}

// gateTeam resolves the four canonical gate colors.
func gateTeam(kind string) (int32, bool) {
	colors := []string{"_r", "_g", "_b", "_y"}
	if !strings.Contains(kind, "_gate_") {
		return 0, false
	}
	for index, suffix := range colors {
		if strings.HasSuffix(kind, suffix) {
			return int32(index + 1), true
		}
	}
	return 0, false
}

// tagVariant maps furniture field kinds to one arena variant.
func tagVariant(kind string) tag.Variant {
	if strings.HasPrefix(kind, "rollerskate_") {
		return tag.Rollerskate
	}
	if strings.HasPrefix(kind, "bunnyrun_") {
		return tag.Bunnyrun
	}
	return tag.IceTag
}

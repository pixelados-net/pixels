package mysterybox

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	chatfilter "github.com/niflaot/pixels/internal/realm/chat/filter"
	essential "github.com/niflaot/pixels/internal/realm/furniture/interactions/essential"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	roomtask "github.com/niflaot/pixels/internal/realm/room/runtime/live/task"
	netconn "github.com/niflaot/pixels/networking/connection"
	outcancel "github.com/niflaot/pixels/networking/outbound/furniture/mysterybox/cancel"
	outprize "github.com/niflaot/pixels/networking/outbound/furniture/mysterybox/prize"
	outwait "github.com/niflaot/pixels/networking/outbound/furniture/mysterybox/wait"
)

// Service coordinates box reveal cancellation and trophy inscriptions.
type Service struct {
	config    Config
	store     Store
	furniture furnitureservice.DefinitionGranter
	states    furnitureservice.StateUpdater
	filter    *chatfilter.Service
	mutex     sync.Mutex
	pending   map[int64]uint64
	sequence  uint64
}

// New creates mystery-box behavior.
func New(config Config, store Store, furniture furnitureservice.DefinitionGranter, states furnitureservice.StateUpdater, filter *chatfilter.Service) *Service {
	return &Service{config: config.Normalize(), store: store, furniture: furniture, states: states, filter: filter, pending: make(map[int64]uint64)}
}

// UseFurniture opens a mystery box through room-owned scheduled work.
func (service *Service) UseFurniture(ctx context.Context, request essential.Request) (bool, error) {
	if request.Item.Definition.InteractionType != "mystery_box" {
		return false, nil
	}
	wait, err := outwait.Encode()
	if err != nil {
		return true, err
	}
	if err = request.Target.Send(ctx, wait); err != nil {
		return true, err
	}
	sequence := service.start(request.PlayerID)
	async := context.WithoutCancel(ctx)
	request.Room.ScheduleReplacing(roomTaskKey(request.PlayerID), service.config.Wait, func(_ time.Time) { service.resolve(async, request, sequence) })
	return true, nil
}

// Cancel invalidates one player's pending reveal and closes its dialog.
func (service *Service) Cancel(ctx context.Context, playerID int64, connection netconn.Context) error {
	service.mutex.Lock()
	delete(service.pending, playerID)
	service.mutex.Unlock()
	packet, err := outcancel.Encode()
	if err != nil {
		return err
	}
	return connection.Send(ctx, packet)
}

// Inscribe stores one filtered permanent mystery-trophy inscription.
func (service *Service) Inscribe(ctx context.Context, playerID int64, roomID int64, itemID int64, current string, text string) (string, error) {
	if service.filter != nil {
		text, _ = service.filter.Censor(text)
	}
	encoded, err := json.Marshal(map[string]any{"author": playerID, "text": text})
	if err != nil {
		return "", err
	}
	_, err = service.states.UpdateState(ctx, furnitureservice.StateParams{ItemID: itemID, RoomID: roomID, Expected: current, Next: string(encoded)})
	return string(encoded), err
}

// Keys returns one account's tracker state.
func (service *Service) Keys(ctx context.Context, playerID int64) (Keys, error) {
	return service.store.FindKeys(ctx, playerID)
}

// start creates one opaque reveal generation.
func (service *Service) start(playerID int64) uint64 {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	service.sequence++
	service.pending[playerID] = service.sequence
	return service.sequence
}

// resolve grants and projects one still-pending reveal.
func (service *Service) resolve(ctx context.Context, request essential.Request, sequence uint64) {
	service.mutex.Lock()
	current, found := service.pending[request.PlayerID]
	if found && current == sequence {
		delete(service.pending, request.PlayerID)
	}
	service.mutex.Unlock()
	if !found || current != sequence {
		return
	}
	items, err := service.furniture.Grant(ctx, furnitureservice.GrantParams{DefinitionID: service.config.PrizeDefinitionID, OwnerPlayerID: request.PlayerID, Quantity: 1, ExtraData: "0"})
	if err != nil || len(items) == 0 {
		return
	}
	packet, err := outprize.Encode("s", int32(items[0].DefinitionID))
	if err == nil {
		_ = request.Target.Send(ctx, packet)
	}
}

// roomTaskKey derives a non-zero replaceable room task key.
func roomTaskKey(playerID int64) roomtask.Key { return roomtask.Key(uint64(playerID)<<1 | 1) }

// Package firework implements charged, exploding, and room-scheduled recharge states.
package firework

import (
	"context"
	"strconv"
	"time"

	fireworkcharged "github.com/niflaot/pixels/internal/realm/furniture/events/fireworkcharged"
	essential "github.com/niflaot/pixels/internal/realm/furniture/interactions/essential"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomtask "github.com/niflaot/pixels/internal/realm/room/runtime/live/task"
	netconn "github.com/niflaot/pixels/networking/connection"
	outstate "github.com/niflaot/pixels/networking/outbound/room/furniture/state"
	"github.com/niflaot/pixels/pkg/bus"
)

const (
	// ChargedState stores a firework ready to explode.
	ChargedState = "1"
	// ExplodingState stores the visible explosion animation.
	ExplodingState = "2"
)

// Service coordinates durable firework states and recharge work.
type Service struct {
	// states persists guarded firework states.
	states furnitureservice.StateUpdater
	// connections broadcasts state changes to current occupants.
	connections *netconn.Registry
	// events publishes successful explosions to progression.
	events bus.Publisher
	// defaultRecharge stores the normalized fallback recharge delay.
	defaultRecharge time.Duration
}

// New creates firework behavior.
func New(config Config, states furnitureservice.StateUpdater, connections *netconn.Registry, events bus.Publisher) *Service {
	config = config.Normalize()
	return &Service{states: states, connections: connections, events: events, defaultRecharge: config.DefaultRecharge}
}

// UseFurniture explodes a charged firework and schedules its recharge.
func (service *Service) UseFurniture(ctx context.Context, request essential.Request) (bool, error) {
	if request.Item.Definition.InteractionType != "firework" {
		return false, nil
	}
	if request.Item.ExtraData != ChargedState {
		return true, nil
	}
	if _, err := service.states.UpdateState(ctx, furnitureservice.StateParams{ItemID: request.Item.ID, RoomID: request.Room.ID(), Expected: ChargedState, Next: ExplodingState}); err != nil {
		return true, err
	}
	request.Room.SetFurnitureExtraData(request.Item.ID, ExplodingState)
	if err := service.broadcast(ctx, request, ExplodingState); err != nil {
		return true, err
	}
	if service.events != nil {
		_ = service.events.Publish(ctx, bus.Event{Name: fireworkcharged.Name, Payload: fireworkcharged.Payload{PlayerID: request.PlayerID, ItemID: request.Item.ID}})
	}
	recharge := service.recharge(request.Item.Definition.CustomParams)
	async := context.WithoutCancel(ctx)
	request.Room.ScheduleReplacing(roomtask.Key(uint64(request.Item.ID)<<1), recharge, func(time.Time) { _ = service.rechargeItem(async, request) })
	return true, nil
}

// rechargeItem persists and broadcasts one scheduled charged state.
func (service *Service) rechargeItem(ctx context.Context, request essential.Request) error {
	if _, err := service.states.UpdateState(ctx, furnitureservice.StateParams{ItemID: request.Item.ID, RoomID: request.Room.ID(), Expected: ExplodingState, Next: ChargedState}); err != nil {
		return err
	}
	request.Room.SetFurnitureExtraData(request.Item.ID, ChargedState)
	return service.broadcast(ctx, request, ChargedState)
}

// broadcast sends one firework state to current occupants.
func (service *Service) broadcast(ctx context.Context, request essential.Request, value string) error {
	state, _ := strconv.Atoi(value)
	packet, err := outstate.Encode(request.Item.ID, state)
	if err != nil {
		return err
	}
	return broadcast.RoomPacket(ctx, service.connections, request.Room, packet, 0)
}

// recharge parses a positive duration override in seconds.
func (service *Service) recharge(custom string) time.Duration {
	seconds, err := strconv.Atoi(custom)
	if err != nil || seconds <= 0 || seconds > 300 {
		return service.defaultRecharge
	}
	return time.Duration(seconds) * time.Second
}

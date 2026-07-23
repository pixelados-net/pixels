// Package games coordinates room-owned furniture game engines.
package games

import (
	"context"
	"strconv"
	"sync"
	"time"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/games/banzai"
	gamesconfig "github.com/niflaot/pixels/internal/realm/room/world/games/config"
	"github.com/niflaot/pixels/internal/realm/room/world/games/freeze"
	"github.com/niflaot/pixels/internal/realm/room/world/games/tag"
	gametimer "github.com/niflaot/pixels/internal/realm/room/world/games/timer"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	wiredgame "github.com/niflaot/pixels/internal/realm/room/world/wired/game"
	netconn "github.com/niflaot/pixels/networking/connection"
	outstate "github.com/niflaot/pixels/networking/outbound/room/furniture/state"
	outplaying "github.com/niflaot/pixels/networking/outbound/room/games/playing"
	"github.com/niflaot/pixels/pkg/bus"
)

// Service coordinates room games on the existing room owner loop.
type Service struct {
	// config stores enabled mechanics and scoring.
	config gamesconfig.Config
	// rooms resolves active runtime worlds.
	rooms *roomlive.Registry
	// wired stores shared team and score state.
	wired *wiredgame.Service
	// coordinator owns shared lifecycle triggers and highscores.
	coordinator *wiredgame.Coordinator
	// connections projects room and player packets.
	connections *netconn.Registry
	// furniture persists authoritative football movement.
	furniture furnitureMover
	// scores persists completed room matches.
	scores ScoreStore
	// events publishes committed progression deltas.
	events bus.Publisher
	// metrics stores lock-free room game telemetry.
	metrics *Metrics
	// mutex protects room game indexes against click/event concurrency.
	mutex sync.Mutex
	// states stores active room game state.
	states map[int64]*roomState
}

// furnitureMover persists one authoritative football placement.
type furnitureMover interface {
	// Move repositions one placed football.
	Move(context.Context, furnitureservice.MoveParams) (furnituremodel.Item, error)
}

// New creates the room-game coordinator.
func New(config gamesconfig.Config, rooms *roomlive.Registry, wired *wiredgame.Service, coordinator *wiredgame.Coordinator, connections *netconn.Registry, furniture *furnitureservice.Service, scores ScoreStore, events bus.Publisher, metrics *Metrics) *Service {
	return &Service{config: config, rooms: rooms, wired: wired, coordinator: coordinator, connections: connections, furniture: furniture, scores: scores, events: events, metrics: metrics, states: make(map[int64]*roomState)}
}

// Cycle advances active game timers from the room-owned scheduler.
func (service *Service) Cycle(ctx context.Context, active *roomlive.Room, now time.Time) error {
	if err := service.cycleFreeze(ctx, active, now); err != nil {
		return err
	}
	if err := service.cycleFootball(ctx, active); err != nil {
		return err
	}
	service.cycleTag(ctx, active, now)
	service.mutex.Lock()
	state := service.states[active.ID()]
	if state == nil || state.timer == nil || !state.timer.Running {
		service.mutex.Unlock()
		return nil
	}
	elapsed := now.Sub(state.lastTick)
	state.lastTick = now
	ended := state.timer.Tick(elapsed)
	remaining, timerID, startedAt := int(state.timer.Remaining/time.Second), state.timerItemID, state.startedAt
	service.mutex.Unlock()
	if err := service.projectState(ctx, active, timerID, remaining); err != nil {
		return err
	}
	if ended {
		return service.finish(ctx, active, startedAt)
	}
	return nil
}

// Close releases one room's game state.
func (service *Service) Close(roomID int64) {
	service.mutex.Lock()
	delete(service.states, roomID)
	service.mutex.Unlock()
	service.wired.Close(roomID)
}

// toggleTimer starts or pauses all game engines in one room.
func (service *Service) toggleTimer(ctx context.Context, request UseRequest) error {
	service.mutex.Lock()
	state := service.stateLocked(request.Room)
	state.timerItemID = request.Item.ID
	starting := !state.timer.Started
	state.timer.Toggle()
	state.lastTick = time.Now()
	remaining := int(state.timer.Remaining / time.Second)
	service.mutex.Unlock()
	if starting {
		service.mutex.Lock()
		state.finishing = false
		state.startedAt = state.lastTick
		service.initializeMatchLocked(request.Room, state, state.lastTick)
		service.mutex.Unlock()
		if err := service.coordinator.Start(ctx, request.Room.ID()); err != nil {
			return err
		}
		if err := service.resetFreezeRound(ctx, request.Room); err != nil {
			return err
		}
		service.metrics.started[gameKindIndex(roomGameKind(request.Room))].Add(1)
	}
	return service.projectState(ctx, request.Room, request.Item.ID, remaining)
}

// projectState updates ephemeral furniture state and broadcasts its native packet.
func (service *Service) projectState(ctx context.Context, active *roomlive.Room, itemID int64, value int) error {
	active.SetFurnitureExtraData(itemID, strconv.Itoa(value))
	packet, err := outstate.Encode(itemID, value)
	if err != nil {
		return err
	}
	return broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
}

// sendPlaying projects the local player's game participation flag.
func (service *Service) sendPlaying(ctx context.Context, active *roomlive.Room, playerID int64, playing bool) error {
	occupant, found := active.Occupant(playerID)
	if !found {
		return nil
	}
	connection, found := service.connections.Get(occupant.ConnectionKind, occupant.ConnectionID)
	if !found {
		return nil
	}
	packet, err := outplaying.Encode(playing)
	if err != nil {
		return err
	}
	return connection.Send(ctx, packet)
}

// roomState stores one room's active furniture-game state.
type roomState struct {
	// timer stores match duration.
	timer *gametimer.Timer
	// timerItemID identifies the controlling furniture.
	timerItemID int64
	// lastTick stores prior cycle time.
	lastTick time.Time
	// board stores Battle Banzai tile state.
	board *banzai.Board
	// tileItems maps board indexes to item ids.
	tileItems []int64
	// tags stores tag arenas by variant.
	tags map[tag.Variant]*tag.Game
	// freezePlayers stores active Freeze participants.
	freezePlayers map[int64]*freeze.Player
	// freezeBalls stores room-tick explosion deadlines.
	freezeBalls []freeze.Throw
	// freezeDrops stores broken blocks and their optional uncollected reward.
	freezeDrops map[int64]freeze.PowerUp
	// footballs stores rolling ball queues by furniture id.
	footballs map[int64]*footballBall
	// footballOrigins stores each ball's kickoff tile.
	footballOrigins map[int64]grid.Point
	// footballLooks stores original figures while a kit is equipped.
	footballLooks map[int64]string
	// nextTagCredit stores the next minute-based Tag progression boundary.
	nextTagCredit time.Time
	// startedAt stores the current match start.
	startedAt time.Time
	// finishing prevents duplicate end coordination across concurrent inputs.
	finishing bool
}

// stateLocked creates room engines from placed furniture.
func (service *Service) stateLocked(active *roomlive.Room) *roomState {
	state := service.states[active.ID()]
	if state != nil {
		return state
	}
	state = &roomState{timer: gametimer.New(timerSteps(active)), tags: map[tag.Variant]*tag.Game{tag.IceTag: tag.New(tag.IceTag), tag.Rollerskate: tag.New(tag.Rollerskate), tag.Bunnyrun: tag.New(tag.Bunnyrun)}, freezePlayers: make(map[int64]*freeze.Player), freezeDrops: make(map[int64]freeze.PowerUp), footballs: make(map[int64]*footballBall), footballOrigins: make(map[int64]grid.Point), footballLooks: make(map[int64]string)}
	for _, ball := range active.FurnitureByInteraction("football") {
		state.footballOrigins[ball.ID] = ball.Point
	}
	service.initializeBanzai(active, state)
	service.states[active.ID()] = state
	return state
}

// initializeBanzai maps arbitrarily placed rectangular tiles into a compact board.
func (service *Service) initializeBanzai(active *roomlive.Room, state *roomState) {
	tiles := active.FurnitureByInteraction("battlebanzai_tile")
	if len(tiles) == 0 {
		return
	}
	minX, maxX, minY, maxY := int(tiles[0].Point.X), int(tiles[0].Point.X), int(tiles[0].Point.Y), int(tiles[0].Point.Y)
	for _, item := range tiles {
		x, y := int(item.Point.X), int(item.Point.Y)
		if x < minX {
			minX = x
		}
		if x > maxX {
			maxX = x
		}
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}
	state.board = banzai.NewBoard(maxX-minX+1, maxY-minY+1)
	state.tileItems = make([]int64, len(state.board.Tiles))
	for _, item := range tiles {
		state.tileItems[(int(item.Point.Y)-minY)*state.board.Width+int(item.Point.X)-minX] = item.ID
	}
}

// roomGameKind derives one match kind from placed furniture.
func roomGameKind(active *roomlive.Room) string {
	interactions := []string{"battlebanzai_tile", "freeze_tile", "football", "icetag_field", "rollerskate_field", "bunnyrun_field"}
	kinds := []string{"banzai", "freeze", "football", "tag", "tag", "tag"}
	for index, interaction := range interactions {
		if len(active.FurnitureByInteraction(interaction)) > 0 {
			return kinds[index]
		}
	}
	return "wired"
}

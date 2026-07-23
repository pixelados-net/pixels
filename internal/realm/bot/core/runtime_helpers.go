package core

import (
	"context"
	"strings"
	"time"

	botrecord "github.com/niflaot/pixels/internal/realm/bot/record"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	sdkbot "github.com/niflaot/pixels/sdk/bot"
	"go.uber.org/zap"
)

// AddPlaced inserts one newly placed bot into a loaded room generation.
func (service *Service) AddPlaced(bot botrecord.Bot) {
	if bot.RoomID == nil {
		return
	}
	service.mutex.Lock()
	state := service.active[*bot.RoomID]
	if state == nil {
		state = &roomState{bots: make(map[int64]*activeBot)}
		service.active[*bot.RoomID] = state
	}
	state.bots[bot.ID] = service.newActiveBot(bot, time.Now())
	state.rebuildSnapshot()
	service.mutex.Unlock()
}

// RemovePlaced removes one bot from a loaded room generation.
func (service *Service) RemovePlaced(roomID int64, botID int64) {
	service.mutex.Lock()
	if state := service.active[roomID]; state != nil {
		delete(state.bots, botID)
		state.rebuildSnapshot()
	}
	service.mutex.Unlock()
}

// ReplacePlaced refreshes one loaded bot after a durable settings save.
func (service *Service) ReplacePlaced(saved botrecord.Bot) {
	if saved.RoomID == nil {
		return
	}
	if bot, found := service.activeByID(*saved.RoomID, saved.ID); found {
		bot.mutex.Lock()
		bot.record = saved
		bot.mutex.Unlock()
	}
}

// UnloadRoom releases one inactive room bot generation.
func (service *Service) UnloadRoom(roomID int64) {
	service.mutex.Lock()
	delete(service.active, roomID)
	service.mutex.Unlock()
}

// roomBots returns a stable pointer snapshot without per-tick database work.
func (service *Service) roomBots(roomID int64) []*activeBot {
	service.mutex.RLock()
	state := service.active[roomID]
	if state == nil {
		service.mutex.RUnlock()
		return nil
	}
	snapshot := state.snapshot.Load()
	service.mutex.RUnlock()
	if snapshot == nil {
		return nil
	}
	return snapshot.bots
}

// PlacedCount returns one loaded room bot count.
func (service *Service) PlacedCount(roomID int64) int {
	service.mutex.RLock()
	state := service.active[roomID]
	count := 0
	if state != nil {
		count = len(state.bots)
	}
	service.mutex.RUnlock()
	return count
}

// ResolveByName returns an immutable active bot snapshot by case-insensitive name.
func (service *Service) ResolveByName(roomID int64, name string) (sdkbot.Bot, bool) {
	for _, bot := range service.roomBots(roomID) {
		bot.mutex.Lock()
		matched := strings.EqualFold(bot.record.Name, name)
		view := service.StaticView(bot.record)
		view.FollowingPlayerID = bot.followingPlayerID
		bot.mutex.Unlock()
		if matched {
			return view, true
		}
	}
	return sdkbot.Bot{}, false
}

// ResolveByID returns an immutable active bot snapshot by durable identifier.
func (service *Service) ResolveByID(roomID int64, botID int64) (sdkbot.Bot, bool) {
	bot, found := service.activeByID(roomID, botID)
	if !found {
		return sdkbot.Bot{}, false
	}
	bot.mutex.Lock()
	view := service.StaticView(bot.record)
	view.FollowingPlayerID = bot.followingPlayerID
	bot.mutex.Unlock()
	return view, true
}

// ChangeFigure persists and projects one active bot figure.
func (service *Service) ChangeFigure(ctx context.Context, roomID int64, botID int64, figure string) error {
	figure = strings.TrimSpace(figure)
	if figure == "" || len(figure) > 512 {
		return botrecord.ErrInvalidSkill
	}
	bot, found := service.activeByID(roomID, botID)
	if !found {
		return botrecord.ErrBotNotFound
	}
	bot.mutex.Lock()
	record := bot.record
	bot.mutex.Unlock()
	record.Figure = figure
	saved, found, err := service.store.Save(ctx, record)
	if err != nil || !found {
		if err != nil {
			return err
		}
		return botrecord.ErrConflict
	}
	service.ReplacePlaced(saved)
	if active, activeFound := service.rooms.Find(roomID); activeFound {
		service.ProjectSpawn(ctx, active, saved)
	}
	return nil
}

// activeByID returns one loaded bot by room and durable id.
func (service *Service) activeByID(roomID int64, botID int64) (*activeBot, bool) {
	service.mutex.RLock()
	state := service.active[roomID]
	var bot *activeBot
	if state != nil {
		bot = state.bots[botID]
	}
	service.mutex.RUnlock()
	return bot, bot != nil
}

// capturePosition queues deferred persistence when a moving bot changed position.
func (service *Service) capturePosition(roomID int64, bot *activeBot, unit roomlive.UnitSnapshot, now time.Time) {
	x, y, z, rotation := int(unit.Position.Point.X), int(unit.Position.Point.Y), unit.Position.Z.Units(), int16(unit.BodyRotation)
	unchanged := bot.record.X != nil && *bot.record.X == x && bot.record.Y != nil && *bot.record.Y == y && bot.record.Z != nil && *bot.record.Z == z && bot.record.Rotation != nil && *bot.record.Rotation == rotation
	if unchanged || now.Sub(bot.lastPositionFlush) < service.config.PositionFlushInterval {
		return
	}
	service.queuePosition(roomID, bot, x, y, z, rotation, now)
}

// queuePosition updates mutable pointers and defers the durable write outside the hot path.
func (service *Service) queuePosition(roomID int64, bot *activeBot, x int, y int, z float64, rotation int16, now time.Time) {
	bot.record.X, bot.record.Y, bot.record.Z, bot.record.Rotation = &x, &y, &z, &rotation
	bot.lastPositionFlush = now
	botID := bot.record.ID
	service.dispatch(func() {
		if err := service.store.SavePosition(context.Background(), botID, roomID, x, y, z, rotation); err != nil && service.log != nil {
			service.log.Warn("save bot position", zap.Int64("bot_id", botID), zap.Error(err))
		}
	})
}

// follow updates one bot goal behind its live target.
func (service *Service) follow(active *roomlive.Room, bot *activeBot, unit roomlive.UnitSnapshot) {
	target, found := active.Unit(bot.followingPlayerID)
	if !found {
		bot.followingPlayerID = 0
		return
	}
	goal, valid := rotatedNeighbor(target.Position.Point, (int(target.BodyRotation)+4)%8)
	if !valid {
		goal, valid = rotatedNeighbor(target.Position.Point, int(target.BodyRotation))
	}
	if valid && squaredDistance(unit.Position.Point, target.Position.Point) >= 4 {
		_, _ = active.MoveTo(EntityKey(bot.id), goal)
	}
}

// StopFollowingEverywhere clears a followed player from every loaded room.
func (service *Service) StopFollowingEverywhere(playerID int64) {
	service.mutex.RLock()
	states := make([]*roomState, 0, len(service.active))
	for _, state := range service.active {
		states = append(states, state)
	}
	service.mutex.RUnlock()
	for _, state := range states {
		snapshot := state.snapshot.Load()
		if snapshot == nil {
			continue
		}
		for _, bot := range snapshot.bots {
			bot.mutex.Lock()
			if bot.followingPlayerID == playerID {
				bot.followingPlayerID = 0
			}
			bot.mutex.Unlock()
		}
	}
}

// RandomWalk implements the SDK bounded random movement action.
func (service *Service) RandomWalk(_ context.Context, view sdkbot.Bot) error {
	active, found := service.rooms.Find(view.RoomID)
	if !found {
		return botrecord.ErrRoomNotFound
	}
	bot, found := service.activeByID(view.RoomID, view.ID)
	if !found {
		return botrecord.ErrBotNotFound
	}
	bot.mutex.Lock()
	delay := 5 + int(service.source.Uint64()%34)
	bot.nextWalk = time.Now().Add(time.Duration(delay) * time.Second)
	bot.mutex.Unlock()
	radius := service.config.WalkRadius
	if !service.config.LimitWalkRadius {
		radius = 1 << 16
	}
	goal, found := active.RandomWalkablePoint(EntityKey(view.ID), radius, service.source.Uint64())
	if !found {
		return nil
	}
	_, err := active.MoveTo(EntityKey(view.ID), goal)
	return err
}

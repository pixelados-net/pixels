package core

import (
	"context"
	"strconv"
	"strings"
	"time"

	botrecord "github.com/niflaot/pixels/internal/realm/bot/record"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	sdkbot "github.com/niflaot/pixels/sdk/bot"
	"go.uber.org/zap"
)

// EnsureRoom loads and attaches placed bots exactly once for an active room generation.
func (service *Service) EnsureRoom(ctx context.Context, active *roomlive.Room) error {
	_, err := service.ensureRoom(ctx, active)
	return err
}

// ensureRoom loads one active room and reports whether this call published the first generation.
func (service *Service) ensureRoom(ctx context.Context, active *roomlive.Room) (bool, error) {
	if active == nil {
		return false, botrecord.ErrRoomNotFound
	}
	roomID := active.ID()
	service.mutex.RLock()
	_, loaded := service.active[roomID]
	service.mutex.RUnlock()
	if loaded {
		return false, nil
	}
	bots, err := service.store.Room(ctx, roomID)
	if err != nil {
		return false, err
	}
	name, ownerName := "", ""
	if service.roomRecords != nil {
		if room, found, findErr := service.roomRecords.FindByID(ctx, roomID); findErr == nil && found {
			name, ownerName = room.Name, room.OwnerName
		}
	}
	state := &roomState{bots: make(map[int64]*activeBot, len(bots)), name: name, ownerName: ownerName}
	now := time.Now()
	for _, bot := range bots {
		if bot.X == nil || bot.Y == nil || bot.Z == nil || bot.Rotation == nil {
			continue
		}
		point, valid := grid.NewPoint(*bot.X, *bot.Y)
		if !valid {
			continue
		}
		position := worldpath.Position{Point: point, Z: grid.HeightFromUnits(*bot.Z)}
		if _, addErr := active.AddEntity(EntityKey(bot.ID), bot.OwnerPlayerID, worldunit.KindBot, position, worldunit.Rotation(*bot.Rotation)); addErr != nil {
			continue
		}
		state.bots[bot.ID] = service.newActiveBot(bot, now)
	}
	state.rebuildSnapshot()
	service.mutex.Lock()
	if existing := service.active[roomID]; existing != nil {
		service.mutex.Unlock()
		return false, nil
	}
	service.active[roomID] = state
	service.mutex.Unlock()
	for _, bot := range bots {
		if state.bots[bot.ID] != nil {
			service.ProjectSpawn(ctx, active, bot)
		}
	}
	return true, nil
}

// SyncPlayer sends every active bot snapshot to one late room entrant.
func (service *Service) SyncPlayer(ctx context.Context, roomID int64, playerID int64) error {
	active, found := service.rooms.Find(roomID)
	if !found {
		return botrecord.ErrRoomNotFound
	}
	loaded, err := service.ensureRoom(ctx, active)
	if err != nil {
		return err
	}
	if loaded {
		return nil
	}
	connection, found := service.playerConnection(playerID)
	if !found {
		return nil
	}
	for _, bot := range service.roomBots(roomID) {
		bot.mutex.Lock()
		record := bot.record
		bot.mutex.Unlock()
		service.projectSpawnConnection(ctx, connection, active, record)
	}
	return nil
}

// Cycle advances every bot from the room's existing single owner tick.
func (service *Service) Cycle(ctx context.Context, active *roomlive.Room, now time.Time) error {
	if err := service.EnsureRoom(ctx, active); err != nil {
		return err
	}
	for _, bot := range service.roomBots(active.ID()) {
		service.cycleBot(ctx, active, bot, now)
	}
	return nil
}

// cycleBot builds one behavior snapshot and delegates without persistence I/O.
func (service *Service) cycleBot(ctx context.Context, active *roomlive.Room, bot *activeBot, now time.Time) {
	unit, found := active.UnitMotion(EntityKey(bot.id))
	if !found {
		bot.mutex.Lock()
		record := bot.record
		bot.mutex.Unlock()
		unit, found = service.reattach(ctx, active, record)
		if !found {
			return
		}
	}
	bot.mutex.Lock()
	service.capturePosition(active.ID(), bot, unit, now)
	if bot.followingPlayerID != 0 {
		service.follow(active, bot, unit)
	}
	view := service.behaviorView(active, bot, now)
	bot.mutex.Unlock()
	if view.ChatDue {
		view.ChatMessage = service.expand(view.ChatMessage, active.ID(), view.Name)
	}
	behavior := bot.behavior
	if behavior != nil {
		if err := behavior.OnCycle(ctx, view, service); err != nil && service.log != nil {
			service.log.Debug("bot cycle rejected", zap.Int64("bot_id", view.ID), zap.Error(err))
		}
	}
}

// reattach restores a bot after an in-place world reload.
func (service *Service) reattach(ctx context.Context, active *roomlive.Room, record botrecord.Bot) (roomlive.UnitSnapshot, bool) {
	if record.X == nil || record.Y == nil || record.Z == nil || record.Rotation == nil {
		return roomlive.UnitSnapshot{}, false
	}
	point, valid := grid.NewPoint(*record.X, *record.Y)
	if !valid {
		return roomlive.UnitSnapshot{}, false
	}
	position := worldpath.Position{Point: point, Z: grid.HeightFromUnits(*record.Z)}
	unit, err := active.AddEntity(EntityKey(record.ID), record.OwnerPlayerID, worldunit.KindBot, position, worldunit.Rotation(*record.Rotation))
	if err != nil {
		return roomlive.UnitSnapshot{}, false
	}
	service.ProjectSpawn(ctx, active, record)
	return unit, true
}

// behaviorView creates an immutable behavior snapshot and advances due chat state.
func (service *Service) behaviorView(active *roomlive.Room, bot *activeBot, now time.Time) sdkbot.Bot {
	record := bot.record
	view := sdkbot.Bot{ID: record.ID, OwnerPlayerID: record.OwnerPlayerID, RoomID: active.ID(), Name: record.Name, BehaviorType: record.BehaviorType, CanWalk: record.CanWalk, FollowingPlayerID: bot.followingPlayerID, WalkDue: !now.Before(bot.nextWalk)}
	if record.ChatAuto && len(record.ChatLines) > 0 && !now.Before(bot.nextChat) {
		if bot.chatIndex < 0 || bot.chatIndex >= len(record.ChatLines) {
			bot.chatIndex = 0
		}
		view.ChatDue = true
		view.ChatMessage = record.ChatLines[bot.chatIndex]
		if record.ChatRandom {
			bot.chatIndex = int(service.source.Uint64() % uint64(len(record.ChatLines)))
		} else {
			bot.chatIndex = (bot.chatIndex + 1) % len(record.ChatLines)
		}
		bot.nextChat = now.Add(time.Duration(record.ChatDelaySeconds) * time.Second)
	}
	return view
}

// expand substitutes every supported chat template variable.
func (service *Service) expand(message string, roomID int64, botName string) string {
	service.mutex.RLock()
	state := service.active[roomID]
	service.mutex.RUnlock()
	if state == nil {
		return message
	}
	active, _ := service.rooms.Find(roomID)
	itemCount, userCount := 0, 0
	if active != nil {
		itemCount = len(active.FurnitureItems())
		userCount = active.Occupancy().Count
	}
	replacer := strings.NewReplacer("%owner%", state.ownerName, "%item_count%", strconv.Itoa(itemCount), "%name%", botName, "%roomname%", state.name, "%user_count%", strconv.Itoa(userCount))
	return replacer.Replace(message)
}

// newActiveBot creates deterministic initial schedules.
func (service *Service) newActiveBot(bot botrecord.Bot, now time.Time) *activeBot {
	return &activeBot{id: bot.ID, behavior: service.behaviors.For(bot.BehaviorType), record: bot, nextChat: now.Add(time.Duration(bot.ChatDelaySeconds) * time.Second), nextWalk: now.Add(5 * time.Second), lastPositionFlush: now}
}

// squaredDistance returns allocation-free tile distance squared.
func squaredDistance(first grid.Point, second grid.Point) int {
	dx, dy := int(first.X)-int(second.X), int(first.Y)-int(second.Y)
	return dx*dx + dy*dy
}

// rotatedNeighbor returns one adjacent tile in an eight-direction rotation.
func rotatedNeighbor(point grid.Point, rotation int) (grid.Point, bool) {
	offsets := [8][2]int{{0, -1}, {1, -1}, {1, 0}, {1, 1}, {0, 1}, {-1, 1}, {-1, 0}, {-1, -1}}
	offset := offsets[rotation&7]
	x, y := int(point.X)+offset[0], int(point.Y)+offset[1]
	return grid.NewPoint(x, y)
}

package games

import (
	"context"
	"sync/atomic"
	"time"

	furniturewalkedoff "github.com/niflaot/pixels/internal/realm/furniture/events/walkedoff"
	roomleft "github.com/niflaot/pixels/internal/realm/room/access/events/left"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/games/tag"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/pkg/bus"
)

// Metrics stores lock-free room game telemetry.
type Metrics struct {
	// started stores starts by stable game kind index.
	started [5]atomic.Uint64
	// ended stores completions by stable game kind index.
	ended [5]atomic.Uint64
	// durationNanoseconds stores aggregate completed duration.
	durationNanoseconds atomic.Uint64
	// tilesLocked stores Banzai locks including captured fills.
	tilesLocked atomic.Uint64
	// freezeBalls stores scheduled Freeze balls.
	freezeBalls atomic.Uint64
	// footballGoals stores directional football goals.
	footballGoals atomic.Uint64
}

// MetricsSnapshot stores one stable telemetry read.
type MetricsSnapshot struct {
	// Started stores match starts by kind.
	Started map[string]uint64 `json:"started"`
	// Ended stores match completions by kind.
	Ended map[string]uint64 `json:"ended"`
	// AverageDurationMilliseconds stores aggregate match duration.
	AverageDurationMilliseconds uint64 `json:"averageDurationMilliseconds"`
	// TilesLocked stores Banzai locks.
	TilesLocked uint64 `json:"tilesLocked"`
	// FreezeBalls stores scheduled balls.
	FreezeBalls uint64 `json:"freezeBalls"`
	// FootballGoals stores scored goals.
	FootballGoals uint64 `json:"footballGoals"`
}

// NewMetrics creates zeroed room game telemetry.
func NewMetrics() *Metrics { return &Metrics{} }

// Snapshot returns one lock-free telemetry snapshot.
func (metrics *Metrics) Snapshot() MetricsSnapshot {
	kinds := [...]string{"banzai", "freeze", "football", "tag", "wired"}
	snapshot := MetricsSnapshot{Started: make(map[string]uint64, len(kinds)), Ended: make(map[string]uint64, len(kinds)), TilesLocked: metrics.tilesLocked.Load(), FreezeBalls: metrics.freezeBalls.Load(), FootballGoals: metrics.footballGoals.Load()}
	var ended uint64
	for index, kind := range kinds {
		snapshot.Started[kind], snapshot.Ended[kind] = metrics.started[index].Load(), metrics.ended[index].Load()
		ended += snapshot.Ended[kind]
	}
	if ended > 0 {
		snapshot.AverageDurationMilliseconds = metrics.durationNanoseconds.Load() / ended / uint64(time.Millisecond)
	}
	return snapshot
}

// gameKindIndex maps stable storage names to counters.
func gameKindIndex(kind string) int {
	switch kind {
	case "banzai":
		return 0
	case "freeze":
		return 1
	case "football":
		return 2
	case "tag":
		return 3
	default:
		return 4
	}
}

// WalkedOff removes players leaving Tag fields.
func (service *Service) WalkedOff(ctx context.Context, event bus.Event) error {
	payload, ok := event.Payload.(furniturewalkedoff.Payload)
	if !ok {
		return nil
	}
	active, found := service.rooms.Find(payload.RoomID)
	if !found {
		return nil
	}
	item, found := active.FurnitureItem(payload.ItemID)
	if !found || !isTagField(item.Definition.InteractionType) {
		return nil
	}
	variant := tagVariant(item.Definition.InteractionType)
	service.mutex.Lock()
	state := service.states[payload.RoomID]
	left := state != nil && state.tags[variant].Leave(payload.PlayerID)
	service.mutex.Unlock()
	if left {
		service.projectTag(active, variant)
		service.wired.ProjectEffect(payload.RoomID, payload.PlayerID, 0)
		return service.sendPlaying(ctx, active, payload.PlayerID, false)
	}
	return nil
}

// PlayerLeft removes room game participation and effects.
func (service *Service) PlayerLeft(ctx context.Context, event bus.Event) error {
	payload, ok := event.Payload.(roomleft.Payload)
	if !ok {
		return nil
	}
	service.mutex.Lock()
	state := service.states[payload.RoomID]
	if state != nil {
		for _, game := range state.tags {
			game.Leave(payload.PlayerID)
		}
		delete(state.freezePlayers, payload.PlayerID)
		delete(state.footballLooks, payload.PlayerID)
	}
	service.mutex.Unlock()
	service.wired.RemovePlayer(payload.RoomID, payload.PlayerID)
	service.wired.ProjectEffect(payload.RoomID, payload.PlayerID, 0)
	if active, found := service.rooms.Find(payload.RoomID); found {
		return service.sendPlaying(ctx, active, payload.PlayerID, false)
	}
	return nil
}

// transferTag transfers the active tag across an adjacent movement.
func (service *Service) transferTag(active *roomlive.Room, moverID int64) {
	mover, found := active.Unit(moverID)
	if !found {
		return
	}
	service.mutex.Lock()
	state := service.states[active.ID()]
	if state == nil {
		service.mutex.Unlock()
		return
	}
	changed := make([]tag.Variant, 0, 1)
	for variant, game := range state.tags {
		tagger := game.Tagger()
		for _, playerID := range game.Players() {
			if playerID == moverID {
				continue
			}
			other, present := active.Unit(playerID)
			adjacent := present && pointsAdjacent(mover.Position.Point, other.Position.Point)
			if tagger == moverID && game.Transfer(moverID, playerID, adjacent) || tagger == playerID && game.Transfer(playerID, moverID, adjacent) {
				changed = append(changed, variant)
				break
			}
		}
	}
	service.mutex.Unlock()
	for _, variant := range changed {
		service.projectTag(active, variant)
		service.flashTagPoles(active, variant)
	}
}

// flashTagPoles projects a short server-owned transfer animation.
func (service *Service) flashTagPoles(active *roomlive.Room, variant tag.Variant) {
	kind := "icetag_pole"
	if variant == tag.Bunnyrun {
		kind = "bunnyrun_pole"
	}
	if variant == tag.Rollerskate {
		kind = "rollerskate_pole"
	}
	for _, pole := range active.FurnitureByInteraction(kind) {
		_ = service.projectState(context.Background(), active, pole.ID, 1)
		id := pole.ID
		active.Schedule(500*time.Millisecond, func(time.Time) { _ = service.projectState(context.Background(), active, id, 0) })
	}
}

// projectTag refreshes participant effects for one variant.
func (service *Service) projectTag(active *roomlive.Room, variant tag.Variant) {
	service.mutex.Lock()
	state := service.states[active.ID()]
	if state == nil {
		service.mutex.Unlock()
		return
	}
	game := state.tags[variant]
	tagger := game.Tagger()
	players := game.Players()
	service.mutex.Unlock()
	for _, playerID := range players {
		occupant, found := active.Occupant(playerID)
		female := found && (occupant.Gender == "F" || occupant.Gender == "f")
		service.wired.ProjectEffect(active.ID(), playerID, tag.Effect(variant, female, playerID == tagger))
	}
}

// cycleTag credits each completed active minute without a separate timer.
func (service *Service) cycleTag(ctx context.Context, active *roomlive.Room, now time.Time) {
	service.mutex.Lock()
	state := service.states[active.ID()]
	if state == nil || state.nextTagCredit.IsZero() || now.Before(state.nextTagCredit) {
		service.mutex.Unlock()
		return
	}
	state.nextTagCredit = state.nextTagCredit.Add(time.Minute)
	ice := state.tags[tag.IceTag].Players()
	rollers := state.tags[tag.Rollerskate].Players()
	service.mutex.Unlock()
	for _, playerID := range ice {
		service.progress(ctx, playerID, "game.tag.minutes", 1)
	}
	for _, playerID := range rollers {
		service.progress(ctx, playerID, "game.rollerskate.minutes", 1)
	}
}

// isTagField reports one Tag arena floor behavior.
func isTagField(kind string) bool {
	return kind == "icetag_field" || kind == "rollerskate_field" || kind == "bunnyrun_field"
}

// pointsAdjacent reports Chebyshev adjacency.
func pointsAdjacent(left grid.Point, right grid.Point) bool {
	dx, dy := int(left.X)-int(right.X), int(left.Y)-int(right.Y)
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx <= 1 && dy <= 1
}

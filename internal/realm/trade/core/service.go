package core

import (
	"context"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	tradecancelled "github.com/niflaot/pixels/internal/realm/trade/events/cancelled"
	tradecompleted "github.com/niflaot/pixels/internal/realm/trade/events/completed"
	tradestarted "github.com/niflaot/pixels/internal/realm/trade/events/started"
	traderecord "github.com/niflaot/pixels/internal/realm/trade/record"
	traderuntime "github.com/niflaot/pixels/internal/realm/trade/runtime"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/redis"
	"net"
	"strconv"
	"time"
)

// Service implements direct-trade lifecycle and settlement.
type Service struct {
	// config stores immutable direct-trade policy.
	config Options
	// registry owns live sessions and staged items.
	registry *traderuntime.Registry
	// players stores live player eligibility and ignore state.
	players *playerlive.Registry
	// rooms stores active room units and trade policy.
	rooms *roomlive.Registry
	// connections projects room status changes to active occupants.
	connections *netconn.Registry
	// furniture reads and mutates item ownership.
	furniture furnitureservice.TradingManager
	// currencies grants redeemed credit furniture value.
	currencies currencyservice.Granter
	// store owns settlement transactions and audit persistence.
	store traderecord.Store
	// permissions resolves restriction bypasses.
	permissions permissionservice.Checker
	// throttle stores distributed trade-start cooldowns.
	throttle *redis.Client
	// events publishes committed direct-trade facts.
	events bus.Publisher
}

// New creates a direct-trade service.
func New(config Options, registry *traderuntime.Registry, players *playerlive.Registry, rooms *roomlive.Registry, connections *netconn.Registry, furniture furnitureservice.TradingManager, currencies currencyservice.Granter, store traderecord.Store, permissions permissionservice.Checker, throttle *redis.Client, events bus.Publisher) *Service {
	return &Service{config: config, registry: registry, players: players, rooms: rooms, connections: connections, furniture: furniture, currencies: currencies, store: store, permissions: permissions, throttle: throttle, events: events}
}

// Registry returns the live trade registry.
func (service *Service) Registry() *traderuntime.Registry { return service.registry }

// Start opens a trade between an actor and target room unit.
func (service *Service) Start(ctx context.Context, actorID int64, targetUnitID int64, actorIP string) (*traderuntime.Session, error) {
	if service.throttle != nil {
		allowed, err := service.throttle.SetIfAbsent(ctx, "trade:start:"+strconv.FormatInt(actorID, 10), []byte{1}, service.config.StartThrottle)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return nil, ErrThrottled
		}
	}
	actor, found := service.players.Find(actorID)
	if !found {
		return nil, ErrUnavailable
	}
	room, found := service.rooms.FindByPlayer(actorID)
	if !found {
		return nil, ErrUnavailable
	}
	targetUnit, found := room.UnitByID(targetUnitID)
	if !found || targetUnit.PlayerID == actorID {
		return nil, ErrUnavailable
	}
	target, found := service.players.Find(targetUnit.PlayerID)
	if !found {
		return nil, ErrUnavailable
	}
	bypass := false
	if service.permissions != nil {
		bypass, _ = service.permissions.HasPermission(ctx, actorID, BypassRestrictions)
	}
	if !service.config.Enabled && !bypass {
		return nil, ErrDisabled
	}
	snapshot := room.Snapshot()
	if !bypass && (snapshot.TradeMode == 0 || snapshot.TradeMode == 1 && snapshot.OwnerPlayerID != actorID) {
		return nil, ErrRoomPolicy
	}
	if target.IsIgnoring(actorID) {
		return nil, ErrIgnored
	}
	if _, busy := service.registry.Find(actorID); busy {
		return nil, ErrUnavailable
	}
	actorSnapshot := actor.Snapshot()
	if !actorSnapshot.AllowTrade || actorSnapshot.Sanctions.TradeLockedAt(time.Now()) {
		return nil, ErrActorNotAllowed
	}
	if _, busy := service.registry.Find(targetUnit.PlayerID); busy {
		return nil, ErrUnavailable
	}
	targetSnapshot := target.Snapshot()
	if !targetSnapshot.AllowTrade || targetSnapshot.Sanctions.TradeLockedAt(time.Now()) {
		return nil, ErrTargetNotAllowed
	}
	actorUnit, found := room.Unit(actorID)
	if !found {
		return nil, ErrUnavailable
	}
	session := &traderuntime.Session{RoomID: snapshot.ID, First: traderuntime.Participant{PlayerID: actorID, UnitID: actorUnit.UnitID, Username: actor.Username(), IP: normalizeIP(actorIP)}, Second: traderuntime.Participant{PlayerID: targetUnit.PlayerID, UnitID: targetUnit.UnitID, Username: target.Username()}}
	if !service.registry.Start(session) {
		return nil, ErrUnavailable
	}
	room.SetUnitStatus(actorID, worldunit.StatusTrade, "")
	room.SetUnitStatus(targetUnit.PlayerID, worldunit.StatusTrade, "")
	service.projectStatuses(ctx, room, actorID, targetUnit.PlayerID)
	if service.events != nil {
		_ = service.events.Publish(ctx, bus.Event{Name: tradestarted.Name, Payload: tradestarted.Payload{RoomID: session.RoomID, FirstPlayerID: session.First.PlayerID, SecondPlayerID: session.Second.PlayerID}})
	}
	return session, nil
}

// normalizeIP removes a transport port and rejects non-IP audit values.
func normalizeIP(address string) string {
	host, _, err := net.SplitHostPort(address)
	if err == nil {
		address = host
	}
	if net.ParseIP(address) == nil {
		return ""
	}
	return address
}

// NormalizeIP exposes canonical audit address normalization to packet adapters.
func NormalizeIP(address string) string { return normalizeIP(address) }

// Confirm confirms one side and settles when both are ready.
func (service *Service) Confirm(ctx context.Context, playerID int64) (bool, error) {
	session, found := service.registry.Find(playerID)
	if !found {
		return false, ErrUnavailable
	}
	both, updated := session.Confirm(playerID)
	if !updated {
		return false, ErrNotAccepted
	}
	if !both {
		return false, nil
	}
	if err := service.settle(ctx, session); err != nil {
		session.FailSettlement()
		service.closeWithReason(playerID, 1)
		return false, err
	}
	service.closeSession(session)
	if service.events != nil {
		first, second := session.Snapshot()
		_ = service.events.Publish(ctx, bus.Event{Name: tradecompleted.Name, Payload: tradecompleted.Payload{RoomID: session.RoomID, FirstPlayerID: first.PlayerID, SecondPlayerID: second.PlayerID, FirstItemIDs: first.Items, SecondItemIDs: second.Items}})
	}
	return true, nil
}

// Close cancels one player's active trade.
func (service *Service) Close(playerID int64) bool {
	return service.closeWithReason(playerID, 0)
}

// closeWithReason cancels one active trade with an event reason.
func (service *Service) closeWithReason(playerID int64, reason int32) bool {
	session, found := service.registry.Close(playerID)
	if !found {
		return false
	}
	service.clearStatuses(session)
	if service.events != nil {
		_ = service.events.Publish(context.Background(), bus.Event{Name: tradecancelled.Name, Payload: tradecancelled.Payload{RoomID: session.RoomID, FirstPlayerID: session.First.PlayerID, SecondPlayerID: session.Second.PlayerID, Reason: reason}})
	}
	return true
}

// closeSession removes a completed session.
func (service *Service) closeSession(session *traderuntime.Session) {
	service.registry.Complete(session.First.PlayerID)
	service.clearStatuses(session)
}

// clearStatuses clears both participant room statuses.
func (service *Service) clearStatuses(session *traderuntime.Session) {
	if room, found := service.rooms.Find(session.RoomID); found {
		room.ClearUnitStatus(session.First.PlayerID, worldunit.StatusTrade)
		room.ClearUnitStatus(session.Second.PlayerID, worldunit.StatusTrade)
		service.projectStatuses(context.Background(), room, session.First.PlayerID, session.Second.PlayerID)
	}
}

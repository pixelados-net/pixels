package roller

import (
	"context"
	"testing"

	rolledevent "github.com/niflaot/pixels/internal/realm/furniture/events/rolled"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outrolling "github.com/niflaot/pixels/networking/outbound/room/furniture/rolling"
	outheightmapupdate "github.com/niflaot/pixels/networking/outbound/room/heightmapupdate"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/zap"
)

// persistenceManager captures one durable roller move.
type persistenceManager struct {
	furnitureservice.Manager
	// moves receives persisted placement requests.
	moves chan furnitureservice.MoveParams
}

// Move captures one durable placement.
func (manager *persistenceManager) Move(_ context.Context, params furnitureservice.MoveParams) (furnituremodel.Item, error) {
	manager.moves <- params
	return furnituremodel.Item{}, nil
}

// TestServiceLifecycleFlushesPersistence verifies bounded work drains at shutdown.
func TestServiceLifecycleFlushesPersistence(t *testing.T) {
	manager := &persistenceManager{moves: make(chan furnitureservice.MoveParams, 1)}
	service := New(Config{NoRules: true}, manager, nil, nil, zap.NewNop())
	if !service.NoRules() {
		t.Fatal("expected no-rules policy")
	}
	service.Start()
	service.enqueuePersistence(9, worldfurniture.Item{ID: 10, OwnerPlayerID: 7, Point: grid.MustPoint(2, 3), Z: 2, Rotation: worldunit.RotationEast})
	service.Stop()
	select {
	case moved := <-manager.moves:
		if moved.RoomID != 9 || moved.ItemID != 10 || moved.Placement.X != 2 || moved.Placement.Y != 3 || moved.Placement.Z != 0.5 {
			t.Fatalf("unexpected move %#v", moved)
		}
	default:
		t.Fatal("persistence was not flushed")
	}
}

// TestBroadcastPublishesConsolidatedEvent verifies applied source data maps safely.
func TestBroadcastPublishesConsolidatedEvent(t *testing.T) {
	local := bus.New()
	var payload rolledevent.Payload
	_, err := local.Subscribe(rolledevent.Name, bus.PriorityNormal, func(_ context.Context, event bus.Event) error {
		payload = event.Payload.(rolledevent.Payload)
		return nil
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	service := testService(local)
	active, err := roomlive.NewRoom(roomlive.Snapshot{ID: 9, MaxUsers: 1})
	if err != nil {
		t.Fatalf("room: %v", err)
	}
	sourceItem := stackedItem(10, 1, 2)
	targetItem := sourceItem
	targetItem.Point = grid.MustPoint(2, 0)
	sourceUnit := roomlive.UnitSnapshot{PlayerID: 7, UnitID: 3}
	sourceUnit.Position.Z = 2
	targetUnit := sourceUnit
	targetUnit.Position.Z = 4
	moved := movedStep{
		step:        step{roller: rollerItem(1, 1), target: grid.MustPoint(2, 0)},
		unitSources: []roomlive.UnitSnapshot{sourceUnit}, units: []roomlive.UnitSnapshot{targetUnit},
		itemSources: []worldfurniture.Item{sourceItem}, items: []worldfurniture.Item{targetItem},
	}
	if err = service.broadcast(context.Background(), active, moved); err != nil {
		t.Fatalf("broadcast: %v", err)
	}
	if payload.RoomID != 9 || payload.RollerItemID != 1 || len(payload.ItemIDs) != 1 || len(payload.PlayerIDs) != 1 {
		t.Fatalf("unexpected payload %#v", payload)
	}
}

// TestBroadcastDoesNotInterruptFurnitureRoll verifies Nitro receives only its authoritative roll and height update.
func TestBroadcastDoesNotInterruptFurnitureRoll(t *testing.T) {
	connections := netconn.NewRegistry()
	sent := registerRollerConnection(t, connections)
	service := testService(nil)
	service.connections = connections
	active := rollerRoom(t, []worldfurniture.Item{rollerItem(1, 1), stackedItem(10, 1, 2)})
	if _, err := active.Join(roomlive.Occupant{PlayerID: 7, ConnectionID: "roller", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join: %v", err)
	}
	source := stackedItem(10, 1, 2)
	target := source
	target.Point = grid.MustPoint(2, 0)
	target.Z = 0
	moved := movedStep{
		step:        step{roller: rollerItem(1, 1), target: grid.MustPoint(2, 0)},
		itemSources: []worldfurniture.Item{source}, items: []worldfurniture.Item{target},
	}
	if err := service.broadcast(context.Background(), active, moved); err != nil {
		t.Fatalf("broadcast: %v", err)
	}
	if len(*sent) != 2 || (*sent)[0].Header != outrolling.Header || (*sent)[1].Header != outheightmapupdate.Header {
		t.Fatalf("unexpected packets %#v", *sent)
	}
}

// registerRollerConnection registers one packet-capturing room connection.
func registerRollerConnection(t *testing.T, connections *netconn.Registry) *[]codec.Packet {
	t.Helper()
	sent := make([]codec.Packet, 0, 2)
	outbound := netconn.NewHandlerRegistry()
	outbound.SetFallback(func(netconn.Context, codec.Packet) error { return nil }, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	session, err := netconn.NewSession(netconn.SessionConfig{
		ID: "roller", Kind: "websocket", Outbound: outbound,
		Sender:   func(_ context.Context, packet codec.Packet) error { sent = append(sent, packet); return nil },
		Disposer: func(context.Context, netconn.Reason) error { return nil },
	})
	if err != nil {
		t.Fatalf("session: %v", err)
	}
	if err = connections.Register(session); err != nil {
		t.Fatalf("register: %v", err)
	}
	return &sent
}

// TestRollerRulesRecognizeChainsAndOpenGates verifies special destination policy.
func TestRollerRulesRecognizeChainsAndOpenGates(t *testing.T) {
	service := &Service{config: Config{}}
	source, target := rollerItem(1, 1), rollerItem(2, 2)
	if !service.validChain(source, target, []worldfurniture.Item{target}) {
		t.Fatal("expected equal-height top roller chain")
	}
	target.Z = 2
	if service.validChain(source, target, []worldfurniture.Item{target}) {
		t.Fatal("accepted unequal-height roller chain")
	}
	gate := worldfurniture.Item{ExtraData: "1", Definition: worldfurniture.Definition{InteractionType: "gate"}}
	if !allowsUsers([]worldfurniture.Item{gate}) {
		t.Fatal("open gate rejected rolled avatar")
	}
	gate.ExtraData = "0"
	if allowsUsers([]worldfurniture.Item{gate}) {
		t.Fatal("closed gate accepted rolled avatar")
	}
	service.config.NoRules = true
	if !service.validChain(source, target, nil) {
		t.Fatal("no-rules policy rejected chain")
	}
}

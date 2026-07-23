package pickup

import (
	"context"
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/command"
	pickedupevent "github.com/niflaot/pixels/internal/realm/furniture/events/pickedup"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	netconn "github.com/niflaot/pixels/networking/connection"
	outrefresh "github.com/niflaot/pixels/networking/outbound/inventory/furniture/refresh"
	outunseen "github.com/niflaot/pixels/networking/outbound/inventory/unseen"
	outstatus "github.com/niflaot/pixels/networking/outbound/room/entities/status"
	outremove "github.com/niflaot/pixels/networking/outbound/room/furniture/remove"
	outheightmapupdate "github.com/niflaot/pixels/networking/outbound/room/heightmapupdate"
	outbubble "github.com/niflaot/pixels/networking/outbound/session/bubblealert"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
	"go.uber.org/zap"
)

// TestHandlePicksUpItemAndBroadcasts verifies the successful pickup path.
func TestHandlePicksUpItemAndBroadcasts(t *testing.T) {
	handler, connections := handlerForTest(t)
	connection, sent := registeredConnectionForTest(t, connections, "conn")
	handler.Furniture = &fakeManager{pickupResult: pickedItemForTest()}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connection, ItemID: 1},
	})
	if err != nil {
		t.Fatalf("handle pickup: %v", err)
	}

	room, _ := handler.Runtime.Find(9)
	if items := room.FurnitureItems(); len(items) != 0 {
		t.Fatalf("expected no furniture items after pickup, got %#v", items)
	}
	if len(*sent) != 3 || (*sent)[0].Header != outremove.Header ||
		(*sent)[1].Header != outunseen.Header || (*sent)[2].Header != outrefresh.Header {
		t.Fatalf("expected floor remove then inventory unseen and refresh packets, got %#v", *sent)
	}
}

// TestHandleBroadcastsRemoveToOtherOccupants verifies both connections in a room observe a pickup.
func TestHandleBroadcastsRemoveToOtherOccupants(t *testing.T) {
	handler, connections := handlerForTest(t)
	actor, actorSent := registeredConnectionForTest(t, connections, "conn")
	_, bystanderSent := registeredConnectionForTest(t, connections, "bystander-conn")
	room, _ := handler.Runtime.Find(9)
	if _, err := room.Join(roomlive.Occupant{
		PlayerID: 8, Username: "bystander", ConnectionID: netconn.ID("bystander-conn"), ConnectionKind: netconn.Kind("websocket"),
	}); err != nil {
		t.Fatalf("join bystander: %v", err)
	}
	handler.Furniture = &fakeManager{pickupResult: pickedItemForTest()}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: actor, ItemID: 1},
	})
	if err != nil {
		t.Fatalf("handle pickup: %v", err)
	}

	if len(*actorSent) != 3 || (*actorSent)[0].Header != outremove.Header ||
		(*actorSent)[1].Header != outunseen.Header || (*actorSent)[2].Header != outrefresh.Header {
		t.Fatalf("expected actor to receive floor remove then inventory unseen and refresh, got %#v", *actorSent)
	}
	if len(*bystanderSent) != 1 || (*bystanderSent)[0].Header != outremove.Header {
		t.Fatalf("expected bystander to receive only floor remove, got %#v", *bystanderSent)
	}
}

// TestHandlePickingUpOccupiedChairStandsOccupantUp verifies picking up a chair stands its occupant
// up and syncs the resulting status to every room connection.
func TestHandlePickingUpOccupiedChairStandsOccupantUp(t *testing.T) {
	handler, connections := handlerForTest(t)
	actor, actorSent := registeredConnectionForTest(t, connections, "conn")
	room, _ := handler.Runtime.Find(9)
	settleUnitOnChair(t, room, 7, 1)
	handler.Furniture = &fakeManager{pickupResult: pickedItemForTest()}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: actor, ItemID: 1},
	})
	if err != nil {
		t.Fatalf("handle pickup: %v", err)
	}

	units := room.Units()
	if len(units) != 1 || unitHasStatus(units[0].Statuses, worldunit.StatusSit) {
		t.Fatalf("expected occupant to stand up after the chair was picked up, got %#v", units)
	}
	if len(*actorSent) != 4 ||
		(*actorSent)[0].Header != outremove.Header ||
		(*actorSent)[1].Header != outstatus.Header ||
		(*actorSent)[2].Header != outunseen.Header ||
		(*actorSent)[3].Header != outrefresh.Header {
		t.Fatalf("expected floor remove, unit status, inventory unseen, then refresh packets, got %#v", *actorSent)
	}
}

// TestHandlePicksUpItemBroadcastsHeightMapUpdate verifies the picked up item's vacated footprint
// is broadcast as a ROOM_HEIGHT_MAP_UPDATE so every client's cached local height map stays in sync.
func TestHandlePicksUpItemBroadcastsHeightMapUpdate(t *testing.T) {
	handler, connections := handlerForTest(t)
	connection, sent := registeredConnectionForTest(t, connections, "conn")
	handler.Furniture = &fakeManager{
		definition:      furnituremodel.Definition{Width: 1, Length: 1},
		definitionFound: true,
		pickupResult:    placedPickedItemForTest(),
	}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connection, ItemID: 1},
	})
	if err != nil {
		t.Fatalf("handle pickup: %v", err)
	}

	if len(*sent) != 4 ||
		(*sent)[0].Header != outremove.Header ||
		(*sent)[1].Header != outheightmapupdate.Header ||
		(*sent)[2].Header != outunseen.Header ||
		(*sent)[3].Header != outrefresh.Header {
		t.Fatalf("expected floor remove, height map update, inventory unseen, then refresh packets, got %#v", *sent)
	}
}

// placedPickedItemForTest returns a picked up item fixture that was on the floor before pickup.
func placedPickedItemForTest() furnituremodel.Item {
	roomID, x, y, z := int64(9), 1, 0, 0.0

	return furnituremodel.Item{
		Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, DefinitionID: 2, OwnerPlayerID: 7,
		RoomID: &roomID, X: &x, Y: &y, Z: &z,
	}
}

// unitHasStatus reports whether a status key is present in a snapshot's status list.
func unitHasStatus(statuses []worldunit.Status, key string) bool {
	for _, status := range statuses {
		if status.Key == key {
			return true
		}
	}

	return false
}

// TestHandleRejectsMissingRoomPresence verifies the room-presence guard.
func TestHandleRejectsMissingRoomPresence(t *testing.T) {
	handler, _ := handlerForTest(t)
	handler.Players = playersForTest(t, false)

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connectionForTest()},
	})
	if err != nil {
		t.Fatalf("expected no error for missing room presence, got %v", err)
	}
}

// TestHandleIgnoresSoftPickupErrors verifies gameplay misses stay silent and send a bubble alert.
func TestHandleIgnoresSoftPickupErrors(t *testing.T) {
	handler, connections := handlerForTest(t)
	connection, sent := registeredConnectionForTest(t, connections, "conn")
	handler.Furniture = &fakeManager{pickupErr: furnitureservice.ErrItemNotFound}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connection, ItemID: 1},
	})
	if err != nil {
		t.Fatalf("expected soft pickup error to stay silent, got %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != outbubble.Header {
		t.Fatalf("expected a bubble alert packet, got %#v", *sent)
	}
}

// TestHandlePropagatesHardErrors verifies unexpected persistence errors surface.
func TestHandlePropagatesHardErrors(t *testing.T) {
	handler, _ := handlerForTest(t)
	expected := errors.New("pickup failed")
	handler.Furniture = &fakeManager{pickupErr: expected}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connectionForTest(), ItemID: 1},
	})
	if !errors.Is(err, expected) {
		t.Fatalf("expected pickup error, got %v", err)
	}
}

// TestHandlePicksUpAndPublishesWithLogger verifies event publication and logging on success.
func TestHandlePicksUpAndPublishesWithLogger(t *testing.T) {
	handler, connections := handlerForTest(t)
	connection, _ := registeredConnectionForTest(t, connections, "conn")
	publisher := &fakePublisher{}
	handler.Events = publisher
	handler.Log = zap.NewNop()
	handler.Furniture = &fakeManager{pickupResult: pickedItemForTest()}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connection, ItemID: 1},
	})
	if err != nil {
		t.Fatalf("handle pickup: %v", err)
	}
	if len(publisher.events) != 1 || publisher.events[0].Name != pickedupevent.Name {
		t.Fatalf("expected pickedup event published, got %#v", publisher.events)
	}
}

// TestHandleLogsRejectionWithLogger verifies rejected pickups log without error when a logger is set.
func TestHandleLogsRejectionWithLogger(t *testing.T) {
	handler, connections := handlerForTest(t)
	connection, _ := registeredConnectionForTest(t, connections, "conn")
	handler.Log = zap.NewNop()
	handler.Furniture = &fakeManager{pickupErr: furnitureservice.ErrItemNotFound}

	err := handler.Handle(context.Background(), command.Envelope[Command]{
		Command: Command{Handler: connection, ItemID: 1},
	})
	if err != nil {
		t.Fatalf("expected soft pickup error to stay silent, got %v", err)
	}
}

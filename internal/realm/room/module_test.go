package room

import (
	"testing"

	roomentry "github.com/niflaot/pixels/internal/realm/room/entry"
	bannedevent "github.com/niflaot/pixels/internal/realm/room/events/banned"
	roomcreated "github.com/niflaot/pixels/internal/realm/room/events/created"
	roomdeleted "github.com/niflaot/pixels/internal/realm/room/events/deleted"
	roomentered "github.com/niflaot/pixels/internal/realm/room/events/entered"
	kickedevent "github.com/niflaot/pixels/internal/realm/room/events/kicked"
	roomleft "github.com/niflaot/pixels/internal/realm/room/events/left"
	mutedevent "github.com/niflaot/pixels/internal/realm/room/events/muted"
	roomoccupancy "github.com/niflaot/pixels/internal/realm/room/events/occupancychanged"
	rightsgranted "github.com/niflaot/pixels/internal/realm/room/events/rightsgranted"
	rightsrevoked "github.com/niflaot/pixels/internal/realm/room/events/rightsrevoked"
	unbannedevent "github.com/niflaot/pixels/internal/realm/room/events/unbanned"
	unmutedevent "github.com/niflaot/pixels/internal/realm/room/events/unmuted"
	roomupdated "github.com/niflaot/pixels/internal/realm/room/events/updated"
	"github.com/niflaot/pixels/internal/realm/room/layout"
	"github.com/niflaot/pixels/internal/realm/room/service"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
)

// TestEventNames verifies room event names are stable.
func TestEventNames(t *testing.T) {
	events := []string{
		string(roomcreated.Name),
		string(roomupdated.Name),
		string(roomdeleted.Name),
		string(roomoccupancy.Name),
		string(roomentered.Name),
		string(roomleft.Name),
		string(rightsgranted.Name),
		string(rightsrevoked.Name),
		string(kickedevent.Name),
		string(mutedevent.Name),
		string(unmutedevent.Name),
		string(bannedevent.Name),
		string(unbannedevent.Name),
	}

	for _, event := range events {
		if event == "" {
			t.Fatal("expected event name")
		}
	}
}

// TestEntryPermissionNodes verifies room entry nodes are registered.
func TestEntryPermissionNodes(t *testing.T) {
	if EnterAny != "room.enter.any" || EnterFull != "room.enter.full" {
		t.Fatalf("unexpected entry nodes enterAny=%q enterFull=%q", EnterAny, EnterFull)
	}
	nodes := []string{
		string(ModerationOwnKick), string(ModerationOwnMute), string(ModerationOwnBan),
		string(ModerationAnyKick), string(ModerationAnyMute), string(ModerationAnyBan),
		string(RightsOwnGrant), string(RightsOwnRevoke), string(RightsAnyGrant),
		string(RightsAnyRevoke), string(Unkickable),
	}
	for _, node := range nodes {
		if node == "" {
			t.Fatal("expected registered room control node")
		}
	}
}

// TestProvidersExposeContracts verifies module helper providers return contracts.
func TestProvidersExposeContracts(t *testing.T) {
	layoutService := layout.NewService(nil)
	roomService := service.New(nil, layoutService)

	if NewLiveRegistry(bus.New(), netconn.NewRegistry(), roomentry.Config{}, nil) == nil {
		t.Fatal("expected live registry")
	}
	if NewLayoutStore(nil) == nil {
		t.Fatal("expected layout store")
	}
	if NewStore(nil) == nil {
		t.Fatal("expected room store")
	}
	if NewLayoutManager(layoutService) == nil {
		t.Fatal("expected layout manager")
	}
	if NewManager(roomService) == nil {
		t.Fatal("expected room manager")
	}
}

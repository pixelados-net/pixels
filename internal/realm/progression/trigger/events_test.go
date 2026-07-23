package trigger

import (
	"context"
	"testing"

	camerapurchased "github.com/niflaot/pixels/internal/realm/camera/gallery/events/purchased"
	cataloggift "github.com/niflaot/pixels/internal/realm/catalog/events/gift"
	craftcrafted "github.com/niflaot/pixels/internal/realm/crafting/recipe/events/crafted"
	craftdiscovered "github.com/niflaot/pixels/internal/realm/crafting/recipe/events/discovered"
	craftrecycled "github.com/niflaot/pixels/internal/realm/crafting/recycler/events/recycled"
	furnitureplaced "github.com/niflaot/pixels/internal/realm/furniture/events/placed"
	postitplaced "github.com/niflaot/pixels/internal/realm/furniture/events/postitplaced"
	surfaceapplied "github.com/niflaot/pixels/internal/realm/furniture/events/surfaceapplied"
	friendaccepted "github.com/niflaot/pixels/internal/realm/messenger/friend/events/requestaccepted"
	playerignored "github.com/niflaot/pixels/internal/realm/messenger/session/events/ignored"
	guideenrolled "github.com/niflaot/pixels/internal/realm/moderation/events/guideenrolled"
	guidesession "github.com/niflaot/pixels/internal/realm/moderation/events/sessioncompleted"
	planthealed "github.com/niflaot/pixels/internal/realm/pet/breeding/plant/events/healed"
	planttreated "github.com/niflaot/pixels/internal/realm/pet/breeding/plant/events/treated"
	petfed "github.com/niflaot/pixels/internal/realm/pet/care/events/fed"
	petleveled "github.com/niflaot/pixels/internal/realm/pet/care/events/leveled"
	petrespected "github.com/niflaot/pixels/internal/realm/pet/care/events/respected"
	petcreated "github.com/niflaot/pixels/internal/realm/pet/identity/events/created"
	profilerespect "github.com/niflaot/pixels/internal/realm/player/profile/events/respectgranted"
	profileupdated "github.com/niflaot/pixels/internal/realm/player/profile/events/updated"
	roomentered "github.com/niflaot/pixels/internal/realm/room/access/events/entered"
	tradecompleted "github.com/niflaot/pixels/internal/realm/trade/events/completed"
	"github.com/niflaot/pixels/pkg/bus"
)

// adapterHandler maps one event through a concrete progression subscriber.
type adapterHandler func(*Subscriber, context.Context, bus.Event) error

// TestGameplayAdapters verifies each gameplay realm maps to its canonical trigger.
func TestGameplayAdapters(t *testing.T) {
	tests := []struct {
		name     string
		handler  adapterHandler
		payload  any
		expected []progressCall
	}{
		{"room", (*Subscriber).roomEntered, roomentered.Payload{PlayerID: 1}, []progressCall{{1, "room.entered", "", 1, false}}},
		{"profile", (*Subscriber).profileUpdated, profileupdated.Payload{PlayerID: 1, Figure: true, Motto: true}, []progressCall{{1, "player.look.changed", "", 1, true}, {1, "player.motto.changed", "", 1, true}}},
		{"respect", (*Subscriber).respect, profilerespect.Payload{ActorPlayerID: 1, TargetPlayerID: 2}, []progressCall{{1, "respect.given", "", 1, false}, {2, "respect.received", "", 1, false}}},
		{"furniture", (*Subscriber).furniturePlaced, furnitureplaced.Payload{PlayerID: 1, DefinitionID: 44}, []progressCall{{1, "room.furni.count", "44", 1, false}}},
		{"surface", (*Subscriber).surfaceApplied, surfaceapplied.Payload{PlayerID: 1, Surface: "wallpaper"}, []progressCall{{1, "room.deco.wallpaper", "", 1, false}}},
		{"postit", (*Subscriber).postItPlaced, postitplaced.Payload{PlayerID: 1, RoomOwnerID: 2}, []progressCall{{1, "postit.placed", "", 1, false}, {2, "postit.received", "", 1, false}}},
		{"friend", (*Subscriber).friendAccepted, friendaccepted.Payload{PlayerOneID: 1, PlayerTwoID: 2}, []progressCall{{1, "friend.count", "", 1, false}, {2, "friend.count", "", 1, false}, {1, "friend.request.quest", "", 1, false}, {2, "friend.request.quest", "", 1, false}}},
		{"ignored", (*Subscriber).playerIgnored, playerignored.Payload{PlayerID: 1}, []progressCall{{1, "selfmod.ignore", "", 1, false}}},
		{"gift", (*Subscriber).gift, cataloggift.Payload{BuyerID: 1, ReceiverID: 2}, []progressCall{{1, "gift.given", "", 1, false}, {2, "gift.received", "", 1, false}}},
		{"camera", (*Subscriber).cameraPurchased, camerapurchased.Payload{PlayerID: 1}, []progressCall{{1, "camera.photo.purchased", "", 1, false}}},
		{"crafted", (*Subscriber).crafted, craftcrafted.Payload{PlayerID: 1}, []progressCall{{1, "crafting.crafted", "", 1, false}}},
		{"secret", (*Subscriber).discovered, craftdiscovered.Payload{PlayerID: 1}, []progressCall{{1, "crafting.crafted.secret", "", 1, false}}},
		{"recycled", (*Subscriber).recycled, craftrecycled.Payload{PlayerID: 1, ItemCount: 5}, []progressCall{{1, "recycler.recycled", "", 5, false}}},
		{"pet", (*Subscriber).petCreated, petcreated.Payload{OwnerPlayerID: 1, TypeID: 1}, []progressCall{{1, "pet.bought", "", 1, false}}},
		{"plant", (*Subscriber).petCreated, petcreated.Payload{OwnerPlayerID: 1, TypeID: 16}, []progressCall{{1, "plant.bred", "", 1, false}}},
		{"pet level", (*Subscriber).petLeveled, petleveled.Payload{OwnerPlayerID: 1, PreviousLevel: 2, Level: 4}, []progressCall{{1, "pet.level", "", 2, false}}},
		{"pet respect", (*Subscriber).petRespected, petrespected.Payload{ActorPlayerID: 1, OwnerPlayerID: 2}, []progressCall{{1, "pet.respect.given", "", 1, false}, {2, "pet.respect.received", "", 1, false}}},
		{"pet feed", (*Subscriber).petFed, petfed.Payload{PlayerID: 1}, []progressCall{{1, "pet.fed", "", 1, false}}},
		{"plant treatment", (*Subscriber).plantTreated, planttreated.Payload{PlayerID: 1}, []progressCall{{1, "plant.treated", "", 1, false}}},
		{"plant revival", (*Subscriber).plantHealed, planthealed.Payload{PlayerID: 1}, []progressCall{{1, "plant.healed", "", 1, false}}},
		{"guide enrollment", (*Subscriber).guideEnrolled, guideenrolled.Payload{PlayerID: 1}, []progressCall{{1, "guide.enrolled", "", 1, true}}},
		{"guide session", (*Subscriber).guideSession, guidesession.Payload{GuideID: 1, RequesterID: 2, Feedback: true}, []progressCall{{1, "guide.request.handled", "", 1, false}, {2, "guide.feedback.given", "", 1, false}, {2, "guide.requested", "", 1, false}, {1, "guide.recommended", "", 1, false}}},
		{"trade", (*Subscriber).tradeCompleted, tradecompleted.Payload{FirstPlayerID: 1, SecondPlayerID: 2}, []progressCall{{1, "trade.completed", "", 1, false}, {2, "trade.completed", "", 1, false}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			progressor := &captureProgressor{}
			subscriber := &Subscriber{engine: progressor}
			if err := test.handler(subscriber, context.Background(), bus.Event{Payload: test.payload}); err != nil {
				t.Fatal(err)
			}
			assertCalls(t, progressor.calls, test.expected)
		})
	}
}

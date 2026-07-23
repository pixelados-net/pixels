package trigger

import (
	"context"
	"strconv"

	camerapurchased "github.com/niflaot/pixels/internal/realm/camera/gallery/events/purchased"
	cataloggift "github.com/niflaot/pixels/internal/realm/catalog/events/gift"
	craftcrafted "github.com/niflaot/pixels/internal/realm/crafting/recipe/events/crafted"
	craftdiscovered "github.com/niflaot/pixels/internal/realm/crafting/recipe/events/discovered"
	craftrecycled "github.com/niflaot/pixels/internal/realm/crafting/recycler/events/recycled"
	fireworkcharged "github.com/niflaot/pixels/internal/realm/furniture/events/fireworkcharged"
	furnitureplaced "github.com/niflaot/pixels/internal/realm/furniture/events/placed"
	postitplaced "github.com/niflaot/pixels/internal/realm/furniture/events/postitplaced"
	surfaceapplied "github.com/niflaot/pixels/internal/realm/furniture/events/surfaceapplied"
	friendaccepted "github.com/niflaot/pixels/internal/realm/messenger/friend/events/requestaccepted"
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

// fireworkCharged advances the durable FireworksCharger trigger.
func (subscriber *Subscriber) fireworkCharged(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(fireworkcharged.Payload); ok {
		subscriber.progress(ctx, payload.PlayerID, "furniture.firework.charged", 1, false)
	}
	return nil
}

// guideEnrolled advances at most once per UTC day to prevent duty-toggle farming.
func (subscriber *Subscriber) guideEnrolled(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(guideenrolled.Payload); ok {
		subscriber.progress(ctx, payload.PlayerID, "guide.enrolled", 1, true)
	}
	return nil
}

// roomEntered advances room exploration.
func (subscriber *Subscriber) roomEntered(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(roomentered.Payload); ok {
		subscriber.progress(ctx, payload.PlayerID, "room.entered", 1, false)
	}
	return nil
}

// profileUpdated advances only the profile fields changed by the mutation.
func (subscriber *Subscriber) profileUpdated(ctx context.Context, event bus.Event) error {
	payload, ok := event.Payload.(profileupdated.Payload)
	if !ok {
		return nil
	}
	if payload.Figure {
		subscriber.progress(ctx, payload.PlayerID, "player.look.changed", 1, true)
	}
	if payload.Motto {
		subscriber.progress(ctx, payload.PlayerID, "player.motto.changed", 1, true)
	}
	return nil
}

// respect advances both directions of a player respect.
func (subscriber *Subscriber) respect(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(profilerespect.Payload); ok {
		subscriber.progress(ctx, payload.ActorPlayerID, "respect.given", 1, false)
		subscriber.progress(ctx, payload.TargetPlayerID, "respect.received", 1, false)
	}
	return nil
}

// furniturePlaced advances durable room decoration count.
func (subscriber *Subscriber) furniturePlaced(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(furnitureplaced.Payload); ok {
		subscriber.progressData(ctx, payload.PlayerID, "room.furni.count", strconv.FormatInt(payload.DefinitionID, 10), 1)
	}
	return nil
}

// surfaceApplied advances the matching room decoration family.
func (subscriber *Subscriber) surfaceApplied(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(surfaceapplied.Payload); ok {
		subscriber.progress(ctx, payload.PlayerID, "room.deco."+payload.Surface, 1, false)
	}
	return nil
}

// postItPlaced advances note placement and receiving counters.
func (subscriber *Subscriber) postItPlaced(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(postitplaced.Payload); ok {
		subscriber.progress(ctx, payload.PlayerID, "postit.placed", 1, false)
		if payload.RoomOwnerID != payload.PlayerID {
			subscriber.progress(ctx, payload.RoomOwnerID, "postit.received", 1, false)
		}
	}
	return nil
}

// friendAccepted advances the symmetric friendship count.
func (subscriber *Subscriber) friendAccepted(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(friendaccepted.Payload); ok {
		subscriber.progress(ctx, payload.PlayerOneID, "friend.count", 1, false)
		subscriber.progress(ctx, payload.PlayerTwoID, "friend.count", 1, false)
		subscriber.progress(ctx, payload.PlayerOneID, "friend.request.quest", 1, false)
		subscriber.progress(ctx, payload.PlayerTwoID, "friend.request.quest", 1, false)
	}
	return nil
}

// gift advances directional catalog gift counters.
func (subscriber *Subscriber) gift(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(cataloggift.Payload); ok {
		subscriber.progress(ctx, payload.BuyerID, "gift.given", 1, false)
		subscriber.progress(ctx, payload.ReceiverID, "gift.received", 1, false)
	}
	return nil
}

// cameraPurchased advances purchased photo count.
func (subscriber *Subscriber) cameraPurchased(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(camerapurchased.Payload); ok {
		subscriber.progress(ctx, payload.PlayerID, "camera.photo.purchased", 1, false)
	}
	return nil
}

// crafted advances normal crafting count.
func (subscriber *Subscriber) crafted(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(craftcrafted.Payload); ok {
		subscriber.progress(ctx, payload.PlayerID, "crafting.crafted", 1, false)
	}
	return nil
}

// discovered advances secret recipe discoveries.
func (subscriber *Subscriber) discovered(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(craftdiscovered.Payload); ok {
		subscriber.progress(ctx, payload.PlayerID, "crafting.crafted.secret", 1, false)
	}
	return nil
}

// recycled advances the exact number of consumed recycler ingredients.
func (subscriber *Subscriber) recycled(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(craftrecycled.Payload); ok {
		subscriber.progress(ctx, payload.PlayerID, "recycler.recycled", int64(payload.ItemCount), false)
	}
	return nil
}

// petCreated advances pet purchases and monsterplant breeding.
func (subscriber *Subscriber) petCreated(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(petcreated.Payload); ok {
		key := "pet.bought"
		if payload.TypeID == 16 {
			key = "plant.bred"
		}
		subscriber.progress(ctx, payload.OwnerPlayerID, key, 1, false)
	}
	return nil
}

// petLeveled advances the owning player's pet-level counter.
func (subscriber *Subscriber) petLeveled(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(petleveled.Payload); ok {
		subscriber.progress(ctx, payload.OwnerPlayerID, "pet.level", int64(payload.Level-payload.PreviousLevel), false)
	}
	return nil
}

// petRespected advances actor and owner pet-respect counters.
func (subscriber *Subscriber) petRespected(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(petrespected.Payload); ok {
		subscriber.progress(ctx, payload.ActorPlayerID, "pet.respect.given", 1, false)
		subscriber.progress(ctx, payload.OwnerPlayerID, "pet.respect.received", 1, false)
	}
	return nil
}

// petFed advances pet care count.
func (subscriber *Subscriber) petFed(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(petfed.Payload); ok {
		subscriber.progress(ctx, payload.PlayerID, "pet.fed", 1, false)
	}
	return nil
}

// plantTreated advances monsterplant treatment count.
func (subscriber *Subscriber) plantTreated(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(planttreated.Payload); ok {
		subscriber.progress(ctx, payload.PlayerID, "plant.treated", 1, false)
	}
	return nil
}

// plantHealed advances monsterplant revival count.
func (subscriber *Subscriber) plantHealed(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(planthealed.Payload); ok {
		subscriber.progress(ctx, payload.PlayerID, "plant.healed", 1, false)
	}
	return nil
}

// guideSession advances both participants and recommendation state.
func (subscriber *Subscriber) guideSession(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(guidesession.Payload); ok {
		subscriber.progress(ctx, payload.GuideID, "guide.request.handled", 1, false)
		subscriber.progress(ctx, payload.RequesterID, "guide.feedback.given", 1, false)
		subscriber.progress(ctx, payload.RequesterID, "guide.requested", 1, false)
		if payload.Feedback {
			subscriber.progress(ctx, payload.GuideID, "guide.recommended", 1, false)
		}
	}
	return nil
}

// tradeCompleted advances both trade participants.
func (subscriber *Subscriber) tradeCompleted(ctx context.Context, event bus.Event) error {
	if payload, ok := event.Payload.(tradecompleted.Payload); ok {
		subscriber.progress(ctx, payload.FirstPlayerID, "trade.completed", 1, false)
		subscriber.progress(ctx, payload.SecondPlayerID, "trade.completed", 1, false)
	}
	return nil
}

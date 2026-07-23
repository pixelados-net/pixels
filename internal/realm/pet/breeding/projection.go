package breeding

import (
	"context"

	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outfailure "github.com/niflaot/pixels/networking/outbound/room/pet/breeding/failure"
	outrequest "github.com/niflaot/pixels/networking/outbound/room/pet/breeding/request"
	petdata "github.com/niflaot/pixels/networking/pet/data"
)

// breedingNest returns the first stable pet breeding nest in one room.
func breedingNest(active *roomlive.Room) (int64, bool) {
	for _, item := range active.FurnitureItems() {
		if item.Definition.InteractionType == "pet_breeding_nest" || item.Definition.InteractionType == "breeding_nest" || item.Definition.InteractionType == "pet_nest" {
			return item.ID, true
		}
	}
	return 0, false
}

// parent maps one parent to the confirmation preview.
func parent(pet petrecord.Pet) outrequest.Parent {
	return outrequest.Parent{ID: pet.ID, Name: pet.Name, Level: pet.Level, Figure: petdata.FigureString(petruntime.InventoryPet(pet).Figure), OwnerName: pet.OwnerName}
}

// ownedPet returns the parent's id owned by the actor.
func ownedPet(actorID int64, first petrecord.Pet, second petrecord.Pet) int64 {
	if first.OwnerPlayerID == actorID {
		return first.ID
	}
	return second.ID
}

// otherPet returns the parent not selected as owned.
func otherPet(actorID int64, first petrecord.Pet, second petrecord.Pet) int64 {
	if first.OwnerPlayerID == actorID {
		return second.ID
	}
	return first.ID
}

// stateCode maps durable state to Nitro's breeding state integer.
func stateCode(value string) int32 {
	if value == "confirmed" {
		return 1
	}
	return 0
}

// sendFailure sends one native breeding failure.
func (service *Service) sendFailure(ctx context.Context, target netconn.Context, reason int32) error {
	packet, err := outfailure.Encode(reason)
	if err != nil {
		return err
	}
	return target.Send(ctx, packet)
}

// sendOwner performs best-effort delivery to one online parent owner.
func (service *Service) sendOwner(ctx context.Context, playerID int64, packet codec.Packet) {
	if service.players == nil || service.connections == nil {
		return
	}
	player, found := service.players.Find(playerID)
	if !found {
		return
	}
	peer := player.Peer()
	if target, targetFound := service.connections.Get(peer.ConnectionKind(), peer.ConnectionID()); targetFound {
		_ = target.Send(ctx, packet)
	}
}

// firstError chooses infrastructure failures before domain fallbacks.
func firstError(err error, fallback error) error {
	if err != nil {
		return err
	}
	return fallback
}

package effect

import (
	"context"
	"math"
	"time"

	effectenabled "github.com/niflaot/pixels/internal/realm/player/events/effectenabled"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomprojection "github.com/niflaot/pixels/internal/realm/room/runtime/projection"
	"github.com/niflaot/pixels/networking/codec"
	outeffect "github.com/niflaot/pixels/networking/outbound/room/entities/effect"
	outadd "github.com/niflaot/pixels/networking/outbound/user/effect/add"
	outlist "github.com/niflaot/pixels/networking/outbound/user/effect/list"
	outselected "github.com/niflaot/pixels/networking/outbound/user/effect/selected"
)

// SendInventory sends one player's complete Nitro effect inventory.
func (service *Service) SendInventory(ctx context.Context, playerID int64) error {
	effects, err := service.List(ctx, playerID)
	if err != nil {
		return err
	}
	now := service.now()
	records := make([]outlist.Effect, 0, len(effects))
	for _, item := range effects {
		records = append(records, listEffect(item, now))
	}
	packet, err := outlist.Encode(records)
	if err != nil {
		return err
	}
	if err = service.send(ctx, playerID, packet); err != nil {
		return err
	}
	active, err := service.store.Active(ctx, playerID)
	if err != nil || active == nil {
		return err
	}
	packet, err = outselected.Encode(*active)
	if err != nil {
		return err
	}
	return service.send(ctx, playerID, packet)
}

// sendAdded sends one incremental effect inventory update.
func (service *Service) sendAdded(ctx context.Context, item Effect) error {
	record := listEffect(item, service.now())
	packet, err := outadd.Encode(record.Type, record.SubType, record.Duration, record.Permanent)
	if err != nil {
		return err
	}
	return service.send(ctx, item.PlayerID, packet)
}

// projectSelection sends one selected effect to its owner and current room.
func (service *Service) projectSelection(ctx context.Context, playerID int64, effectID int32, source Source) error {
	if service.players != nil {
		if player, found := service.players.Find(playerID); found {
			var selected *int32
			if effectID > 0 {
				selected = &effectID
			}
			player.SetActiveEffect(selected)
		}
	}
	packet, err := outselected.Encode(effectID)
	if err != nil {
		return err
	}
	if err = service.send(ctx, playerID, packet); err != nil {
		return err
	}
	if service.rooms == nil {
		service.publish(ctx, effectenabled.Name, effectenabled.Payload{PlayerID: playerID, EffectID: effectID, Source: string(source)})
		return nil
	}
	active, found := service.rooms.FindByPlayer(playerID)
	if found {
		unit, unitFound := active.SetUnitEffect(playerID, effectID)
		if unitFound {
			roomPacket, encodeErr := outeffect.Encode(unit.UnitID, roomprojection.EffectID(unit), 0)
			if encodeErr != nil {
				return encodeErr
			}
			err = broadcast.RoomPacket(ctx, service.connections, active, roomPacket, 0)
		}
	}
	service.publish(ctx, effectenabled.Name, effectenabled.Payload{PlayerID: playerID, EffectID: effectID, Source: string(source)})
	return err
}

// send writes one packet when the player remains online.
func (service *Service) send(ctx context.Context, playerID int64, packet codec.Packet) error {
	if service.players == nil || service.connections == nil {
		return nil
	}
	player, found := service.players.Find(playerID)
	if !found {
		return nil
	}
	peer := player.Peer()
	connection, found := service.connections.Get(peer.ConnectionKind(), peer.ConnectionID())
	if !found {
		return nil
	}
	return connection.Send(ctx, packet)
}

// listEffect maps one domain effect into Nitro inventory fields.
func listEffect(item Effect, now time.Time) outlist.Effect {
	inactive := item.RemainingCharges
	if item.ActivatedAt != nil && inactive > 0 {
		inactive--
	}
	return outlist.Effect{Type: item.ID, Duration: wireDuration(item), InactiveEffectsInInventory: inactive,
		SecondsLeftIfActive: item.SecondsLeft(now), Permanent: item.Permanent()}
}

// wireDuration maps permanent effects to Nitro's non-expiring duration sentinel.
func wireDuration(item Effect) int32 {
	if item.Permanent() {
		return math.MaxInt32
	}
	return item.DurationSeconds
}

// rankEffect resolves one synthetic primary-group effect without persistence.
func (service *Service) rankEffect(ctx context.Context, playerID int64) (Effect, bool, error) {
	if service.permissions == nil {
		return Effect{}, false, nil
	}
	group, found, err := service.permissions.PrimaryGroup(ctx, playerID)
	if err != nil || !found || group.RoomEffectID == nil {
		return Effect{}, false, err
	}
	return Effect{PlayerID: playerID, ID: *group.RoomEffectID, RemainingCharges: 1, Synthetic: true}, true, nil
}

package broadcast

import (
	"context"
	"errors"

	roomdoorbell "github.com/niflaot/pixels/internal/realm/room/access/doorbell"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	outdoorbelldenied "github.com/niflaot/pixels/networking/outbound/room/doorbell/denied"
)

// NewDoorbellPublisher creates an expired doorbell request broadcaster.
func NewDoorbellPublisher(connections *netconn.Registry) roomlive.DoorbellPublisher {
	return func(ctx context.Context, active *roomlive.Room, expired []roomdoorbell.Expired) error {
		if connections == nil || active == nil || len(expired) == 0 {
			return nil
		}
		var result error
		for _, item := range expired {
			result = errors.Join(result, notifyExpiredWaiter(ctx, item), notifyExpiredOwner(ctx, connections, active, item))
		}

		return result
	}
}

// notifyExpiredWaiter rejects one waiting player.
func notifyExpiredWaiter(ctx context.Context, expired roomdoorbell.Expired) error {
	denied, err := outdoorbelldenied.Encode("")
	if err != nil {
		return err
	}
	if err := expired.Handler.Send(ctx, denied); err != nil {
		return err
	}

	return nil
}

// notifyExpiredOwner closes the owner's waiting-player prompt.
func notifyExpiredOwner(ctx context.Context, connections *netconn.Registry, active *roomlive.Room, expired roomdoorbell.Expired) error {
	packet, err := outdoorbelldenied.Encode(expired.Username)
	if err != nil {
		return err
	}
	ownerID := active.Snapshot().OwnerPlayerID
	for _, occupant := range active.Occupants() {
		if occupant.PlayerID != ownerID {
			continue
		}
		connection, found := connections.Get(occupant.ConnectionKind, occupant.ConnectionID)
		if found {
			return connection.Send(ctx, packet)
		}
	}

	return nil
}

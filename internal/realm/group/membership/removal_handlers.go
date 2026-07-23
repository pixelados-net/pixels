package membership

import (
	"context"
	"errors"

	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	groupruntime "github.com/niflaot/pixels/internal/realm/group/runtime"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inclearfavorite "github.com/niflaot/pixels/networking/inbound/group/membership/favorite/clear"
	insetfavorite "github.com/niflaot/pixels/networking/inbound/group/membership/favorite/set"
	inremove "github.com/niflaot/pixels/networking/inbound/group/membership/remove"
	inconfirm "github.com/niflaot/pixels/networking/inbound/group/membership/remove/confirm"
	outinfo "github.com/niflaot/pixels/networking/outbound/group/identity/info"
	outfavorite "github.com/niflaot/pixels/networking/outbound/group/membership/favorite/update"
	outconfirm "github.com/niflaot/pixels/networking/outbound/group/membership/remove/confirm"
	outprofilechanged "github.com/niflaot/pixels/networking/outbound/user/profile/changed"
	"github.com/niflaot/pixels/pkg/i18n"
)

// confirmRemoval sends the current durable HQ furniture count.
func (handler Handler) confirmRemoval(connection netconn.Context, packet codec.Packet) error {
	payload, err := inconfirm.Decode(packet)
	if err != nil {
		return err
	}
	actorID, err := handler.Delivery.PlayerID(connection)
	if err != nil {
		return err
	}
	count, err := handler.Membership.ConfirmRemoval(context.Background(), actorID, payload.GroupID, payload.PlayerID)
	if err != nil {
		return handler.feedback(connection, err, "group.member.rank_forbidden")
	}
	response, err := outconfirm.Encode(payload.PlayerID, int32(count))
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// remove leaves a group, cancels a request, or removes a lower-ranked member.
func (handler Handler) remove(connection netconn.Context, packet codec.Packet) error {
	payload, err := inremove.Decode(packet)
	if err != nil {
		return err
	}
	actorID, err := handler.Delivery.PlayerID(connection)
	if err != nil {
		return err
	}
	if _, err = handler.Membership.Remove(context.Background(), actorID, payload.GroupID, payload.PlayerID); err != nil {
		return handler.feedback(connection, err, "group.member.remove_failed")
	}
	if actorID != payload.PlayerID {
		_ = handler.projectInformation(payload.PlayerID, payload.GroupID)
		return nil
	}
	if err = handler.sendGroups(connection, false); err != nil {
		return err
	}
	return handler.sendInformation(connection, actorID, payload.GroupID)
}

// sendInformation refreshes Nitro's current membership controls for one group.
func (handler Handler) sendInformation(connection netconn.Context, playerID int64, groupID int64) error {
	response, err := handler.informationPacket(playerID, groupID)
	if err != nil {
		return handler.feedback(connection, err, "group.member.remove_failed")
	}
	return connection.Send(context.Background(), response)
}

// projectInformation refreshes one online target's current membership controls.
func (handler Handler) projectInformation(playerID int64, groupID int64) error {
	response, err := handler.informationPacket(playerID, groupID)
	if err != nil {
		return err
	}
	_, err = handler.Delivery.Send(context.Background(), playerID, response)
	return err
}

// informationPacket builds one player's authoritative group information projection.
func (handler Handler) informationPacket(playerID int64, groupID int64) (codec.Packet, error) {
	group, role, member, pending, favorite, err := handler.Membership.Information(context.Background(), playerID, groupID)
	if err != nil {
		return codec.Packet{}, err
	}
	return outinfo.Encode(group, role, member, pending, favorite, false, member && role <= grouprecord.Admin)
}

// setFavorite selects one active membership and updates Nitro immediately.
func (handler Handler) setFavorite(connection netconn.Context, packet codec.Packet) error {
	groupID, err := insetfavorite.Decode(packet)
	if err != nil {
		return err
	}
	return handler.favorite(connection, &groupID)
}

// clearFavorite removes one active favorite selection.
func (handler Handler) clearFavorite(connection netconn.Context, packet codec.Packet) error {
	if _, err := inclearfavorite.Decode(packet); err != nil {
		return err
	}
	return handler.favorite(connection, nil)
}

// favorite persists and sends the session-local preference projection.
func (handler Handler) favorite(connection netconn.Context, groupID *int64) error {
	playerID, err := handler.Delivery.PlayerID(connection)
	if err != nil {
		return err
	}
	if err = handler.Membership.SetFavorite(context.Background(), playerID, groupID); err != nil {
		return handler.feedback(connection, err, "group.join.closed")
	}
	selectedID, status, name := int64(-1), int32(-1), ""
	if groupID != nil {
		selectedID = *groupID
		groups, listErr := handler.Membership.PlayerGroups(context.Background(), playerID)
		if listErr != nil {
			return listErr
		}
		for _, item := range groups {
			if item.Group.ID == selectedID {
				status, name = int32(item.Role), item.Group.Name
				break
			}
		}
	}
	response, err := outfavorite.Encode(-1, selectedID, status, name)
	if err != nil {
		return err
	}
	if err = connection.Send(context.Background(), response); err != nil {
		return err
	}
	response, err = outprofilechanged.Encode(int32(playerID))
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// feedback maps one expected mutation failure to a localized alert.
func (handler Handler) feedback(connection netconn.Context, err error, fallback string) error {
	key := fallback
	if errors.Is(err, grouprecord.ErrLimit) {
		key = "group.member.furniture_cleanup_limit"
	} else if errors.Is(err, grouprecord.ErrConflict) {
		key = "group.edit.conflict"
	} else if errors.Is(err, grouprecord.ErrForbidden) {
		key = "group.member.rank_forbidden"
	}
	return groupruntime.SendError(context.Background(), connection, handler.Translations, i18n.Key(key))
}

// joinFailure maps durable admission errors to Nitro reason values.
func joinFailure(err error) int32 {
	if errors.Is(err, grouprecord.ErrLimit) {
		return 2
	}
	return 1
}

// managementFailure maps protected role errors to Nitro reason values.
func managementFailure(err error) int32 {
	if errors.Is(err, grouprecord.ErrConflict) {
		return 2
	}
	return 1
}

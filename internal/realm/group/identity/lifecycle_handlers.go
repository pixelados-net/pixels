package identity

import (
	"context"
	"errors"

	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	groupruntime "github.com/niflaot/pixels/internal/realm/group/runtime"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	insavebadge "github.com/niflaot/pixels/networking/inbound/group/badge/save"
	indelete "github.com/niflaot/pixels/networking/inbound/group/identity/delete"
	outdeactivated "github.com/niflaot/pixels/networking/outbound/group/identity/deactivated"
	"github.com/niflaot/pixels/pkg/i18n"
)

// saveBadge handles owner badge edits.
func (handler Handler) saveBadge(connection netconn.Context, packet codec.Packet) error {
	payload, err := insavebadge.Decode(packet)
	if err != nil {
		return err
	}
	playerID, err := handler.Delivery.PlayerID(connection)
	if err != nil {
		return err
	}
	group, found, err := handler.Identity.Group(context.Background(), payload.GroupID, false)
	if err != nil || !found {
		return handler.feedback(connection, "group.deactivated")
	}
	if _, err = handler.Identity.SaveBadge(context.Background(), playerID, payload.GroupID, group.Version, payload.Parts); err != nil {
		return handler.feedback(connection, "group.create.badge_invalid")
	}
	return nil
}

// deactivate handles owner group deletion.
func (handler Handler) deactivate(connection netconn.Context, packet codec.Packet) error {
	groupID, err := indelete.Decode(packet)
	if err != nil {
		return err
	}
	playerID, err := handler.Delivery.PlayerID(connection)
	if err != nil {
		return err
	}
	group, found, err := handler.Identity.Group(context.Background(), groupID, false)
	if err != nil || !found {
		return handler.feedback(connection, "group.deactivated")
	}
	if _, err = handler.Identity.Deactivate(context.Background(), playerID, groupID, group.Version); err != nil {
		return handler.feedback(connection, "group.edit.forbidden")
	}
	response, err := outdeactivated.Encode(groupID)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// feedback sends one localized expected error.
func (handler Handler) feedback(connection netconn.Context, key string) error {
	return groupruntime.SendError(context.Background(), connection, handler.Translations, i18n.Key(key))
}

// errorKey maps common domain errors to stable translation keys.
func errorKey(err error, fallback string) string {
	if errors.Is(err, grouprecord.ErrLimit) {
		return "group.create.limit"
	}
	if errors.Is(err, grouprecord.ErrForbidden) {
		return "group.create.club_required"
	}
	return fallback
}

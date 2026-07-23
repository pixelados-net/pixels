package session

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	messengerservice "github.com/niflaot/pixels/internal/realm/messenger/core"
	"github.com/niflaot/pixels/internal/realm/messenger/runtime/delivery"
	ignoredevent "github.com/niflaot/pixels/internal/realm/messenger/session/events/ignored"
	netconn "github.com/niflaot/pixels/networking/connection"
	outlist "github.com/niflaot/pixels/networking/outbound/user/ignore/list"
	outresult "github.com/niflaot/pixels/networking/outbound/user/ignore/result"
	"github.com/niflaot/pixels/pkg/bus"
)

const (
	// ListName identifies ignored-user loading.
	ListName command.Name = "messenger.privacy.ignore.list"
	// IgnoreName identifies ignored-user creation.
	IgnoreName command.Name = "messenger.privacy.ignore.add"
	// UnignoreName identifies ignored-user removal.
	UnignoreName command.Name = "messenger.privacy.ignore.remove"
	// RelationshipsName identifies public relationship loading.
	RelationshipsName command.Name = "messenger.privacy.relationships"
)

// Command carries one privacy operation.
type PrivacyCommand struct {
	// Connection stores the requesting connection.
	Connection netconn.Context
	// Name stores the command identity.
	Name command.Name
	// Username optionally identifies a target by name.
	Username string
	// PlayerID optionally identifies a target by id.
	PlayerID int64
}

// CommandName returns the selected privacy command identity.
func (input PrivacyCommand) CommandName() command.Name { return input.Name }

// Handler executes privacy commands.
type PrivacyHandler struct {
	// Messenger stores privacy persistence and projections.
	Messenger *messengerservice.Service
	// Delivery resolves authenticated players.
	Delivery *delivery.Sender
	// Events publishes committed privacy mutations.
	Events bus.Publisher
}

// Handle executes one privacy command.
func (handler PrivacyHandler) Handle(ctx context.Context, input PrivacyCommand) error {
	actorID, err := handler.Delivery.PlayerID(input.Connection)
	if err != nil {
		return err
	}
	switch input.Name {
	case ListName:
		return handler.sendIgnored(ctx, input.Connection, actorID)
	case IgnoreName:
		return handler.ignore(ctx, input, actorID)
	case UnignoreName:
		return handler.unignore(ctx, input, actorID)
	case RelationshipsName:
		return handler.sendRelationships(ctx, input.Connection, input.PlayerID)
	default:
		return nil
	}
}

// sendIgnored loads and projects ignored usernames.
func (handler PrivacyHandler) sendIgnored(ctx context.Context, connection netconn.Context, actorID int64) error {
	items, err := handler.Messenger.Ignored(ctx, actorID)
	if err != nil {
		return err
	}
	names := make([]string, len(items))
	for index := range items {
		names[index] = items[index].Username
	}
	packet, err := outlist.Encode(names)
	if err != nil {
		return err
	}
	return connection.Send(ctx, packet)
}

// ignore applies one username or id ignore.
func (handler PrivacyHandler) ignore(ctx context.Context, input PrivacyCommand, actorID int64) error {
	item, found, err := handler.Messenger.IgnoreByName(ctx, actorID, input.Username)
	if input.PlayerID > 0 {
		item, found, err = handler.Messenger.IgnoreByID(ctx, actorID, input.PlayerID)
	}
	if err != nil {
		return err
	}
	state := outresult.Failed
	if found || item.PlayerID > 0 {
		state = outresult.Ignored
		if handler.Events != nil {
			_ = handler.Events.Publish(ctx, bus.Event{Name: ignoredevent.Name, Payload: ignoredevent.Payload{PlayerID: actorID, IgnoredPlayerID: item.PlayerID}})
		}
	}
	packet, err := outresult.Encode(state, item.Username)
	if err != nil {
		return err
	}
	return input.Connection.Send(ctx, packet)
}

// unignore removes one username ignore.
func (handler PrivacyHandler) unignore(ctx context.Context, input PrivacyCommand, actorID int64) error {
	item, _, err := handler.Messenger.UnignoreByName(ctx, actorID, input.Username)
	if err != nil {
		return err
	}
	packet, err := outresult.Encode(outresult.Unignored, item.Username)
	if err != nil {
		return err
	}
	return input.Connection.Send(ctx, packet)
}

// sendRelationships sends public relationship category summaries.
func (handler PrivacyHandler) sendRelationships(ctx context.Context, connection netconn.Context, playerID int64) error {
	items, err := handler.Messenger.Relationships(ctx, playerID)
	if err != nil {
		return err
	}
	packet, err := delivery.RelationshipPacket(playerID, items)
	if err != nil {
		return err
	}
	return connection.Send(ctx, packet)
}

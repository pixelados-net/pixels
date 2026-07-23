// Package session handles messenger bootstrap, refresh, search, and discovery.
package session

import (
	"context"
	"time"

	"github.com/niflaot/pixels/internal/command"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	messengerservice "github.com/niflaot/pixels/internal/realm/messenger/core"
	"github.com/niflaot/pixels/internal/realm/messenger/runtime/delivery"
	netconn "github.com/niflaot/pixels/networking/connection"
	outinit "github.com/niflaot/pixels/networking/outbound/messenger/session/init"
	outsearch "github.com/niflaot/pixels/networking/outbound/messenger/session/search"
	outfind "github.com/niflaot/pixels/networking/outbound/messenger/social/findroom"
	outforward "github.com/niflaot/pixels/networking/outbound/room/forward"
	outprofile "github.com/niflaot/pixels/networking/outbound/user/profile"
)

const (
	// InitName identifies messenger initialization.
	InitName command.Name = "messenger.session.init"
	// RefreshName identifies friend-list refresh.
	RefreshName command.Name = "messenger.session.refresh"
	// RequestsName identifies pending-request refresh.
	RequestsName command.Name = "messenger.session.requests"
	// SearchName identifies player search.
	SearchName command.Name = "messenger.session.search"
	// FindRoomName identifies populated-room discovery.
	FindRoomName command.Name = "messenger.session.find_room"
	// ProfileName identifies public profile requests.
	ProfileName command.Name = "messenger.session.profile"
)

// ConnectionCommand carries a requesting connection.
type ConnectionCommand struct {
	// Connection stores the requesting connection.
	Connection netconn.Context
	// Name stores the command identity.
	Name command.Name
}

// CommandName returns the configured command identity.
func (input ConnectionCommand) CommandName() command.Name { return input.Name }

// SearchCommand requests a username-prefix search.
type SearchCommand struct {
	// Connection stores the requesting connection.
	Connection netconn.Context
	// Term stores the client search term.
	Term string
}

// ProfileCommand requests one public user profile.
type ProfileCommand struct {
	// Connection stores the requesting connection.
	Connection netconn.Context
	// PlayerID identifies the requested profile.
	PlayerID int64
	// OpenWindow reports whether Nitro should open the profile window.
	OpenWindow bool
}

// CommandName returns the profile command identity.
func (ProfileCommand) CommandName() command.Name { return ProfileName }

// CommandName returns the search command identity.
func (SearchCommand) CommandName() command.Name { return SearchName }

// Handler executes messenger session commands.
type Handler struct {
	// Messenger stores messenger behavior.
	Messenger *messengerservice.Service
	// Delivery sends packets and resolves actors.
	Delivery *delivery.Sender
	// Groups reads immutable social-group profile entries.
	Groups ProfileGroups
	// Players resolves exact usernames for compatibility profile requests.
	Players PlayerFinder
}

// ProfileGroups reads one player's active social-group entries.
type ProfileGroups interface {
	// PlayerGroups returns active memberships and favorite state.
	PlayerGroups(context.Context, int64) ([]grouprecord.PlayerGroup, error)
}

// HandleInit sends limits and friend cards.
func (handler Handler) HandleInit(ctx context.Context, input ConnectionCommand) error {
	playerID, err := handler.Delivery.PlayerID(input.Connection)
	if err != nil {
		return err
	}
	effective, normal, club, err := handler.Messenger.Limits(ctx, playerID)
	if err != nil {
		return err
	}
	packet, err := outinit.Encode(int32(effective), int32(normal), int32(club))
	if err == nil {
		err = input.Connection.Send(ctx, packet)
	}
	if err == nil {
		err = handler.sendFriends(ctx, input.Connection, playerID)
	}
	return err
}

// HandleRefresh sends the complete current friend list.
func (handler Handler) HandleRefresh(ctx context.Context, input ConnectionCommand) error {
	playerID, err := handler.Delivery.PlayerID(input.Connection)
	if err != nil {
		return err
	}
	return handler.sendFriends(ctx, input.Connection, playerID)
}

// HandleRequests sends pending requests and the embedded pending count.
func (handler Handler) HandleRequests(ctx context.Context, input ConnectionCommand) error {
	playerID, err := handler.Delivery.PlayerID(input.Connection)
	if err != nil {
		return err
	}
	return handler.sendRequests(ctx, input.Connection, playerID)
}

// HandleSearch sends a Nitro-native split friend and non-friend search response.
func (handler Handler) HandleSearch(ctx context.Context, input SearchCommand) error {
	playerID, err := handler.Delivery.PlayerID(input.Connection)
	if err != nil {
		return err
	}
	result, err := handler.Messenger.Search(ctx, playerID, input.Term)
	if err != nil || result.Throttled {
		return err
	}
	friendIDs, err := handler.Messenger.FriendIDs(ctx, playerID)
	if err != nil {
		return err
	}
	known := make(map[int64]struct{}, len(friendIDs))
	for _, friendID := range friendIDs {
		known[friendID] = struct{}{}
	}
	friends := make([]outsearch.Result, 0, len(result.Results))
	others := make([]outsearch.Result, 0, len(result.Results))
	for _, item := range result.Results {
		projected := outsearch.Result{PlayerID: item.PlayerID, Username: item.Username, Motto: item.Motto, Online: item.Online, CanFollow: item.CanFollow, Gender: item.Gender, Look: item.Look}
		if _, found := known[item.PlayerID]; found {
			friends = append(friends, projected)
		} else {
			others = append(others, projected)
		}
	}
	packet, err := outsearch.Encode(friends, others)
	if err != nil {
		return err
	}
	return input.Connection.Send(ctx, packet)
}

// HandleFindRoom forwards the player to one populated active room.
func (handler Handler) HandleFindRoom(ctx context.Context, input ConnectionCommand) error {
	roomID, findErr := handler.Messenger.FindPopulatedRoom()
	success := findErr == nil
	packet, err := outfind.Encode(success)
	if err != nil {
		return err
	}
	if err = input.Connection.Send(ctx, packet); err != nil || !success {
		return err
	}
	packet, err = outforward.Encode(int32(roomID))
	if err != nil {
		return err
	}
	return input.Connection.Send(ctx, packet)
}

// HandleProfile sends Nitro's native public user profile.
func (handler Handler) HandleProfile(ctx context.Context, input ProfileCommand) error {
	viewerID, err := handler.Delivery.PlayerID(input.Connection)
	if err != nil {
		return err
	}
	profile, err := handler.Messenger.Profile(ctx, viewerID, input.PlayerID)
	if err != nil {
		return err
	}
	registration := ""
	if !profile.Record.Player.CreatedAt.IsZero() {
		registration = profile.Record.Player.CreatedAt.Format("02-01-2006")
	}
	secondsSinceVisit := int32(0)
	if profile.Record.Player.LastSeenAt != nil {
		seconds := time.Since(*profile.Record.Player.LastSeenAt) / time.Second
		if seconds > 0 && seconds < time.Duration(^uint32(0)>>1) {
			secondsSinceVisit = int32(seconds)
		}
	}
	groups := []grouprecord.PlayerGroup{}
	if handler.Groups != nil {
		groups, err = handler.Groups.PlayerGroups(ctx, input.PlayerID)
		if err != nil {
			return err
		}
	}
	packet, err := outprofile.Encode(profile.Record.Player.ID, profile.Record.Player.Username, profile.Record.Profile.Look,
		profile.Record.Profile.Motto, registration, groups, int32(profile.FriendCount), profile.IsFriend, profile.RequestSent,
		profile.Online, secondsSinceVisit, input.OpenWindow)
	if err != nil {
		return err
	}
	if err = input.Connection.Send(ctx, packet); err != nil {
		return err
	}
	if input.OpenWindow {
		handler.Messenger.ViewProfile(viewerID, input.PlayerID)
	}
	return nil
}

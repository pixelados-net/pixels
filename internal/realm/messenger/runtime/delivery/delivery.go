// Package delivery sends messenger packets through existing session bindings.
package delivery

import (
	"context"

	messengermodel "github.com/niflaot/pixels/internal/realm/messenger/record"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	friendcard "github.com/niflaot/pixels/networking/outbound/messenger/friend/card"
	outrelationships "github.com/niflaot/pixels/networking/outbound/user/relationships"
)

// Sender delivers packets to authenticated players without owning presence state.
type Sender struct {
	// bindings maps players to active connections.
	bindings *binding.Registry
	// connections stores active transport-agnostic sessions.
	connections *netconn.Registry
}

// New creates messenger packet delivery behavior.
func New(bindings *binding.Registry, connections *netconn.Registry) *Sender {
	return &Sender{bindings: bindings, connections: connections}
}

// PlayerID resolves the authenticated player behind one connection.
func (sender *Sender) PlayerID(connection netconn.Context) (int64, error) {
	if sender == nil || sender.bindings == nil {
		return 0, binding.ErrBindingNotFound
	}
	bound, found := sender.bindings.FindByConnection(binding.ConnectionKey{Kind: connection.ConnectionKind, ID: connection.ConnectionID})
	if !found {
		return 0, binding.ErrBindingNotFound
	}
	return bound.PlayerID, nil
}

// Send sends one packet to an online player and reports whether delivery occurred.
func (sender *Sender) Send(ctx context.Context, playerID int64, packet codec.Packet) (bool, error) {
	if sender == nil || sender.bindings == nil || sender.connections == nil {
		return false, nil
	}
	bound, found := sender.bindings.FindByPlayer(playerID)
	if !found {
		return false, nil
	}
	connection, found := sender.connections.Get(bound.ConnectionKind, bound.ConnectionID)
	if !found {
		return false, nil
	}
	return true, connection.Send(ctx, packet)
}

// Online reports whether one player has a live bound connection.
func (sender *Sender) Online(playerID int64) bool {
	if sender == nil || sender.bindings == nil || sender.connections == nil {
		return false
	}
	bound, found := sender.bindings.FindByPlayer(playerID)
	if !found {
		return false
	}
	_, found = sender.connections.Get(bound.ConnectionKind, bound.ConnectionID)
	return found
}

// FriendCard maps a protocol-neutral card into Nitro's wire projection.
func FriendCard(card messengermodel.Card) friendcard.Card {
	return friendcard.Card{
		PlayerID: card.ID, Username: card.Username, Gender: card.Gender, Online: card.Online,
		FollowingAllowed: card.FollowingAllowed, Look: card.Look, CategoryID: card.CategoryID,
		Motto: card.Motto, Relationship: uint16(card.Relation),
	}
}

// FriendCards maps protocol-neutral cards into Nitro's wire projection.
func FriendCards(cards []messengermodel.Card) []friendcard.Card {
	result := make([]friendcard.Card, len(cards))
	for index, card := range cards {
		result[index] = FriendCard(card)
	}
	return result
}

// RelationshipPacket maps public summaries into Nitro's wire projection.
func RelationshipPacket(playerID int64, items []messengermodel.RelationshipSummary) (codec.Packet, error) {
	entries := make([]outrelationships.Entry, len(items))
	for index := range items {
		entries[index] = outrelationships.Entry{
			Type: int32(items[index].Relation), Count: items[index].Count,
			FriendID: items[index].Sample.PlayerID, FriendName: items[index].Sample.Username,
			FriendLook: items[index].SampleLook,
		}
	}
	return outrelationships.Encode(playerID, entries)
}

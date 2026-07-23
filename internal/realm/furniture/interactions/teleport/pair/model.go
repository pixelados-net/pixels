// Package pair manages durable furniture teleport pairings.
package pair

import "errors"

var (
	// ErrInvalidPair reports malformed or self-referencing item ids.
	ErrInvalidPair = errors.New("invalid teleport pair")
	// ErrItemNotFound reports a missing furniture item.
	ErrItemNotFound = errors.New("teleport furniture item not found")
	// ErrNotTeleport reports a furniture definition without teleport behavior.
	ErrNotTeleport = errors.New("furniture item is not a teleport")
	// ErrNotOwner reports pairing by a player who does not own both items.
	ErrNotOwner = errors.New("player does not own both teleport items")
)

// Pair stores one symmetric teleport item relationship.
type Pair struct {
	// ItemOneID stores the lower canonical item id.
	ItemOneID int64
	// ItemTwoID stores the higher canonical item id.
	ItemTwoID int64
}

// New canonicalizes two distinct positive item ids.
func New(first int64, second int64) (Pair, error) {
	if first <= 0 || second <= 0 || first == second {
		return Pair{}, ErrInvalidPair
	}
	if first > second {
		first, second = second, first
	}

	return Pair{ItemOneID: first, ItemTwoID: second}, nil
}

// Other returns the item paired with the supplied id.
func (pair Pair) Other(itemID int64) (int64, bool) {
	switch itemID {
	case pair.ItemOneID:
		return pair.ItemTwoID, true
	case pair.ItemTwoID:
		return pair.ItemOneID, true
	default:
		return 0, false
	}
}

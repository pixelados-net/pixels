package service

import "errors"

var (
	// ErrInvalidPlayerID reports a malformed player id.
	ErrInvalidPlayerID = errors.New("invalid furniture player id")

	// ErrInvalidRoomID reports a malformed room id.
	ErrInvalidRoomID = errors.New("invalid furniture room id")

	// ErrInvalidItemID reports a malformed furniture item id.
	ErrInvalidItemID = errors.New("invalid furniture item id")

	// ErrInvalidDefinitionID reports a malformed furniture definition id.
	ErrInvalidDefinitionID = errors.New("invalid furniture definition id")

	// ErrInvalidQuantity reports a non-positive furniture grant quantity.
	ErrInvalidQuantity = errors.New("invalid furniture quantity")

	// ErrInvalidPlacement reports malformed floor placement input.
	ErrInvalidPlacement = errors.New("invalid furniture placement")

	// ErrItemNotFound reports a missing furniture item.
	ErrItemNotFound = errors.New("furniture item not found")

	// ErrDefinitionNotFound reports a missing furniture definition.
	ErrDefinitionNotFound = errors.New("furniture definition not found")

	// ErrNotItemOwner reports an actor that does not own the furniture item.
	ErrNotItemOwner = errors.New("actor does not own furniture item")

	// ErrItemNotInInventory reports an item that is not available to place.
	ErrItemNotInInventory = errors.New("furniture item not in inventory")

	// ErrItemNotPlaced reports an item that is not available to move or pick up.
	ErrItemNotPlaced = errors.New("furniture item not placed")

	// ErrItemNotInRoom reports an item outside the room authorized for a mutation.
	ErrItemNotInRoom = errors.New("furniture item not in authorized room")

	// ErrStateConflict reports a concurrent furniture state mutation.
	ErrStateConflict = errors.New("furniture state changed concurrently")

	// ErrGiftWriterUnavailable reports furniture persistence without wrapped grants.
	ErrGiftWriterUnavailable = errors.New("furniture gift writer unavailable")

	// ErrGiftOpenerUnavailable reports furniture persistence without gift opening support.
	ErrGiftOpenerUnavailable = errors.New("furniture gift opener unavailable")

	// ErrItemNotGift reports an open request for a non-gift item.
	ErrItemNotGift = errors.New("furniture item is not an unopened gift")

	// ErrItemStaged reports an item locked by an active direct trade.
	ErrItemStaged = errors.New("furniture item staged in trade")

	// ErrStackHeightUnavailable reports persistence without the optional exact-height capability.
	ErrStackHeightUnavailable = errors.New("furniture stack height persistence unavailable")
)

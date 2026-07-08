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

	// ErrInvalidPlacement reports malformed floor placement input.
	ErrInvalidPlacement = errors.New("invalid furniture placement")

	// ErrItemNotFound reports a missing furniture item.
	ErrItemNotFound = errors.New("furniture item not found")

	// ErrNotItemOwner reports an actor that does not own the furniture item.
	ErrNotItemOwner = errors.New("actor does not own furniture item")

	// ErrItemNotInInventory reports an item that is not available to place.
	ErrItemNotInInventory = errors.New("furniture item not in inventory")

	// ErrItemNotPlaced reports an item that is not available to move or pick up.
	ErrItemNotPlaced = errors.New("furniture item not placed")
)

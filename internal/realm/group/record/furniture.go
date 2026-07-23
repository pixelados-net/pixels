package record

// GroupFurnitureLink binds one furniture item to a social group without copying identity data.
type GroupFurnitureLink struct {
	// ItemID identifies the furniture instance.
	ItemID int64
	// RoomID identifies the current placed room when loaded.
	RoomID int64
	// GroupID identifies the linked active group.
	GroupID int64
}

// ReturnedFurniture identifies one headquarters item moved back to inventory.
type ReturnedFurniture struct {
	// ItemID identifies the returned furniture instance.
	ItemID int64
	// OwnerPlayerID identifies the inventory receiving the item.
	OwnerPlayerID int64
	// Wall reports whether the item uses the wall-furniture protocol.
	Wall bool
}

// FurnitureReturn describes one committed headquarters cleanup operation.
type FurnitureReturn struct {
	// RoomID identifies the headquarters whose active world must be updated.
	RoomID int64
	// Items stores the exact furniture instances returned to inventory.
	Items []ReturnedFurniture
}

// Count returns the number of furniture instances returned to inventory.
func (result FurnitureReturn) Count() int {
	return len(result.Items)
}

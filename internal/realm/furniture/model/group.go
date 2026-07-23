package model

// GroupData stores warmed social-group identity for one linked furniture item.
type GroupData struct {
	// GroupID identifies the linked social group.
	GroupID int64
	// BadgeCode stores its compiled badge.
	BadgeCode string
	// ColorAHex stores the primary resolved RGB value.
	ColorAHex string
	// ColorBHex stores the secondary resolved RGB value.
	ColorBHex string
}

// GroupPolicy resolves linked furniture identity without database access.
type GroupPolicy interface {
	// Furniture returns warmed group identity for one linked item.
	Furniture(int64) (GroupData, bool)
}

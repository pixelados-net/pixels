// Package recycled contains the committed recycler event.
package recycled

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies one committed recycle operation.
const Name bus.Name = "recycler.recycled"

// Payload stores bounded recycler identifiers.
type Payload struct {
	PlayerID          int64
	PrizeDefinitionID int64
	ItemCount         int32
}

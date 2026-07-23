// Package crafted contains the committed crafting event.
package crafted

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies one committed craft.
const Name bus.Name = "crafting.crafted"

// Payload stores bounded craft identifiers.
type Payload struct {
	PlayerID           int64
	RecipeID           int64
	RewardDefinitionID int64
}

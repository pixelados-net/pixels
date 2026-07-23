// Package discovered contains the secret recipe discovery event.
package discovered

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies one first durable discovery.
const Name bus.Name = "crafting.recipe.discovered"

// Payload stores bounded discovery identifiers.
type Payload struct {
	PlayerID int64
	RecipeID int64
}

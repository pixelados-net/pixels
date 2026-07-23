// Package exhausted contains the limited recipe exhaustion event.
package exhausted

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies one limited stock transition to zero.
const Name bus.Name = "crafting.recipe.exhausted"

// Payload stores the exhausted recipe identifier.
type Payload struct{ RecipeID int64 }

// Package live stores live furniture inventory viewer state for one player.
package live

import (
	"sync"
	"time"
)

// Holder stores active inventory UI state for one player.
type Holder struct {
	// mutex protects holder state.
	mutex sync.RWMutex

	// initializedAt stores when the holder opened.
	initializedAt time.Time
}

// NewHolder creates an inventory holder.
func NewHolder() *Holder {
	return &Holder{initializedAt: time.Now()}
}

// InitializedAt returns when the holder opened.
func (holder *Holder) InitializedAt() time.Time {
	holder.mutex.RLock()
	defer holder.mutex.RUnlock()

	return holder.initializedAt
}

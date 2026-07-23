package runtime

import (
	"context"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
)

// activateNode projects one executed trigger, extra, or effect box.
func (engine *Engine) activateNode(ctx context.Context, roomID int64, node *configuration.Node) {
	if engine.activator == nil || node == nil {
		return
	}
	_ = engine.activator.Activate(ctx, roomID, node.ItemID)
}

// activateNodes projects a stable group of executed extra boxes.
func (engine *Engine) activateNodes(ctx context.Context, roomID int64, nodes []*configuration.Node) {
	for _, node := range nodes {
		engine.activateNode(ctx, roomID, node)
	}
}

package place

import "context"

// projectionEffect applies one committed placement to live clients and room state.
type projectionEffect func(context.Context) error

// projectionQueue collects placement projections owned by an outer transaction.
type projectionQueue struct {
	// effects stores work in persistence order.
	effects []projectionEffect
}

// projectionKey isolates deferred placement work in a context.
type projectionKey struct{}

// DeferProjections returns a context that captures placement side effects and a completion callback.
func DeferProjections(ctx context.Context) (context.Context, func(context.Context) error) {
	queue := &projectionQueue{}
	complete := func(commitCtx context.Context) error {
		for _, effect := range queue.effects {
			if err := effect(commitCtx); err != nil {
				return err
			}
		}
		return nil
	}
	return context.WithValue(ctx, projectionKey{}, queue), complete
}

// project runs an ordinary placement immediately or appends it to the active transaction queue.
func project(ctx context.Context, effect projectionEffect) error {
	if queue, ok := ctx.Value(projectionKey{}).(*projectionQueue); ok {
		queue.effects = append(queue.effects, effect)
		return nil
	}
	return effect(ctx)
}

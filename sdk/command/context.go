package command

import (
	"context"

	"go.minekube.com/brigodier"
)

// senderKey prevents context collisions outside this SDK package.
type senderKey struct{}

// WithSender attaches one command sender to a context.
func WithSender(ctx context.Context, sender Sender) context.Context {
	return context.WithValue(ctx, senderKey{}, sender)
}

// SenderFrom returns the active command sender.
func SenderFrom(ctx context.Context) (Sender, bool) {
	sender, found := ctx.Value(senderKey{}).(Sender)
	return sender, found
}

// RequiresPermission creates a Brigadier node requirement backed by Sender.
func RequiresPermission(node string) brigodier.RequireFn {
	return func(ctx context.Context) bool {
		sender, found := SenderFrom(ctx)
		return found && sender.HasPermission(node)
	}
}

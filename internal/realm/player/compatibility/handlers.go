// Package compatibility registers intentionally retired user protocol handlers.
package compatibility

import (
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	openwelcome "github.com/niflaot/pixels/networking/inbound/furniture/welcome/open"
	changeemail "github.com/niflaot/pixels/networking/inbound/user/contact/change"
	statusemail "github.com/niflaot/pixels/networking/inbound/user/contact/status"
	welcomeemail "github.com/niflaot/pixels/networking/inbound/user/contact/welcome"
	nuxgifts "github.com/niflaot/pixels/networking/inbound/user/nux/gifts"
	nuxproceed "github.com/niflaot/pixels/networking/inbound/user/nux/proceed"
	classification "github.com/niflaot/pixels/networking/inbound/user/profile/classification"
)

// Register installs authenticated no-op handlers for retired user features.
func Register(registry *netconn.HandlerRegistry) {
	if registry == nil {
		return
	}
	register(registry, statusemail.Header, noop(statusemail.Decode))
	register(registry, changeemail.Header, noop(changeemail.Decode))
	register(registry, welcomeemail.Header, noop(welcomeemail.Decode))
	register(registry, nuxproceed.Header, noop(nuxproceed.Decode))
	register(registry, nuxgifts.Header, noop(nuxgifts.Decode))
	register(registry, openwelcome.Header, noop(openwelcome.Decode))
	register(registry, classification.Header, noop(classification.Decode))
}

// register installs one retired packet without widening its policy.
func register(registry *netconn.HandlerRegistry, header uint16, handler netconn.Handler) {
	_ = registry.Register(header, handler)
}

// noop creates a handler that validates wire and performs no behavior.
func noop[T any](decode func(codec.Packet) (T, error)) netconn.Handler {
	return func(_ netconn.Context, packet codec.Packet) error {
		_, err := decode(packet)
		return err
	}
}

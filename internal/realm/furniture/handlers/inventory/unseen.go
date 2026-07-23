package inventory

import (
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inunseencategory "github.com/niflaot/pixels/networking/inbound/inventory/unseen/category"
	inunseenitems "github.com/niflaot/pixels/networking/inbound/inventory/unseen/items"
)

// RegisterUnseen registers Nitro's transient inventory acknowledgement packets.
func RegisterUnseen(registry *netconn.HandlerRegistry) {
	_ = registry.Register(inunseencategory.Header, acknowledgeCategory)
	_ = registry.Register(inunseenitems.Header, acknowledgeItems)
}

// acknowledgeCategory validates one transient category acknowledgement.
func acknowledgeCategory(_ netconn.Context, packet codec.Packet) error {
	_, err := inunseencategory.Decode(packet)
	return err
}

// acknowledgeItems validates one transient item acknowledgement.
func acknowledgeItems(_ netconn.Context, packet codec.Packet) error {
	_, err := inunseenitems.Decode(packet)
	return err
}

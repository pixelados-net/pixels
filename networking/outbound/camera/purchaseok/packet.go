// Package purchaseok contains CAMERA_PURCHASE_OK.
package purchaseok

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CAMERA_PURCHASE_OK.
const Header uint16 = 2783

// Encode creates a purchase confirmation.
func Encode() (codec.Packet, error) { return codec.NewPacket(Header, nil) }

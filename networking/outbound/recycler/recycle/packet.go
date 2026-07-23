// Package recycle contains the RECYCLER_FINISHED outbound packet.
package recycle

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies RECYCLER_FINISHED.
	Header uint16 = 468
	// Complete identifies one committed recycle operation.
	Complete int32 = 1
	// Closed identifies a disabled recycler.
	Closed int32 = 2
)

// Encode creates one recycler completion packet.
func Encode(code int32, prizeDefinitionID int64) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(code), codec.Int32(int32(prizeDefinitionID)))
}

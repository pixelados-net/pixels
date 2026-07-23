// Package event encodes ROOM_EVENT projections.
package event

import "github.com/niflaot/pixels/networking/codec"

// Header identifies ROOM_EVENT.
const Header uint16 = 1840

// Data contains one active room promotion.
type Data struct {
	AdID                   int32
	OwnerAvatarID          int32
	OwnerAvatarName        string
	RoomID                 int32
	EventType              int32
	Name                   string
	Description            string
	MinutesSinceCreation   int32
	MinutesUntilExpiration int32
	CategoryID             int32
}

// Definition describes RoomEventData in renderer order.
var Definition = codec.Definition{codec.Named("adId", codec.Int32Field), codec.Named("ownerAvatarId", codec.Int32Field), codec.Named("ownerAvatarName", codec.StringField), codec.Named("roomId", codec.Int32Field), codec.Named("eventType", codec.Int32Field), codec.Named("name", codec.StringField), codec.Named("description", codec.StringField), codec.Named("minutesSinceCreation", codec.Int32Field), codec.Named("minutesUntilExpiration", codec.Int32Field), codec.Named("categoryId", codec.Int32Field)}

// Encode creates one active room event projection.
func Encode(data Data) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(data.AdID), codec.Int32(data.OwnerAvatarID), codec.String(data.OwnerAvatarName), codec.Int32(data.RoomID), codec.Int32(data.EventType), codec.String(data.Name), codec.String(data.Description), codec.Int32(data.MinutesSinceCreation), codec.Int32(data.MinutesUntilExpiration), codec.Int32(data.CategoryID))
}

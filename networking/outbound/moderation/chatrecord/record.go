// Package chatrecord encodes Nitro moderator chat records.
package chatrecord

import "github.com/niflaot/pixels/networking/codec"

// Line stores one visible moderation chat line.
type Line struct {
	Timestamp   string
	PlayerID    int32
	Username    string
	Message     string
	Highlighted bool
}

// Context stores one room context with integer id and string name.
type Context struct {
	RoomID   int32
	RoomName string
}

// Record stores one moderator chatlog record.
type Record struct {
	Type    uint8
	Context Context
	Lines   []Line
}

// Append appends one record to an existing payload.
func Append(payload []byte, record Record) ([]byte, error) {
	contextCount := int32(0)
	if record.Context.RoomID > 0 {
		contextCount = 2
	}
	next, err := codec.AppendPayload(payload, codec.Definition{codec.ByteField, codec.Uint16Field}, codec.Byte(record.Type), codec.Uint16(uint16(contextCount)))
	if err != nil {
		return payload, err
	}
	if contextCount > 0 {
		next, err = codec.AppendPayload(next, codec.Definition{codec.StringField, codec.ByteField, codec.Int32Field, codec.StringField, codec.ByteField, codec.StringField}, codec.String("roomId"), codec.Byte(1), codec.Int32(record.Context.RoomID), codec.String("roomName"), codec.Byte(2), codec.String(record.Context.RoomName))
		if err != nil {
			return payload, err
		}
	}
	next, err = codec.AppendPayload(next, codec.Definition{codec.Uint16Field}, codec.Uint16(uint16(len(record.Lines))))
	if err != nil {
		return payload, err
	}
	for _, line := range record.Lines {
		next, err = codec.AppendPayload(next, codec.Definition{codec.StringField, codec.Int32Field, codec.StringField, codec.StringField, codec.BooleanField}, codec.String(line.Timestamp), codec.Int32(line.PlayerID), codec.String(line.Username), codec.String(line.Message), codec.Bool(line.Highlighted))
		if err != nil {
			return payload, err
		}
	}
	return next, nil
}

// Package result encodes USER_IGNORED_RESULT changes.
package result

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies USER_IGNORED_RESULT.
	Header uint16 = 207
	// Failed reports an ignored-user mutation that was not applied.
	Failed int32 = 0
	// Ignored reports a newly ignored user.
	Ignored int32 = 1
	// Unignored reports a removed ignore.
	Unignored int32 = 3
)

// Encode creates one ignored-user mutation result.
func Encode(state int32, username string) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.StringField}, codec.Int32(state), codec.String(username))
}

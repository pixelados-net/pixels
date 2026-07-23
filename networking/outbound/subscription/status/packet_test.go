package status

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesSubscriptionState verifies every required status field.
func TestEncodeWritesSubscriptionState(t *testing.T) {
	packet, err := Encode(State{ProductName: "club_habbo", DaysToPeriodEnd: 31, ResponseType: 2, EverMember: true})
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v error=%v", packet, err)
	}
	values, rest, err := codec.DecodePayload(nil, Definition, packet.Payload)
	if err != nil || len(rest) != 0 || values[0].String != "club_habbo" || values[1].Int32 != 31 || values[4].Int32 != 2 || !values[5].Boolean {
		t.Fatalf("values=%#v rest=%d error=%v", values, len(rest), err)
	}
}

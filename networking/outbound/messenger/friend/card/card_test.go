package friendcard

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestAppendUsesNitroWireOrder verifies all friend-card fields.
func TestAppendUsesNitroWireOrder(t *testing.T) {
	payload, err := Append(nil, Card{PlayerID: 7, Username: "demo", Gender: 1, Online: true, FollowingAllowed: true, Look: "look", CategoryID: 2, Motto: "motto", Relationship: 3})
	if err != nil {
		t.Fatal(err)
	}
	values, rest, err := codec.DecodePayload(nil, definition, payload)
	if err != nil || len(rest) != 0 || values[0].Int32 != 7 || values[1].String != "demo" || values[13].Uint16 != 3 {
		t.Fatalf("unexpected values=%#v rest=%d err=%v", values, len(rest), err)
	}
}

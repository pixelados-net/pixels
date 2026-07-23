package requested

import (
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	"testing"
)

// TestEncodeWritesPendingMemberData verifies live request compatibility.
func TestEncodeWritesPendingMemberData(t *testing.T) {
	packet, err := Encode(7, grouprecord.Request{PlayerID: 8, Username: "alice"})
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}

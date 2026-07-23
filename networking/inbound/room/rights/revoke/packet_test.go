package revoke

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsPlayerIDs verifies bounded multi-player decoding.
func TestDecodeReadsPlayerIDs(t *testing.T) {
	payload, err := codec.AppendPayload(nil, CountDefinition, codec.Int32(2))
	if err == nil {
		payload, err = codec.AppendPayload(payload, PlayerDefinition, codec.Int32(7))
	}
	if err == nil {
		payload, err = codec.AppendPayload(payload, PlayerDefinition, codec.Int32(8))
	}
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := Decode(codec.Packet{Header: Header, Payload: payload})
	if err != nil || len(decoded.PlayerIDs) != 2 || decoded.PlayerIDs[1] != 8 {
		t.Fatalf("unexpected payload %#v err=%v", decoded, err)
	}
}

// TestDecodeRejectsOversizedCount verifies allocation bounds.
func TestDecodeRejectsOversizedCount(t *testing.T) {
	packet, err := codec.NewPacket(Header, CountDefinition, codec.Int32(maxPlayers+1))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Decode(packet); err == nil {
		t.Fatal("expected oversized count rejection")
	}
}

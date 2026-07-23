package answers

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeValidatesHeaderBoundsAndTrailingBytes verifies bounded quiz submissions.
func TestDecodeValidatesHeaderBoundsAndTrailingBytes(t *testing.T) {
	packet, err := codec.NewPacket(Header, codec.Definition{codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.String("SafetyQuiz"), codec.Int32(2), codec.Int32(3), codec.Int32(1))
	if err != nil {
		t.Fatalf("create packet: %v", err)
	}
	request, err := Decode(packet)
	if err != nil || request.Code != "SafetyQuiz" || len(request.AnswerIDs) != 2 || request.AnswerIDs[1] != 1 {
		t.Fatalf("unexpected request %#v error=%v", request, err)
	}
	packet.Header++
	if _, err = Decode(packet); err == nil {
		t.Fatal("expected header validation error")
	}
}

package session

import (
	"errors"
	"testing"

	tradecore "github.com/niflaot/pixels/internal/realm/trade/core"
	traderuntime "github.com/niflaot/pixels/internal/realm/trade/runtime"
	"github.com/niflaot/pixels/networking/codec"
	outopenfailed "github.com/niflaot/pixels/networking/outbound/trade/openfailed"
	outothernotallowed "github.com/niflaot/pixels/networking/outbound/trade/othernotallowed"
	outyounotallowed "github.com/niflaot/pixels/networking/outbound/trade/younotallowed"
)

// startFailureCase describes one domain-to-protocol failure mapping.
type startFailureCase struct {
	// name identifies the test case.
	name string
	// serviceErr stores the domain failure.
	serviceErr error
	// header stores the expected packet header.
	header uint16
	// reason stores the expected open-failure reason.
	reason int32
}

// TestOpenPacketUsesPlayerIDs verifies room-local unit ids never reach Nitro's web-id lookup.
func TestOpenPacketUsesPlayerIDs(t *testing.T) {
	session := &traderuntime.Session{
		First:  traderuntime.Participant{PlayerID: 101, UnitID: 1},
		Second: traderuntime.Participant{PlayerID: 202, UnitID: 2},
	}
	packet, err := openPacket(session)
	values, decodeErr := codec.DecodePacketExact(packet, codec.Definition{
		codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field,
	})
	if err != nil || decodeErr != nil {
		t.Fatalf("encode=%v decode=%v", err, decodeErr)
	}
	if values[0].Int32 != 101 || values[2].Int32 != 202 {
		t.Fatalf("first=%d second=%d", values[0].Int32, values[2].Int32)
	}
}

// TestStartFailurePacketUsesSpecificNitroReason verifies hotel and room failures remain distinct.
func TestStartFailurePacketUsesSpecificNitroReason(t *testing.T) {
	testCases := []startFailureCase{
		{name: "hotel disabled", serviceErr: tradecore.ErrDisabled, header: outopenfailed.Header, reason: 1},
		{name: "room disabled", serviceErr: tradecore.ErrRoomPolicy, header: outopenfailed.Header, reason: 6},
		{name: "actor throttled", serviceErr: tradecore.ErrThrottled, header: outopenfailed.Header, reason: 7},
		{name: "target unavailable", serviceErr: tradecore.ErrUnavailable, header: outopenfailed.Header, reason: 8},
		{name: "unknown failure", serviceErr: errors.New("unknown failure"), header: outopenfailed.Header, reason: 8},
		{name: "actor locked", serviceErr: tradecore.ErrActorNotAllowed, header: outyounotallowed.Header},
		{name: "target locked", serviceErr: tradecore.ErrTargetNotAllowed, header: outothernotallowed.Header},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			packet, err := startFailurePacket(testCase.serviceErr)
			if err != nil || packet.Header != testCase.header {
				t.Fatalf("packet=%#v err=%v", packet, err)
			}
			if packet.Header != outopenfailed.Header {
				if len(packet.Payload) != 0 {
					t.Fatalf("unexpected payload %v", packet.Payload)
				}
				return
			}
			values, decodeErr := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.StringField})
			if decodeErr != nil || values[0].Int32 != testCase.reason || values[1].String != "" {
				t.Fatalf("values=%#v err=%v", values, decodeErr)
			}
		})
	}
}

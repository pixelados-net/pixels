package kickback

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesMinutesUntilPayday verifies Nitro's countdown unit position.
func TestEncodeWritesMinutesUntilPayday(t *testing.T) {
	packet, err := Encode(Info{Streak: 7, FirstSubscriptionDate: "2026-07-01", Percentage: 0.1,
		CreditsSpent: 20, StreakBonus: 5, MonthlyBonus: 2, MinutesUntilPayday: 1440})
	definition := codec.Definition{codec.Int32Field, codec.StringField, codec.DoubleField, codec.Int32Field,
		codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}
	values, rest, decodeErr := codec.DecodePayload(nil, definition, packet.Payload)
	if err != nil || decodeErr != nil || len(rest) != 0 || values[8].Int32 != 1440 {
		t.Fatalf("values=%#v rest=%d encode=%v decode=%v", values, len(rest), err, decodeErr)
	}
}

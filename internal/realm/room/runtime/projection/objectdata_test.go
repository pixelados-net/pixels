package projection

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestSpecializedObjectDataEncodesMannequinMap verifies outfit data uses Nitro's map format.
func TestSpecializedObjectDataEncodesMannequinMap(t *testing.T) {
	data := SpecializedObjectData("mannequin", `{"gender":"M","figure":"ch-1-1","name":"Look"}`)
	payload, err := data.Append(nil)
	if err != nil {
		t.Fatalf("append mannequin data: %v", err)
	}
	values, rest, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field}, payload)
	if err != nil || values[0].Int32 != 1 || values[1].Int32 != 3 || len(rest) == 0 {
		t.Fatalf("unexpected map prefix %#v rest=%d err=%v", values, len(rest), err)
	}
}

// TestSpecializedObjectDataEncodesTonerArray verifies toner state uses Nitro's int-array format.
func TestSpecializedObjectDataEncodesTonerArray(t *testing.T) {
	data := SpecializedObjectData("background_toner", "1:20:30:40")
	payload, err := data.Append(nil)
	if err != nil {
		t.Fatalf("append toner data: %v", err)
	}
	values, _, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}, payload)
	if err != nil || values[0].Int32 != 5 || values[1].Int32 != 4 || values[5].Int32 != 40 {
		t.Fatalf("unexpected toner payload %#v err=%v", values, err)
	}
}

// TestSpecializedObjectDataRejectsMalformedState verifies legacy invalid values fall back safely.
func TestSpecializedObjectDataRejectsMalformedState(t *testing.T) {
	if SpecializedObjectData("mannequin", "invalid") != nil || SpecializedObjectData("background_toner", "1:2") != nil || SpecializedObjectData("background_toner", "2:20:30:40") != nil || SpecializedObjectData("background_toner", "1:256:30:40") != nil {
		t.Fatal("expected malformed specialized data to be rejected")
	}
}

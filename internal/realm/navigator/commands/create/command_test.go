package create

import (
	"testing"

	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
)

// TestCreateParamsMapsProtocolValues verifies create command mapping.
func TestCreateParamsMapsProtocolValues(t *testing.T) {
	params := createParams(Command{
		RoomName:        "Demo Room",
		RoomDescription: "hello",
		ModelName:       "model_a",
		CategoryID:      5,
		MaxVisitors:     25,
		TradeType:       int32(roommodel.TradeModeAllowed),
	}, 7, "demo")

	if params.OwnerPlayerID != 7 || params.OwnerName != "demo" || params.Name != "Demo Room" {
		t.Fatalf("unexpected owner params %#v", params)
	}
	if params.CategoryID == nil || *params.CategoryID != 5 {
		t.Fatalf("unexpected category %#v", params.CategoryID)
	}
	if params.TradeMode != roommodel.TradeModeAllowed {
		t.Fatalf("unexpected trade mode %d", params.TradeMode)
	}
}

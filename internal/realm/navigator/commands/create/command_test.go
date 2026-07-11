package create

import (
	"bytes"
	"strings"
	"testing"

	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

// TestCommandIdentityAndLogging verifies command routing and diagnostics.
func TestCommandIdentityAndLogging(t *testing.T) {
	input := Command{RoomName: "Room", ModelName: "model_a", CategoryID: 1, MaxVisitors: 25, TradeType: 2}
	if input.CommandName() != Name {
		t.Fatalf("name=%q", input.CommandName())
	}
	var output bytes.Buffer
	logger := zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), zapcore.AddSync(&output), zap.DebugLevel))
	logger.Debug("command", zap.Object("command", input))
	if !strings.Contains(output.String(), "model_a") {
		t.Fatalf("log=%s", output.String())
	}
}

// TestCategoryIDMapsOptionalValues verifies nullable categories.
func TestCategoryIDMapsOptionalValues(t *testing.T) {
	if categoryID(0) != nil || categoryID(-1) != nil {
		t.Fatal("expected nil category")
	}
}

package floorplan

import (
	"errors"
	"strings"
	"testing"
)

// TestValidateAcceptsNormalizedFloorplan verifies valid geometry and newline normalization.
func TestValidateAcceptsNormalizedFloorplan(t *testing.T) {
	validated, err := Validate(Config{RejectZeroEffectiveHeight: true}, SaveParams{
		Heightmap: "00\n0X", DoorX: 0, DoorY: 0, DoorDirection: 2,
		WallThickness: -1, FloorThickness: 1, WallHeight: -1,
	})
	if err != nil {
		t.Fatalf("validate floor plan: %v", err)
	}
	if validated.Params.Heightmap != "00\r0x" || validated.Grid.ValidCount() != 3 {
		t.Fatalf("unexpected validated floor plan %#v", validated)
	}
}

// TestValidateAggregatesIndependentFailures verifies one save reports every invalid setting.
func TestValidateAggregatesIndependentFailures(t *testing.T) {
	_, err := Validate(Config{RejectZeroEffectiveHeight: true}, SaveParams{
		Heightmap: "xx", DoorX: 4, DoorY: 4, DoorDirection: 9,
		WallThickness: 2, FloorThickness: -3, WallHeight: 16,
	})
	var validation ValidationErrors
	if !errors.As(err, &validation) {
		t.Fatalf("expected validation errors, got %v", err)
	}
	expected := []ErrorCode{CodeInvalidDirection, CodeInvalidWallThickness, CodeInvalidFloorThickness, CodeInvalidWallHeight, CodeZeroHeight, CodeInvalidDoor}
	for _, code := range expected {
		if !containsCode(validation.Codes, code) {
			t.Fatalf("missing code %q in %#v", code, validation.Codes)
		}
	}
}

// TestValidateRejectsDimensionLimits verifies editor limits stay outside the shared parser.
func TestValidateRejectsDimensionLimits(t *testing.T) {
	wide := strings.Repeat("0", MaxDimension+1)
	_, err := Validate(Config{}, SaveParams{Heightmap: wide, DoorDirection: 2, WallHeight: -1})
	var validation ValidationErrors
	if !errors.As(err, &validation) || !containsCode(validation.Codes, CodeTooLargeWidth) {
		t.Fatalf("expected width error, got %v", err)
	}
}

// containsCode reports whether a validation code is present.
func containsCode(codes []ErrorCode, expected ErrorCode) bool {
	for _, code := range codes {
		if code == expected {
			return true
		}
	}

	return false
}

// BenchmarkValidateSave measures maximum-size floor plan validation.
func BenchmarkValidateSave(b *testing.B) {
	row := strings.Repeat("0", MaxDimension)
	heightmap := strings.Repeat(row+"\r", MaxDimension-1) + row
	params := SaveParams{Heightmap: heightmap, DoorDirection: 2, WallHeight: -1}
	for b.Loop() {
		if _, err := Validate(Config{RejectZeroEffectiveHeight: true}, params); err != nil {
			b.Fatal(err)
		}
	}
}

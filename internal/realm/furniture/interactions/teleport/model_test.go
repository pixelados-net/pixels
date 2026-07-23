package teleport

import (
	"os"
	"testing"

	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// TestLoadConfig verifies teleport environment policy.
func TestLoadConfig(t *testing.T) {
	previous, present := os.LookupEnv("PIXELS_FURNITURE_TELEPORT_BYPASS_LOCKED")
	_ = os.Setenv("PIXELS_FURNITURE_TELEPORT_BYPASS_LOCKED", "true")
	t.Cleanup(func() {
		if present {
			_ = os.Setenv("PIXELS_FURNITURE_TELEPORT_BYPASS_LOCKED", previous)
		} else {
			_ = os.Unsetenv("PIXELS_FURNITURE_TELEPORT_BYPASS_LOCKED")
		}
	})
	config, err := LoadConfig()
	if err != nil || !config.BypassLocked {
		t.Fatalf("unexpected config %#v err=%v", config, err)
	}
}

// TestFrontPoint verifies cardinal teleport exit coordinates.
func TestFrontPoint(t *testing.T) {
	item := worldfurniture.Item{Point: grid.MustPoint(4, 4), Rotation: worldunit.RotationNorth}
	expected := []grid.Point{grid.MustPoint(4, 3), grid.MustPoint(5, 4), grid.MustPoint(4, 5), grid.MustPoint(3, 4)}
	rotations := []worldunit.Rotation{worldunit.RotationNorth, worldunit.RotationEast, worldunit.RotationSouth, worldunit.RotationWest}
	for index, rotation := range rotations {
		item.Rotation = rotation
		point, ok := frontPoint(item)
		if !ok || point != expected[index] {
			t.Fatalf("rotation %d point=%#v ok=%v", rotation, point, ok)
		}
	}
}

// TestDelayForDistinguishesTile verifies zero-delay tile transitions.
func TestDelayForDistinguishesTile(t *testing.T) {
	transit := Transit{Source: worldfurniture.Item{Definition: worldfurniture.Definition{InteractionType: "teleport_tile"}}}
	if delayFor(transit) != 0 {
		t.Fatal("expected tile transition without visual delay")
	}
	transit.Source.Definition.InteractionType = "teleport"
	if delayFor(transit) != phaseDelay {
		t.Fatal("expected pad visual delay")
	}
}

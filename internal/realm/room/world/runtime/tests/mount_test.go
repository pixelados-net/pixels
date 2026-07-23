package tests

import (
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldruntime "github.com/niflaot/pixels/internal/realm/room/world/runtime"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// TestMountedUnitsShareEveryMovementTick verifies riding never follows through delayed teleports.
func TestMountedUnitsShareEveryMovementTick(t *testing.T) {
	world := mountedWorld(t)
	plan, err := world.PlanMovement(7, grid.MustPoint(3, 0))
	if err != nil {
		t.Fatal(err)
	}
	roomPath, err := world.FindPath(plan)
	if err != nil {
		t.Fatal(err)
	}
	if err = world.ApplyMovement(7, roomPath, false); err != nil {
		t.Fatal(err)
	}
	movements := world.Tick()
	if len(movements) != 2 {
		t.Fatalf("movements=%+v", movements)
	}
	rider, riderFound := world.Unit(7)
	mount, mountFound := world.Unit(-50)
	if !riderFound || !mountFound || rider.Position != mount.Position || rider.BodyRotation != mount.BodyRotation {
		t.Fatalf("rider=%+v mount=%+v", rider, mount)
	}
	if rider.RenderOffset != worldunit.RidingHeightOffset || mount.RenderOffset != 0 {
		t.Fatalf("rider offset=%d mount offset=%d", rider.RenderOffset, mount.RenderOffset)
	}
}

// TestMountedTeleportAndDismountRemainAuthoritative verifies direct movement and cleanup affect both units.
func TestMountedTeleportAndDismountRemainAuthoritative(t *testing.T) {
	world := mountedWorld(t)
	point := grid.MustPoint(3, 0)
	if _, err := world.TeleportUnit(7, point, worldunit.RotationWest, false); err != nil {
		t.Fatal(err)
	}
	rider, _, err := world.SetMount(7, -50, false)
	if err != nil {
		t.Fatal(err)
	}
	mount, found := world.Unit(-50)
	if !found || rider.Position != mount.Position || rider.Position.Point != point {
		t.Fatalf("rider=%+v mount=%+v", rider, mount)
	}
	if rider.RenderOffset != 0 {
		t.Fatalf("rider offset=%d", rider.RenderOffset)
	}
}

// mountedWorld creates one flat world with a linked player and pet.
func mountedWorld(t testing.TB) *worldruntime.World {
	t.Helper()
	roomGrid, err := grid.Parse("0000", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	world, err := worldruntime.New(worldruntime.Config{
		Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)},
	})
	if err != nil {
		t.Fatal(err)
	}
	world.AddUnit(7)
	if _, err = world.AddEntity(-50, 7, worldunit.KindPet, worldpath.Position{Point: grid.MustPoint(1, 0)}, worldunit.RotationEast); err != nil {
		t.Fatal(err)
	}
	if _, _, err = world.SetMount(7, -50, true); err != nil {
		t.Fatal(err)
	}

	return world
}

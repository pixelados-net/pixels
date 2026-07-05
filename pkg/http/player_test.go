package http

import (
	"context"

	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// testFinder returns persistent test player records.
type testFinder struct{}

// FindByID finds a test player by id.
func (finder testFinder) FindByID(ctx context.Context, id int64) (playerservice.Record, bool, error) {
	if id != 2 {
		return playerservice.Record{}, false, nil
	}

	return testRecord(id), true, nil
}

// FindByUsername finds a test player by username.
func (finder testFinder) FindByUsername(context.Context, string) (playerservice.Record, bool, error) {
	return testRecord(2), true, nil
}

// testRecord returns a persistent test player record.
func testRecord(id int64) playerservice.Record {
	return playerservice.Record{
		Player: playermodel.Player{
			Base:     sharedmodel.Base{Identity: sharedmodel.Identity{ID: id}},
			Username: "test_player",
		},
		Profile: playermodel.Profile{
			PlayerID:        id,
			Look:            "hd-180-1",
			Gender:          playermodel.GenderMale,
			Motto:           "Test fixture.",
			AllowNameChange: true,
		},
	}
}

package mysterybox

import (
	"context"
	"encoding/json"
	"testing"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
)

// stateUpdaterForTest captures a trophy state mutation.
type stateUpdaterForTest struct {
	// params stores the last state mutation.
	params furnitureservice.StateParams
}

// UpdateState captures and returns the requested next state.
func (updater *stateUpdaterForTest) UpdateState(_ context.Context, params furnitureservice.StateParams) (furnituremodel.Item, error) {
	updater.params = params
	return furnituremodel.Item{ExtraData: params.Next}, nil
}

// keyStoreForTest returns deterministic tracker colors.
type keyStoreForTest struct{}

// FindKeys returns deterministic tracker colors.
func (keyStoreForTest) FindKeys(context.Context, int64) (Keys, error) {
	return Keys{BoxColor: "blue", KeyColor: "gold"}, nil
}

// TestInscribePersistsStructuredData verifies trophy text is not stored as raw protocol data.
func TestInscribePersistsStructuredData(t *testing.T) {
	states := &stateUpdaterForTest{}
	service := New(Config{}, keyStoreForTest{}, nil, states, nil)
	encoded, err := service.Inscribe(context.Background(), 7, 8, 9, "old", "hola")
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]any
	if err = json.Unmarshal([]byte(encoded), &value); err != nil {
		t.Fatal(err)
	}
	if value["text"] != "hola" || states.params.Expected != "old" || states.params.Next != encoded {
		t.Fatalf("value=%#v params=%#v", value, states.params)
	}
}

// TestKeysDelegatesToStore verifies reconnect bootstrap uses durable account state.
func TestKeysDelegatesToStore(t *testing.T) {
	service := New(Config{}, keyStoreForTest{}, nil, &stateUpdaterForTest{}, nil)
	keys, err := service.Keys(context.Background(), 7)
	if err != nil || keys.BoxColor != "blue" || keys.KeyColor != "gold" {
		t.Fatalf("keys=%#v err=%v", keys, err)
	}
}

// BenchmarkRoomTaskKey measures allocation-free reveal key derivation.
func BenchmarkRoomTaskKey(b *testing.B) {
	for range b.N {
		_ = roomTaskKey(7)
	}
}

package decoration

import (
	"context"
	"errors"
	"testing"
)

// TestApplySurfaceValidatesInputBeforePersistence verifies surface safety guards.
func TestApplySurfaceValidatesInputBeforePersistence(t *testing.T) {
	store := &fakeStore{changed: true}
	service := New(store)
	tests := []struct {
		name    string
		surface Surface
		value   string
		err     error
	}{
		{name: "floor", surface: SurfaceFloor, value: "101"},
		{name: "landscape", surface: SurfaceLandscape, value: "3.1"},
		{name: "unknown", surface: Surface("ceiling"), value: "1", err: ErrInvalidSurface},
		{name: "unsafe", surface: SurfaceWallpaper, value: "1;drop", err: ErrInvalidSurfaceValue},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := service.ApplySurface(context.Background(), 1, 2, 3, test.surface, test.value)
			if !errors.Is(err, test.err) {
				t.Fatalf("expected %v, got %v", test.err, err)
			}
		})
	}
}

// TestPlacePostItValidatesModernCoordinates verifies malformed wall input never reaches persistence.
func TestPlacePostItValidatesModernCoordinates(t *testing.T) {
	store := &fakeStore{changed: true}
	service := New(store)
	if err := service.PlacePostIt(context.Background(), 1, 2, 3, ":w=2,3 l=4,5 r"); err != nil {
		t.Fatalf("place valid post-it: %v", err)
	}
	if store.postItData != DefaultPostItData {
		t.Fatalf("expected renderable initial data %q, got %q", DefaultPostItData, store.postItData)
	}
	if err := service.PlacePostIt(context.Background(), 1, 2, 3, "outside"); !errors.Is(err, ErrInvalidWallPosition) {
		t.Fatalf("expected wall position error, got %v", err)
	}
}

// TestSaveDimmerNormalizesAndValidatesPreset verifies Nitro's bounded preset contract.
func TestSaveDimmerNormalizesAndValidatesPreset(t *testing.T) {
	store := &fakeStore{changed: true}
	service := New(store)
	state, err := service.SaveDimmer(context.Background(), 1, 2, Preset{ID: 1, Color: "#74f5f5", Brightness: 180}, true)
	if err != nil || state.Presets[0].Color != "#74F5F5" {
		t.Fatalf("unexpected normalized state %#v err=%v", state, err)
	}
	_, err = service.SaveDimmer(context.Background(), 1, 2, Preset{ID: 4, Color: "#74F5F5", Brightness: 180}, true)
	if !errors.Is(err, ErrInvalidDimmerPreset) {
		t.Fatalf("expected invalid preset, got %v", err)
	}
}

// TestServicePropagatesPersistenceAndGuardResults verifies durable errors and stale mutations remain distinct.
func TestServicePropagatesPersistenceAndGuardResults(t *testing.T) {
	expected := errors.New("store unavailable")
	for _, test := range []struct {
		name string
		call func(*Service) error
	}{
		{name: "surface", call: func(service *Service) error {
			return service.ApplySurface(context.Background(), 1, 2, 3, SurfaceFloor, "101")
		}},
		{name: "post-it", call: func(service *Service) error {
			return service.PlacePostIt(context.Background(), 1, 2, 3, ":w=2,3 l=4,5 r")
		}},
		{name: "save dimmer", call: func(service *Service) error {
			_, err := service.SaveDimmer(context.Background(), 1, 2, Preset{ID: 1, Color: "#000000", Brightness: 255}, false)
			return err
		}},
		{name: "toggle dimmer", call: func(service *Service) error {
			_, err := service.ToggleDimmer(context.Background(), 1, 2)
			return err
		}},
	} {
		t.Run(test.name+" unavailable", func(t *testing.T) {
			if err := test.call(New(&fakeStore{})); !errors.Is(err, ErrDecorationUnavailable) {
				t.Fatalf("expected unavailable error, got %v", err)
			}
		})
		t.Run(test.name+" store error", func(t *testing.T) {
			if err := test.call(New(&fakeStore{err: expected})); !errors.Is(err, expected) {
				t.Fatalf("expected store error, got %v", err)
			}
		})
	}
}

// TestLoadAndToggleDimmerReturnDurableState verifies read and toggle state projection.
func TestLoadAndToggleDimmerReturnDurableState(t *testing.T) {
	expected := DimmerState{ItemID: 7, ExtraData: "2,1,1,#000000,255"}
	service := New(&fakeStore{changed: true, found: true, state: expected})
	loaded, found, err := service.LoadDimmer(context.Background(), 9)
	if err != nil || !found || loaded.ItemID != expected.ItemID || loaded.ExtraData != expected.ExtraData {
		t.Fatalf("unexpected load state=%#v found=%t err=%v", loaded, found, err)
	}
	toggled, err := service.ToggleDimmer(context.Background(), 9, 1)
	if err != nil || toggled.ItemID != expected.ItemID || toggled.ExtraData != expected.ExtraData {
		t.Fatalf("unexpected toggle state=%#v err=%v", toggled, err)
	}
}

// fakeStore stores deterministic decoration test results.
type fakeStore struct {
	// changed controls guarded mutations.
	changed bool
	// found controls dimmer reads.
	found bool
	// state stores the projected dimmer state.
	state DimmerState
	// err stores the persistence failure.
	err error
	// postItData stores the initial durable note state.
	postItData string
}

// ConsumeSurface returns the configured guarded result.
func (store *fakeStore) ConsumeSurface(context.Context, int64, int64, int64, Surface, string) (bool, error) {
	return store.changed, store.err
}

// PlacePostIt returns the configured guarded result.
func (store *fakeStore) PlacePostIt(_ context.Context, _ int64, _ int64, _ int64, _ string, data string) (bool, error) {
	store.postItData = data
	return store.changed, store.err
}

// LoadDimmer returns one deterministic preset.
func (store *fakeStore) LoadDimmer(context.Context, int64) (DimmerState, bool, error) {
	return store.state, store.found, store.err
}

// SaveDimmer returns the normalized supplied preset.
func (store *fakeStore) SaveDimmer(_ context.Context, _ int64, _ int64, preset Preset, _ bool) (DimmerState, bool, error) {
	if store.state.ItemID != 0 || store.err != nil {
		return store.state, store.changed, store.err
	}
	return DimmerState{Presets: []Preset{preset}}, store.changed, nil
}

// ToggleDimmer returns a deterministic mutation result.
func (store *fakeStore) ToggleDimmer(context.Context, int64, int64) (DimmerState, bool, error) {
	return store.state, store.changed, store.err
}

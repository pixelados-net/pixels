package live

import (
	"sync"

	catalogviewer "github.com/niflaot/pixels/internal/realm/catalog/viewer/live"
	inventoryviewer "github.com/niflaot/pixels/internal/realm/furniture/viewer/live"
	currencyholder "github.com/niflaot/pixels/internal/realm/inventory/currency/holder"
	navviewer "github.com/niflaot/pixels/internal/realm/navigator/viewer/live"
)

// Player is the live runtime player composition root.
type Player struct {
	// mutex protects runtime snapshot replacement.
	mutex sync.RWMutex

	// snapshot stores durable player data copied into runtime state.
	snapshot Snapshot

	// peer stores the authenticated connection binding.
	peer SessionPeer

	// navigator stores navigator UI state when opened.
	navigator *navviewer.Viewer

	// catalog stores catalog UI state when opened.
	catalog *catalogviewer.Viewer

	// inventory stores furniture inventory viewer state when opened.
	inventory *inventoryviewer.Holder

	// currencies stores the player's currency capability.
	currencies *currencyholder.Holder

	// ignored stores player ids hidden by this player.
	ignored map[int64]struct{}

	// profileViewer stores the currently opened public profile.
	profileViewer ProfileViewer

	// room stores the player's current room presence.
	room RoomPresence
}

// NewPlayer creates a live player.
func NewPlayer(snapshot Snapshot, peer SessionPeer) (*Player, error) {
	if !snapshot.Valid() {
		return nil, ErrInvalidPlayer
	}
	if !peer.Valid() {
		return nil, ErrInvalidPeer
	}

	return &Player{
		snapshot:   snapshot,
		peer:       peer,
		currencies: currencyholder.New(snapshot.ID),
		ignored:    make(map[int64]struct{}),
	}, nil
}

// ID returns the player id.
func (player *Player) ID() int64 {
	player.mutex.RLock()
	defer player.mutex.RUnlock()

	return player.snapshot.ID
}

// Username returns the player username.
func (player *Player) Username() string {
	player.mutex.RLock()
	defer player.mutex.RUnlock()

	return player.snapshot.Username
}

// Snapshot returns a copy of the runtime player snapshot.
func (player *Player) Snapshot() Snapshot {
	player.mutex.RLock()
	defer player.mutex.RUnlock()

	return player.snapshot
}

// ReplaceSnapshot replaces durable runtime data.
func (player *Player) ReplaceSnapshot(snapshot Snapshot) error {
	if !snapshot.Valid() || snapshot.ID != player.ID() {
		return ErrInvalidPlayer
	}

	player.mutex.Lock()
	defer player.mutex.Unlock()

	player.snapshot = snapshot

	return nil
}

// Peer returns the player session peer.
func (player *Player) Peer() SessionPeer {
	player.mutex.RLock()
	defer player.mutex.RUnlock()

	return player.peer
}

// OpenNavigator creates or returns the player's navigator viewer.
func (player *Player) OpenNavigator() *navviewer.Viewer {
	player.mutex.Lock()
	defer player.mutex.Unlock()

	if player.navigator == nil {
		player.navigator = navviewer.NewViewer()
	}

	return player.navigator
}

// Navigator returns the player's navigator viewer.
func (player *Player) Navigator() (*navviewer.Viewer, bool) {
	player.mutex.RLock()
	defer player.mutex.RUnlock()

	if player.navigator == nil {
		return nil, false
	}

	return player.navigator, true
}

// CloseNavigator removes the player's navigator viewer.
func (player *Player) CloseNavigator() (*navviewer.Viewer, bool) {
	player.mutex.Lock()
	defer player.mutex.Unlock()

	if player.navigator == nil {
		return nil, false
	}

	viewer := player.navigator
	player.navigator = nil

	return viewer, true
}

// OpenInventory creates or returns the player's inventory viewer holder.
func (player *Player) OpenInventory() *inventoryviewer.Holder {
	player.mutex.Lock()
	defer player.mutex.Unlock()

	if player.inventory == nil {
		player.inventory = inventoryviewer.NewHolder()
	}

	return player.inventory
}

// Inventory returns the player's inventory viewer holder.
func (player *Player) Inventory() (*inventoryviewer.Holder, bool) {
	player.mutex.RLock()
	defer player.mutex.RUnlock()

	if player.inventory == nil {
		return nil, false
	}

	return player.inventory, true
}

// CloseInventory removes the player's inventory viewer holder.
func (player *Player) CloseInventory() (*inventoryviewer.Holder, bool) {
	player.mutex.Lock()
	defer player.mutex.Unlock()

	if player.inventory == nil {
		return nil, false
	}

	holder := player.inventory
	player.inventory = nil

	return holder, true
}

// Currencies returns the player's composed currency capability.
func (player *Player) Currencies() *currencyholder.Holder {
	return player.currencies
}

// EnterRoom stores the player's current room id.
func (player *Player) EnterRoom(roomID int64) error {
	if roomID <= 0 {
		return ErrInvalidRoomPresence
	}

	player.mutex.Lock()
	defer player.mutex.Unlock()

	player.room.currentID = roomID

	return nil
}

// CurrentRoom returns the player's current room id.
func (player *Player) CurrentRoom() (int64, bool) {
	player.mutex.RLock()
	defer player.mutex.RUnlock()

	return player.room.Current()
}

// LeaveRoom clears the player's current room id.
func (player *Player) LeaveRoom() (int64, bool) {
	player.mutex.Lock()
	defer player.mutex.Unlock()

	roomID, found := player.room.Current()
	player.room.currentID = 0

	return roomID, found
}

// RoomPresence stores live room presence for one player.
type RoomPresence struct {
	// currentID identifies the room currently joined by the player.
	currentID int64
}

// Current returns the current room id.
func (presence RoomPresence) Current() (int64, bool) {
	if presence.currentID <= 0 {
		return 0, false
	}

	return presence.currentID, true
}

package unit

import "github.com/niflaot/pixels/internal/realm/room/world/path"

// Kind stores the room unit kind.
type Kind uint8

const (
	// KindPlayer marks a player-controlled unit.
	KindPlayer Kind = iota + 1

	// KindBot marks a bot-controlled unit.
	KindBot

	// KindPet marks a pet-controlled unit.
	KindPet
)

// Unit stores runtime state for an entity inside a room.
type Unit struct {
	// id stores the room-local unit id.
	id int64

	// ownerID stores the optional durable owner id.
	ownerID int64

	// kind stores the unit kind.
	kind Kind

	// position stores the current unit position.
	position path.Position

	// previous stores the previous unit position.
	previous path.Position

	// body stores the body rotation.
	body Rotation

	// head stores the head rotation.
	head Rotation

	// goal stores the current movement goal.
	goal path.Position

	// hasGoal reports whether goal is active.
	hasGoal bool

	// steps stores pending path steps.
	steps []path.Step

	// statuses stores current unit statuses.
	statuses statuses
}

// New creates a room unit.
func New(params Params) (*Unit, error) {
	if params.ID <= 0 || params.Kind == 0 {
		return nil, ErrInvalidUnit
	}

	return &Unit{
		id:       params.ID,
		ownerID:  params.OwnerID,
		kind:     params.Kind,
		position: params.Position,
		previous: params.Position,
		body:     params.Body,
		head:     params.Head,
	}, nil
}

// Params stores unit creation input.
type Params struct {
	// ID stores the room-local unit id.
	ID int64

	// OwnerID stores the optional durable owner id.
	OwnerID int64

	// Kind stores the unit kind.
	Kind Kind

	// Position stores the initial position.
	Position path.Position

	// Body stores the initial body rotation.
	Body Rotation

	// Head stores the initial head rotation.
	Head Rotation
}

// ID returns the room-local unit id.
func (unit *Unit) ID() int64 {
	return unit.id
}

// OwnerID returns the optional durable owner id.
func (unit *Unit) OwnerID() int64 {
	return unit.ownerID
}

// Kind returns the unit kind.
func (unit *Unit) Kind() Kind {
	return unit.kind
}

// Position returns the current position.
func (unit *Unit) Position() path.Position {
	return unit.position
}

// Previous returns the previous position.
func (unit *Unit) Previous() path.Position {
	return unit.previous
}

// BodyRotation returns the body rotation.
func (unit *Unit) BodyRotation() Rotation {
	return unit.body
}

// HeadRotation returns the head rotation.
func (unit *Unit) HeadRotation() Rotation {
	return unit.head
}

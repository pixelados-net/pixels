package look

import "errors"

var (
	// ErrPlayerNotInRoom reports a look command without active room state.
	ErrPlayerNotInRoom = errors.New("player not in room")

	// ErrInvalidTarget reports a target that cannot be represented as a grid point.
	ErrInvalidTarget = errors.New("invalid look target")
)

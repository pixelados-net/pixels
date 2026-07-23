package record

import "errors"

var (
	// ErrFavoriteUnavailable reports an invalid, duplicate, or over-limit favorite.
	ErrFavoriteUnavailable = errors.New("navigator favorite unavailable")
)

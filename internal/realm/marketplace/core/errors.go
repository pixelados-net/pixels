package core

import "errors"

var (
	// ErrDisabled reports a globally disabled Marketplace.
	ErrDisabled = errors.New("marketplace disabled")
	// ErrInvalidPrice reports a price outside configured bounds.
	ErrInvalidPrice = errors.New("invalid marketplace price")
	// ErrNoToken reports a seller without listing tokens.
	ErrNoToken = errors.New("marketplace token required")
	// ErrListingUnavailable reports a missing, expired, or completed listing.
	ErrListingUnavailable = errors.New("marketplace listing unavailable")
	// ErrOwnListing reports a buyer attempting to purchase their own listing.
	ErrOwnListing = errors.New("cannot buy own marketplace listing")
)

package service

import "errors"

var (
	// ErrInvalidPlayerID reports an invalid catalog buyer id.
	ErrInvalidPlayerID = errors.New("invalid catalog player id")

	// ErrInvalidOfferID reports an invalid catalog offer id.
	ErrInvalidOfferID = errors.New("invalid catalog offer id")

	// ErrPageNotFound reports a missing catalog page.
	ErrPageNotFound = errors.New("catalog page not found")

	// ErrOfferNotFound reports a missing catalog offer.
	ErrOfferNotFound = errors.New("catalog offer not found")

	// ErrOfferNotVisible reports a purchase outside page or offer access policy.
	ErrOfferNotVisible = errors.New("catalog offer not visible")

	// ErrOfferDisabled reports a disabled catalog offer.
	ErrOfferDisabled = errors.New("catalog offer disabled")

	// ErrLimitedSoldOut reports an exhausted LTD series.
	ErrLimitedSoldOut = errors.New("catalog limited offer sold out")

	// ErrLimitedCompletion reports an inconsistent LTD completion.
	ErrLimitedCompletion = errors.New("catalog limited purchase could not complete")
)

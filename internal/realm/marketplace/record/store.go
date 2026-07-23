package record

import (
	"context"
	"time"
)

// Store persists Marketplace state and owns its transaction boundary.
type Store interface {
	// WithinTransaction runs work atomically.
	WithinTransaction(context.Context, func(context.Context) error) error
	// TokenBalance reads a player's listing tokens.
	TokenBalance(context.Context, int64) (int32, error)
	// AddTokens atomically adds a non-negative token package.
	AddTokens(context.Context, int64, int32) (int32, error)
	// SpendToken atomically spends one token.
	SpendToken(context.Context, int64) (bool, error)
	// CreateListing inserts an open listing.
	CreateListing(context.Context, Listing) (Listing, error)
	// FindListingForUpdate locks and reads one listing.
	FindListingForUpdate(context.Context, int64) (Listing, bool, error)
	// FindCheapestListing returns the cheapest current replacement for a definition.
	FindCheapestListing(context.Context, int64) (Listing, bool, error)
	// MarkSold conditionally completes one open listing.
	MarkSold(context.Context, int64, int64) (bool, error)
	// CloseListing conditionally closes one open listing.
	CloseListing(context.Context, int64, int64, bool) (Listing, bool, error)
	// SearchOffers aggregates open unexpired listings by furniture and LTD serial.
	SearchOffers(context.Context, Search) ([]SearchOffer, int32, error)
	// ListOwnListings lists one seller's visible listings and pending proceeds.
	ListOwnListings(context.Context, int64, time.Time) ([]Listing, error)
	// RedeemSold marks all unredeemed sold listings and returns their raw total.
	RedeemSold(context.Context, int64) (int64, int32, error)
	// ExpireListings closes expired listings and returns them.
	ExpireListings(context.Context, int32) ([]Listing, error)
	// DefinitionStats returns recent daily history and current open count.
	DefinitionStats(context.Context, int64, int32) ([]DayStat, int32, error)
}

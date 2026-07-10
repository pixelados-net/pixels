package admin

import "errors"

var (
	// ErrInvalidPage reports invalid catalog page input.
	ErrInvalidPage = errors.New("invalid catalog page")
	// ErrPageNotFound reports a missing catalog page.
	ErrPageNotFound = errors.New("catalog page not found")
	// ErrInvalidItem reports invalid catalog offer input.
	ErrInvalidItem = errors.New("invalid catalog item")
	// ErrItemNotFound reports a missing catalog offer.
	ErrItemNotFound = errors.New("catalog item not found")
	// ErrDefinitionNotFound reports a missing furniture definition.
	ErrDefinitionNotFound = errors.New("catalog furniture definition not found")
	// ErrConflict reports an optimistic locking conflict.
	ErrConflict = errors.New("catalog record changed concurrently")
	// ErrLimitedBelowSales reports an LTD stack smaller than committed sales.
	ErrLimitedBelowSales = errors.New("catalog limited stack is below committed sales")
)

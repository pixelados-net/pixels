package service

import "errors"

var (
	// ErrCommerceUnavailable reports a store without extended catalog persistence.
	ErrCommerceUnavailable = errors.New("catalog commerce persistence unavailable")

	// ErrInvalidAmount reports an invalid bulk purchase amount.
	ErrInvalidAmount = errors.New("invalid catalog purchase amount")
	// ErrOfferNotGiftable reports a gift request for a regular offer.
	ErrOfferNotGiftable = errors.New("catalog offer is not giftable")
	// ErrGiftReceiverNotFound reports an unknown gift recipient.
	ErrGiftReceiverNotFound = errors.New("gift receiver not found")
	// ErrVoucherInvalid reports an unknown, disabled, or expired voucher.
	ErrVoucherInvalid = errors.New("catalog voucher is invalid")
	// ErrVoucherAlreadyUsed reports a repeated player redemption.
	ErrVoucherAlreadyUsed = errors.New("catalog voucher already used")
	// ErrVoucherExhausted reports an exhausted global voucher cap.
	ErrVoucherExhausted = errors.New("catalog voucher exhausted")
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

	// ErrTeleportPairing reports an invalid or unavailable teleport pair grant.
	ErrTeleportPairing = errors.New("catalog teleport offer could not create a pair")
	// ErrGroupSelection reports an invalid social-group product selection.
	ErrGroupSelection = errors.New("catalog social-group selection is invalid")
)

package codec

import "errors"

var (
	// ErrFrameTooSmall reports a frame shorter than the packet header.
	ErrFrameTooSmall = errors.New("frame too small")

	// ErrInvalidField reports a payload field type mismatch.
	ErrInvalidField = errors.New("invalid field")

	// ErrPayloadTooLarge reports payload data that exceeds uint32 frame length.
	ErrPayloadTooLarge = errors.New("payload too large")

	// ErrStringTooLarge reports a string that exceeds uint16 byte length.
	ErrStringTooLarge = errors.New("string too large")

	// ErrTruncatedPayload reports a payload that ends before all fields decode.
	ErrTruncatedPayload = errors.New("truncated payload")

	// ErrUnexpectedPayload reports extra payload bytes after decoding all fields.
	ErrUnexpectedPayload = errors.New("unexpected payload")

	// ErrUnexpectedHeader reports a packet decoded with the wrong packet definition.
	ErrUnexpectedHeader = errors.New("unexpected header")
)

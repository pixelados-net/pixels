package codec

import (
	"encoding/binary"
	"math"
)

// AppendPayload appends encoded payload values to dst using definition order.
func AppendPayload(dst []byte, definition Definition, values ...Value) ([]byte, error) {
	if len(values) > len(definition) {
		return dst, ErrInvalidField
	}

	for index, field := range definition {
		if index >= len(values) {
			if field.Optional {
				continue
			}

			return dst, ErrInvalidField
		}

		value := values[index]
		var err error
		dst, err = appendValue(dst, field, value)
		if err != nil {
			return dst, err
		}
	}

	return dst, nil
}

// DecodePayload decodes payload values from src using definition order.
func DecodePayload(dst []Value, definition Definition, src []byte) ([]Value, []byte, error) {
	for _, field := range definition {
		value, rest, err := decodeValue(field, src)
		if err != nil {
			if field.Optional && err == ErrTruncatedPayload {
				return dst, src, nil
			}

			return dst, src, err
		}

		dst = append(dst, value)
		src = rest
	}

	return dst, src, nil
}

// appendValue appends one encoded value to dst.
func appendValue(dst []byte, field Field, value Value) ([]byte, error) {
	switch field.Kind {
	case BooleanKind:
		if value.Boolean {
			return append(dst, 1), nil
		}
		return append(dst, 0), nil
	case Int32Kind:
		return binary.BigEndian.AppendUint32(dst, uint32(value.Int32)), nil
	case Uint16Kind:
		return binary.BigEndian.AppendUint16(dst, value.Uint16), nil
	case Uint32Kind:
		return binary.BigEndian.AppendUint32(dst, value.Uint32), nil
	case StringKind:
		return appendString(dst, value.String)
	case ByteKind:
		return append(dst, value.Byte), nil
	case DoubleKind:
		return binary.BigEndian.AppendUint64(dst, math.Float64bits(value.Double)), nil
	default:
		return dst, ErrInvalidField
	}
}

// decodeValue decodes one value from src.
func decodeValue(field Field, src []byte) (Value, []byte, error) {
	switch field.Kind {
	case BooleanKind:
		if len(src) < 1 {
			return Value{}, src, ErrTruncatedPayload
		}
		return Bool(src[0] != 0), src[1:], nil
	case Int32Kind:
		if len(src) < 4 {
			return Value{}, src, ErrTruncatedPayload
		}
		return Int32(int32(binary.BigEndian.Uint32(src[:4]))), src[4:], nil
	case Uint16Kind:
		if len(src) < 2 {
			return Value{}, src, ErrTruncatedPayload
		}
		return Uint16(binary.BigEndian.Uint16(src[:2])), src[2:], nil
	case Uint32Kind:
		if len(src) < 4 {
			return Value{}, src, ErrTruncatedPayload
		}
		return Uint32(binary.BigEndian.Uint32(src[:4])), src[4:], nil
	case StringKind:
		return decodeString(src)
	case ByteKind:
		if len(src) < 1 {
			return Value{}, src, ErrTruncatedPayload
		}
		return Byte(src[0]), src[1:], nil
	case DoubleKind:
		if len(src) < 8 {
			return Value{}, src, ErrTruncatedPayload
		}
		return Float64(math.Float64frombits(binary.BigEndian.Uint64(src[:8]))), src[8:], nil
	default:
		return Value{}, src, ErrInvalidField
	}
}

// appendString appends a protocol string to dst.
func appendString(dst []byte, value string) ([]byte, error) {
	if len(value) > int(^uint16(0)) {
		return dst, ErrStringTooLarge
	}

	dst = binary.BigEndian.AppendUint16(dst, uint16(len(value)))
	dst = append(dst, value...)

	return dst, nil
}

// decodeString decodes a protocol string from src.
func decodeString(src []byte) (Value, []byte, error) {
	if len(src) < 2 {
		return Value{}, src, ErrTruncatedPayload
	}

	size := int(binary.BigEndian.Uint16(src[:2]))
	if len(src) < 2+size {
		return Value{}, src, ErrTruncatedPayload
	}

	return String(string(src[2 : 2+size])), src[2+size:], nil
}

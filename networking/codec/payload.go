package codec

import "encoding/binary"

// AppendPayload appends encoded payload values to dst using definition order.
func AppendPayload(dst []byte, definition Definition, values ...Value) ([]byte, error) {
	if len(definition) != len(values) {
		return dst, ErrInvalidField
	}

	for index, field := range definition {
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
			return dst, src, err
		}

		dst = append(dst, value)
		src = rest
	}

	return dst, src, nil
}

// appendValue appends one encoded value to dst.
func appendValue(dst []byte, field Field, value Value) ([]byte, error) {
	switch field {
	case BooleanField:
		if value.Boolean {
			return append(dst, 1), nil
		}
		return append(dst, 0), nil
	case Int32Field:
		return binary.BigEndian.AppendUint32(dst, uint32(value.Int32)), nil
	case Uint16Field:
		return binary.BigEndian.AppendUint16(dst, value.Uint16), nil
	case Uint32Field:
		return binary.BigEndian.AppendUint32(dst, value.Uint32), nil
	case StringField:
		return appendString(dst, value.String)
	default:
		return dst, ErrInvalidField
	}
}

// decodeValue decodes one value from src.
func decodeValue(field Field, src []byte) (Value, []byte, error) {
	switch field {
	case BooleanField:
		if len(src) < 1 {
			return Value{}, src, ErrTruncatedPayload
		}
		return Bool(src[0] != 0), src[1:], nil
	case Int32Field:
		if len(src) < 4 {
			return Value{}, src, ErrTruncatedPayload
		}
		return Int32(int32(binary.BigEndian.Uint32(src[:4]))), src[4:], nil
	case Uint16Field:
		if len(src) < 2 {
			return Value{}, src, ErrTruncatedPayload
		}
		return Uint16(binary.BigEndian.Uint16(src[:2])), src[2:], nil
	case Uint32Field:
		if len(src) < 4 {
			return Value{}, src, ErrTruncatedPayload
		}
		return Uint32(binary.BigEndian.Uint32(src[:4])), src[4:], nil
	case StringField:
		return decodeString(src)
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

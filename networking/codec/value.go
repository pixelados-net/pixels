package codec

// Field describes one payload field in declaration order.
type Field uint8

const (
	// BooleanField encodes a boolean as uint8.
	BooleanField Field = iota + 1

	// Int32Field encodes a signed 32-bit integer.
	Int32Field

	// Uint16Field encodes an unsigned 16-bit integer.
	Uint16Field

	// Uint32Field encodes an unsigned 32-bit integer.
	Uint32Field

	// StringField encodes UTF-8 text with a uint16 byte length prefix.
	StringField
)

// Definition describes a packet payload in wire order.
type Definition []Field

// Value contains one decoded or encodable payload value.
type Value struct {
	Boolean bool
	Int32   int32
	Uint16  uint16
	Uint32  uint32
	String  string
}

// Bool returns a boolean payload value.
func Bool(value bool) Value {
	return Value{Boolean: value}
}

// Int32 returns an int32 payload value.
func Int32(value int32) Value {
	return Value{Int32: value}
}

// Uint16 returns a uint16 payload value.
func Uint16(value uint16) Value {
	return Value{Uint16: value}
}

// Uint32 returns a uint32 payload value.
func Uint32(value uint32) Value {
	return Value{Uint32: value}
}

// String returns a string payload value.
func String(value string) Value {
	return Value{String: value}
}

package codec

// Kind names a primitive payload field type.
type Kind uint8

const (
	// BooleanKind names a boolean encoded as uint8.
	BooleanKind Kind = iota + 1

	// Int32Kind names a signed 32-bit integer.
	Int32Kind

	// Uint16Kind names an unsigned 16-bit integer.
	Uint16Kind

	// Uint32Kind names an unsigned 32-bit integer.
	Uint32Kind

	// StringKind names UTF-8 text with a uint16 byte length prefix.
	StringKind
)

var (
	// BooleanField encodes a required boolean as uint8.
	BooleanField = Field{Kind: BooleanKind}

	// Int32Field encodes a required signed 32-bit integer.
	Int32Field = Field{Kind: Int32Kind}

	// Uint16Field encodes a required unsigned 16-bit integer.
	Uint16Field = Field{Kind: Uint16Kind}

	// Uint32Field encodes a required unsigned 32-bit integer.
	Uint32Field = Field{Kind: Uint32Kind}

	// StringField encodes a required UTF-8 string with a uint16 byte length prefix.
	StringField = Field{Kind: StringKind}
)

// Field describes one payload field in declaration order.
type Field struct {
	Name     string
	Kind     Kind
	Optional bool
}

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

// Optional returns an optional field declaration.
func Optional(field Field) Field {
	field.Optional = true

	return field
}

// Named returns a field declaration with a protocol field name.
func Named(name string, field Field) Field {
	field.Name = name

	return field
}

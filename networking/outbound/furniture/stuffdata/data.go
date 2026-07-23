// Package stuffdata encodes Nitro furniture object data fragments.
package stuffdata

import "github.com/niflaot/pixels/networking/codec"

const (
	// legacyFormat identifies Nitro legacy string furniture data.
	legacyFormat int32 = 0
	// mapFormat identifies Nitro map-style furniture object data.
	mapFormat int32 = 1
	// stringArrayFormat identifies Nitro string-array furniture object data.
	stringArrayFormat int32 = 2

	// intArrayFormat identifies Nitro integer-array furniture object data.
	intArrayFormat int32 = 5
)

// Pair stores one object-data key and value.
type Pair struct {
	// Key stores the object-data key.
	Key string
	// Value stores the object-data value.
	Value string
}

// Data stores one specialized furniture object-data representation.
type Data struct {
	// Legacy stores legacy string object data.
	Legacy *string
	// Pairs stores map object data.
	Pairs []Pair
	// Ints stores integer-array object data.
	Ints []int32
	// Strings stores string-array object data.
	Strings []string
	// Highscore stores highscore board object data.
	Highscore *Highscore
}

// Legacy creates legacy string object data.
func Legacy(value string) *Data { return &Data{Legacy: &value} }

// Map creates map-style object data.
func Map(pairs []Pair) *Data { return &Data{Pairs: pairs} }

// IntArray creates integer-array object data.
func IntArray(values []int32) *Data { return &Data{Ints: values} }

// StringArray creates string-array object data.
func StringArray(values []string) *Data { return &Data{Strings: values} }

// Board creates highscore board object data.
func Board(value Highscore) *Data { return &Data{Highscore: &value} }

// Append appends the selected specialized representation.
func (data *Data) Append(dst []byte) ([]byte, error) {
	if data == nil {
		return dst, nil
	}
	if data.Pairs != nil {
		return AppendMap(dst, data.Pairs)
	}
	if data.Legacy != nil {
		return codec.AppendPayload(dst, codec.Definition{codec.Int32Field, codec.StringField}, codec.Int32(legacyFormat), codec.String(*data.Legacy))
	}
	if data.Highscore != nil {
		return AppendHighscore(dst, *data.Highscore)
	}
	if data.Strings != nil {
		return AppendStringArray(dst, data.Strings)
	}

	return AppendIntArray(dst, data.Ints)
}

// AppendStringArray appends string-array furniture object data.
func AppendStringArray(dst []byte, values []string) ([]byte, error) {
	payload, err := codec.AppendPayload(dst, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(stringArrayFormat), codec.Int32(int32(len(values))))
	if err != nil {
		return dst, err
	}
	for _, value := range values {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField}, codec.String(value))
		if err != nil {
			return dst, err
		}
	}
	return payload, nil
}

// AppendIntArray appends integer-array furniture object data.
func AppendIntArray(dst []byte, values []int32) ([]byte, error) {
	payload, err := codec.AppendPayload(dst, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(intArrayFormat), codec.Int32(int32(len(values))))
	if err != nil {
		return dst, err
	}
	for _, value := range values {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(value))
		if err != nil {
			return dst, err
		}
	}

	return payload, nil
}

// AppendMap appends map-style furniture object data.
func AppendMap(dst []byte, pairs []Pair) ([]byte, error) {
	payload, err := codec.AppendPayload(dst, codec.Definition{codec.Int32Field, codec.Int32Field},
		codec.Int32(mapFormat), codec.Int32(int32(len(pairs))))
	if err != nil {
		return dst, err
	}
	for _, pair := range pairs {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField, codec.StringField},
			codec.String(pair.Key), codec.String(pair.Value))
		if err != nil {
			return dst, err
		}
	}

	return payload, nil
}

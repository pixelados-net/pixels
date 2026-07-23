// Package common encodes shared WIRED definition fragments.
package common

import "github.com/niflaot/pixels/networking/codec"

// AppendTriggerable appends Nitro's shared WIRED definition fields.
func AppendTriggerable(dst []byte, selectionEnabled bool, selectionLimit int32, selected []int64, spriteID int32, itemID int64, stringParam string, intParams []int32, selectionMode int32) ([]byte, error) {
	payload, err := codec.AppendPayload(dst, codec.Definition{codec.BooleanField, codec.Int32Field, codec.Int32Field},
		codec.Bool(selectionEnabled), codec.Int32(selectionLimit), codec.Int32(int32(len(selected))))
	if err != nil {
		return dst, err
	}
	for _, selectedID := range selected {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(int32(selectedID)))
		if err != nil {
			return dst, err
		}
	}
	payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field},
		codec.Int32(spriteID), codec.Int32(int32(itemID)), codec.String(stringParam), codec.Int32(int32(len(intParams))))
	if err != nil {
		return dst, err
	}
	for _, param := range intParams {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(param))
		if err != nil {
			return dst, err
		}
	}
	return codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(selectionMode))
}

// AppendInt32s appends one count-prefixed integer vector.
func AppendInt32s(dst []byte, values []int32) ([]byte, error) {
	payload, err := codec.AppendPayload(dst, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(values))))
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

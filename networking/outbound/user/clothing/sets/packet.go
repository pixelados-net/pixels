// Package sets contains the USER_CLOTHING outbound packet.
package sets

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_CLOTHING.
const Header uint16 = 1450

// Definition describes the figure-set count.
var Definition = codec.Definition{codec.Named("figureSetCount", codec.Int32Field)}

// FigureSetDefinition describes one unlocked figure set.
var FigureSetDefinition = codec.Definition{codec.Named("figureSetId", codec.Int32Field)}

// ProductDefinition describes one bound furniture product code.
var ProductDefinition = codec.Definition{codec.Named("productCode", codec.StringField)}

// Encode creates USER_CLOTHING in the exact two-list Renderer order.
func Encode(figureSetIDs []int32, productCodes []string) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, Definition, codec.Int32(int32(len(figureSetIDs))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, figureSetID := range figureSetIDs {
		payload, err = codec.AppendPayload(payload, FigureSetDefinition, codec.Int32(figureSetID))
		if err != nil {
			return codec.Packet{}, err
		}
	}
	payload, err = codec.AppendPayload(payload, Definition, codec.Int32(int32(len(productCodes))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, productCode := range productCodes {
		payload, err = codec.AppendPayload(payload, ProductDefinition, codec.String(productCode))
		if err != nil {
			return codec.Packet{}, err
		}
	}
	return codec.Packet{Header: Header, Payload: payload}, nil
}

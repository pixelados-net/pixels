// Package metadata contains the NAVIGATOR_METADATA outbound packet.
package metadata

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the NAVIGATOR_METADATA packet identifier.
	Header uint16 = 3052
)

// Context contains one navigator top-level context.
type Context struct {
	// Code identifies the top-level context.
	Code string
	// SavedSearches stores nested saved searches.
	SavedSearches []SavedSearch
}

// SavedSearch contains one metadata saved search entry.
type SavedSearch struct {
	// ID identifies the saved search.
	ID int32
	// Code stores the search code.
	Code string
	// Filter stores the saved filter.
	Filter string
	// Localization stores the display localization.
	Localization string
}

// Definition describes the NAVIGATOR_METADATA payload fields.
var Definition = codec.Definition{codec.Named("contextCount", codec.Int32Field)}

// ContextDefinition describes one top-level context entry.
var ContextDefinition = codec.Definition{
	codec.Named("code", codec.StringField),
	codec.Named("savedSearchCount", codec.Int32Field),
}

// SavedSearchDefinition describes one metadata saved search entry.
var SavedSearchDefinition = codec.Definition{
	codec.Named("id", codec.Int32Field),
	codec.Named("code", codec.StringField),
	codec.Named("filter", codec.StringField),
	codec.Named("localization", codec.StringField),
}

// Encode creates a NAVIGATOR_METADATA packet.
func Encode(contexts []Context) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, Definition, codec.Int32(int32(len(contexts))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, context := range contexts {
		payload, err = appendContext(payload, context)
		if err != nil {
			return codec.Packet{}, err
		}
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}

// appendContext appends one metadata context.
func appendContext(dst []byte, context Context) ([]byte, error) {
	dst, err := codec.AppendPayload(dst, ContextDefinition, codec.String(context.Code), codec.Int32(int32(len(context.SavedSearches))))
	if err != nil {
		return dst, err
	}
	for _, saved := range context.SavedSearches {
		dst, err = codec.AppendPayload(dst, SavedSearchDefinition,
			codec.Int32(saved.ID),
			codec.String(saved.Code),
			codec.String(saved.Filter),
			codec.String(saved.Localization),
		)
		if err != nil {
			return dst, err
		}
	}

	return dst, nil
}

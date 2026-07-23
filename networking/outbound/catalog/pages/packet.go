// Package pages contains the CATALOG_INDEX outbound packet.
package pages

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the CATALOG_INDEX packet identifier.
	Header uint16 = 1032
)

// Node describes one recursive catalog page tree entry.
type Node struct {
	// Visible reports whether the node is visible.
	Visible bool
	// IconImage stores the client page icon image.
	IconImage int32
	// PageID identifies the catalog page.
	PageID int32
	// Name stores the stable page slug.
	Name string
	// Localization stores the localized page title.
	Localization string
	// OfferIDs stores offer ids exposed by the page node.
	OfferIDs []int32
	// Children stores nested catalog pages.
	Children []Node
}

// Encode creates a CATALOG_INDEX packet with Nitro's synthetic root node.
func Encode(nodes []Node, mode string, additions ...bool) (codec.Packet, error) {
	payload, err := appendNode(nil, Node{Visible: true, PageID: -1, Name: "root", Children: nodes})
	if err != nil {
		return codec.Packet{}, err
	}
	newAdditions := len(additions) > 0 && additions[0]
	payload, err = codec.AppendPayload(payload, codec.Definition{codec.BooleanField, codec.StringField}, codec.Bool(newAdditions), codec.String(mode))
	if err != nil {
		return codec.Packet{}, err
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}

// appendNode appends one recursive catalog page node.
func appendNode(dst []byte, node Node) ([]byte, error) {
	dst, err := codec.AppendPayload(dst, codec.Definition{
		codec.BooleanField, codec.Int32Field, codec.Int32Field, codec.StringField,
		codec.StringField, codec.Int32Field,
	}, codec.Bool(node.Visible), codec.Int32(node.IconImage), codec.Int32(node.PageID),
		codec.String(node.Name), codec.String(node.Localization), codec.Int32(int32(len(node.OfferIDs))))
	if err != nil {
		return dst, err
	}
	for _, offerID := range node.OfferIDs {
		dst, err = codec.AppendPayload(dst, codec.Definition{codec.Int32Field}, codec.Int32(offerID))
		if err != nil {
			return dst, err
		}
	}
	dst, err = codec.AppendPayload(dst, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(node.Children))))
	if err != nil {
		return dst, err
	}
	for _, child := range node.Children {
		dst, err = appendNode(dst, child)
		if err != nil {
			return dst, err
		}
	}

	return dst, nil
}

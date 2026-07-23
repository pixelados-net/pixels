package list

import "github.com/niflaot/pixels/networking/codec"

// headerDefinition returns the fragment header field order.
func headerDefinition() codec.Definition {
	return codec.Definition{
		codec.Named("totalFragments", codec.Int32Field),
		codec.Named("fragmentNumber", codec.Int32Field),
		codec.Named("itemCount", codec.Int32Field),
	}
}

// itemPrefixDefinition returns the inventory item prefix field order.
func itemPrefixDefinition() codec.Definition {
	return codec.Definition{
		codec.Named("giftAdjustedId", codec.Int32Field),
		codec.Named("typeCode", codec.StringField),
		codec.Named("id", codec.Int32Field),
		codec.Named("spriteId", codec.Int32Field),
		codec.Named("kind", codec.Int32Field),
	}
}

// regularDataDefinition returns regular furniture data fields.
func regularDataDefinition() codec.Definition {
	return codec.Definition{codec.Named("limitedFlag", codec.Int32Field), codec.Named("extradata", codec.StringField)}
}

// limitedDataDefinition returns LTD furniture data fields.
func limitedDataDefinition() codec.Definition {
	return codec.Definition{codec.Named("limitedFlag", codec.Int32Field), codec.Named("limitedNumber", codec.Int32Field), codec.Named("limitedTotal", codec.Int32Field)}
}

// itemSuffixDefinition returns inventory permission and rental fields.
func itemSuffixDefinition() codec.Definition {
	return codec.Definition{
		codec.Named("allowRecycle", codec.BooleanField),
		codec.Named("allowTrade", codec.BooleanField),
		codec.Named("allowInventoryStack", codec.BooleanField),
		codec.Named("allowMarketplace", codec.BooleanField),
		codec.Named("unknown1", codec.Int32Field),
		codec.Named("hasRentPeriodStarted", codec.BooleanField),
		codec.Named("unknown2", codec.Int32Field),
	}
}

// itemDefinition returns the complete regular item shape used by packet tests.
func itemDefinition() codec.Definition {
	definition := append(codec.Definition{}, itemPrefixDefinition()...)
	definition = append(definition, regularDataDefinition()...)
	return append(definition, itemSuffixDefinition()...)
}

// floorDefinition returns fields present only on floor inventory items.
func floorDefinition() codec.Definition {
	return codec.Definition{
		codec.Named("songId", codec.StringField),
		codec.Named("floorTrailerKind", codec.Int32Field),
	}
}

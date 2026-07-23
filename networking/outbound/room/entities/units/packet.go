// Package units contains the UNIT outbound packet.
package units

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the UNIT packet identifier.
	Header uint16 = 374

	// AvatarType is the Nitro avatar unit type.
	AvatarType int32 = 1

	// RentableBotType is the Nitro owner-configurable bot unit type.
	RentableBotType int32 = 4

	// PetType is Nitro's pet room-unit type.
	PetType int32 = 2
)

// Unit stores one avatar unit record.
type Unit struct {
	// Type stores the Nitro unit type and defaults to AvatarType.
	Type int32
	// UserID stores the durable player id.
	UserID int64

	// Name stores the visible player name.
	Name string

	// Motto stores the player motto.
	Motto string

	// Figure stores the Nitro figure string.
	Figure string

	// RoomIndex stores the room-local unit id.
	RoomIndex int64

	// X stores the unit tile x coordinate.
	X int32

	// Y stores the unit tile y coordinate.
	Y int32

	// Z stores the unit vertical height.
	Z string

	// Direction stores the body direction.
	Direction int32

	// Gender stores the avatar gender code.
	Gender string

	// GroupID stores the active group id.
	GroupID int32

	// GroupStatus stores the active group status.
	GroupStatus int32

	// GroupName stores the active group name.
	GroupName string

	// SwimFigure stores the optional swim figure.
	SwimFigure string

	// AchievementScore stores the visible achievement score.
	AchievementScore int32

	// Moderator reports whether the user has moderator badge state.
	Moderator bool

	// OwnerID stores the owner of a rentable bot.
	OwnerID int64

	// OwnerName stores the owner name of a rentable bot.
	OwnerName string

	// Skills stores the available rentable bot command identifiers.
	Skills []uint16

	// PetSpecies stores the renderer pet type identifier.
	PetSpecies int32
	// PetRarity stores the pet rarity category.
	PetRarity int32
	// HasSaddle reports whether a rideable pet has a saddle.
	HasSaddle bool
	// IsRiding reports whether a rider is mounted on the pet.
	IsRiding bool
	// CanBreed reports whether the pet may breed now.
	CanBreed bool
	// CanHarvest reports whether a monsterplant may be harvested now.
	CanHarvest bool
	// CanRevive reports whether a monsterplant may be revived now.
	CanRevive bool
	// HasBreedingPermission reports whether the viewer may breed the pet.
	HasBreedingPermission bool
	// PetLevel stores the visible pet level.
	PetLevel int32
	// Posture stores the current renderer posture.
	Posture string
}

// Encode creates a UNIT packet.
func Encode(records []Unit) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(records))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, record := range records {
		payload, err = appendUnit(payload, record)
		if err != nil {
			return codec.Packet{}, err
		}
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}

// appendUnit appends one avatar unit.
func appendUnit(dst []byte, record Unit) ([]byte, error) {
	unitType := record.Type
	if unitType == 0 {
		unitType = AvatarType
	}
	payload, err := codec.AppendPayload(dst, baseDefinition(),
		codec.Int32(int32(record.UserID)),
		codec.String(record.Name),
		codec.String(record.Motto),
		codec.String(record.Figure),
		codec.Int32(int32(record.RoomIndex)),
		codec.Int32(record.X),
		codec.Int32(record.Y),
		codec.String(record.Z),
		codec.Int32(record.Direction),
		codec.Int32(unitType),
	)
	if err != nil {
		return nil, err
	}
	if unitType == RentableBotType {
		payload, err = codec.AppendPayload(payload, botDefinition(),
			codec.String(record.Gender), codec.Int32(int32(record.OwnerID)), codec.String(record.OwnerName), codec.Int32(int32(len(record.Skills))),
		)
		if err != nil {
			return nil, err
		}
		for _, skill := range record.Skills {
			payload, err = codec.AppendPayload(payload, codec.Definition{codec.Uint16Field}, codec.Uint16(skill))
			if err != nil {
				return nil, err
			}
		}
		return payload, nil
	}
	if unitType == PetType {
		return codec.AppendPayload(payload, petDefinition(),
			codec.Int32(record.PetSpecies), codec.Int32(int32(record.OwnerID)), codec.String(record.OwnerName),
			codec.Int32(record.PetRarity), codec.Bool(record.HasSaddle), codec.Bool(record.IsRiding),
			codec.Bool(record.CanBreed), codec.Bool(record.CanHarvest), codec.Bool(record.CanRevive),
			codec.Bool(record.HasBreedingPermission), codec.Int32(record.PetLevel), codec.String(record.Posture),
		)
	}
	return codec.AppendPayload(payload, avatarDefinition(),
		codec.String(record.Gender),
		codec.Int32(record.GroupID),
		codec.Int32(record.GroupStatus),
		codec.String(record.GroupName),
		codec.String(record.SwimFigure),
		codec.Int32(record.AchievementScore),
		codec.Bool(record.Moderator),
	)
}

// baseDefinition returns fields shared by every room unit.
func baseDefinition() codec.Definition {
	return codec.Definition{
		codec.Named("userId", codec.Int32Field),
		codec.Named("name", codec.StringField),
		codec.Named("custom", codec.StringField),
		codec.Named("figure", codec.StringField),
		codec.Named("roomIndex", codec.Int32Field),
		codec.Named("x", codec.Int32Field),
		codec.Named("y", codec.Int32Field),
		codec.Named("z", codec.StringField),
		codec.Named("direction", codec.Int32Field),
		codec.Named("type", codec.Int32Field),
	}
}

// avatarDefinition returns avatar-only room unit fields.
func avatarDefinition() codec.Definition {
	return codec.Definition{
		codec.Named("sex", codec.StringField),
		codec.Named("groupId", codec.Int32Field),
		codec.Named("groupStatus", codec.Int32Field),
		codec.Named("groupName", codec.StringField),
		codec.Named("swimFigure", codec.StringField),
		codec.Named("activityPoints", codec.Int32Field),
		codec.Named("isModerator", codec.BooleanField),
	}
}

// botDefinition returns rentable-bot fields before its variable skill list.
func botDefinition() codec.Definition {
	return codec.Definition{
		codec.Named("sex", codec.StringField),
		codec.Named("ownerId", codec.Int32Field),
		codec.Named("ownerName", codec.StringField),
		codec.Named("skillCount", codec.Int32Field),
	}
}

// petDefinition returns Nitro's exact pet-only room-unit fields.
func petDefinition() codec.Definition {
	return codec.Definition{
		codec.Named("petType", codec.Int32Field), codec.Named("ownerId", codec.Int32Field),
		codec.Named("ownerName", codec.StringField), codec.Named("rarity", codec.Int32Field),
		codec.Named("hasSaddle", codec.BooleanField), codec.Named("isRiding", codec.BooleanField),
		codec.Named("canBreed", codec.BooleanField), codec.Named("canHarvest", codec.BooleanField),
		codec.Named("canRevive", codec.BooleanField), codec.Named("hasBreedingPermission", codec.BooleanField),
		codec.Named("level", codec.Int32Field), codec.Named("posture", codec.StringField),
	}
}

// Package info encodes PET_INFO.
package info

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PET_INFO.
const Header uint16 = 2901

// Info stores Nitro's complete pet information view.
type Info struct {
	// ID identifies the pet.
	ID int64
	// Name stores the visible pet name.
	Name string
	// Level stores the current level.
	Level int32
	// MaximumLevel stores the species level cap.
	MaximumLevel int32
	// Experience stores total experience.
	Experience int32
	// LevelExperienceGoal stores the next threshold.
	LevelExperienceGoal int32
	// Energy stores current energy.
	Energy int32
	// MaximumEnergy stores maximum energy.
	MaximumEnergy int32
	// Happiness stores current happiness.
	Happiness int32
	// MaximumHappiness stores maximum happiness.
	MaximumHappiness int32
	// Respect stores accumulated respect.
	Respect int32
	// OwnerID identifies the owner.
	OwnerID int64
	// AgeDays stores full days since creation.
	AgeDays int32
	// OwnerName stores the owner display name.
	OwnerName string
	// Rarity stores the rarity category.
	Rarity int32
	// Saddle reports equipped saddle state.
	Saddle bool
	// Rider reports current rider state.
	Rider bool
	// SkillThresholds stores command unlock levels.
	SkillThresholds []int32
	// PubliclyRideable stores Nitro's numeric public riding flag.
	PubliclyRideable int32
	// Breedable reports current breeding eligibility.
	Breedable bool
	// FullyGrown reports monsterplant maturity.
	FullyGrown bool
	// Dead reports monsterplant death.
	Dead bool
	// UnknownRarity stores Nitro's secondary rarity value.
	UnknownRarity int32
	// MaximumTimeToLive stores monsterplant lifetime seconds.
	MaximumTimeToLive int32
	// RemainingTimeToLive stores remaining lifetime seconds.
	RemainingTimeToLive int32
	// RemainingGrowTime stores remaining grow seconds.
	RemainingGrowTime int32
	// PubliclyBreedable reports public breeding permission.
	PubliclyBreedable bool
}

// Encode creates PET_INFO.
func Encode(value Info) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, baseDefinition(),
		codec.Int32(int32(value.ID)), codec.String(value.Name), codec.Int32(value.Level), codec.Int32(value.MaximumLevel),
		codec.Int32(value.Experience), codec.Int32(value.LevelExperienceGoal), codec.Int32(value.Energy), codec.Int32(value.MaximumEnergy),
		codec.Int32(value.Happiness), codec.Int32(value.MaximumHappiness), codec.Int32(value.Respect), codec.Int32(int32(value.OwnerID)),
		codec.Int32(value.AgeDays), codec.String(value.OwnerName), codec.Int32(value.Rarity), codec.Bool(value.Saddle), codec.Bool(value.Rider),
		codec.Int32(int32(len(value.SkillThresholds))))
	if err != nil {
		return codec.Packet{}, err
	}
	for _, threshold := range value.SkillThresholds {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(threshold))
		if err != nil {
			return codec.Packet{}, err
		}
	}
	payload, err = codec.AppendPayload(payload, tailDefinition(), codec.Int32(value.PubliclyRideable), codec.Bool(value.Breedable),
		codec.Bool(value.FullyGrown), codec.Bool(value.Dead), codec.Int32(value.UnknownRarity), codec.Int32(value.MaximumTimeToLive),
		codec.Int32(value.RemainingTimeToLive), codec.Int32(value.RemainingGrowTime), codec.Bool(value.PubliclyBreedable))
	return codec.Packet{Header: Header, Payload: payload}, err
}

// baseDefinition returns fields preceding variable skill thresholds.
func baseDefinition() codec.Definition {
	return codec.Definition{
		codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field,
		codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field,
		codec.Int32Field, codec.StringField, codec.Int32Field, codec.BooleanField, codec.BooleanField, codec.Int32Field,
	}
}

// tailDefinition returns fields following variable skill thresholds.
func tailDefinition() codec.Definition {
	return codec.Definition{codec.Int32Field, codec.BooleanField, codec.BooleanField, codec.BooleanField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.BooleanField}
}

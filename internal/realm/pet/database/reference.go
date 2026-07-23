package database

import (
	"context"
	"time"

	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
)

// Species lists all species reference rows.
func (repository *Repository) Species(ctx context.Context) ([]petrecord.Species, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select type_id,slug,display_key,behavior_kind,max_level,rideable,breedable,plant,enabled from pet_species order by type_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]petrecord.Species, 0, 36)
	for rows.Next() {
		item := petrecord.Species{}
		if err = rows.Scan(&item.TypeID, &item.Slug, &item.DisplayKey, &item.BehaviorKind, &item.MaximumLevel, &item.Rideable, &item.Breedable, &item.Plant, &item.Enabled); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// Breeds lists all breed reference rows.
func (repository *Repository) Breeds(ctx context.Context) ([]petrecord.Breed, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select type_id,breed_id,palette_id,color,sellable,rarity,enabled from pet_breeds order by type_id,breed_id,palette_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]petrecord.Breed, 0)
	for rows.Next() {
		item := petrecord.Breed{}
		if err = rows.Scan(&item.TypeID, &item.BreedID, &item.PaletteID, &item.Color, &item.Sellable, &item.Rarity, &item.Enabled); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// Commands lists every command and species association.
func (repository *Repository) Commands(ctx context.Context) ([]petrecord.Command, map[int32][]int32, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select id,name_key,required_level,family,energy_cost,happiness_cost,experience_reward,duration_ms,cooldown_ms,enabled from pet_commands order by id`)
	if err != nil {
		return nil, nil, err
	}
	commands := make([]petrecord.Command, 0, 46)
	for rows.Next() {
		item := petrecord.Command{}
		var duration, cooldown int64
		if err = rows.Scan(&item.ID, &item.NameKey, &item.RequiredLevel, &item.Family, &item.EnergyCost, &item.HappinessCost, &item.ExperienceReward, &duration, &cooldown, &item.Enabled); err != nil {
			rows.Close()
			return nil, nil, err
		}
		item.Duration, item.Cooldown = time.Duration(duration)*time.Millisecond, time.Duration(cooldown)*time.Millisecond
		commands = append(commands, item)
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return nil, nil, err
	}
	rows.Close()
	links, err := repository.executorFor(ctx).Query(ctx, `select type_id,command_id from pet_species_commands order by type_id,command_id`)
	if err != nil {
		return nil, nil, err
	}
	defer links.Close()
	bySpecies := make(map[int32][]int32, 36)
	for links.Next() {
		var typeID, commandID int32
		if err = links.Scan(&typeID, &commandID); err != nil {
			return nil, nil, err
		}
		bySpecies[typeID] = append(bySpecies[typeID], commandID)
	}
	return commands, bySpecies, links.Err()
}

// ProductRules lists typed pet product rules.
func (repository *Repository) ProductRules(ctx context.Context) ([]petrecord.ProductRule, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select furniture_definition_id,kind,type_id,energy_delta,happiness_delta,experience_delta,consumable,enabled from pet_product_rules order by furniture_definition_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]petrecord.ProductRule, 0)
	for rows.Next() {
		item := petrecord.ProductRule{}
		if err = rows.Scan(&item.DefinitionID, &item.Kind, &item.TypeID, &item.EnergyDelta, &item.HappinessDelta, &item.ExperienceDelta, &item.Consumable, &item.Enabled); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// Vocals lists localized weighted species vocalizations.
func (repository *Repository) Vocals(ctx context.Context) ([]petrecord.Vocal, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select type_id,mood,text_key,weight,cooldown_ms,enabled from pet_vocals order by type_id,id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]petrecord.Vocal, 0, 36)
	for rows.Next() {
		item := petrecord.Vocal{}
		var cooldown int64
		if err = rows.Scan(&item.TypeID, &item.Mood, &item.TextKey, &item.Weight, &cooldown, &item.Enabled); err != nil {
			return nil, err
		}
		item.Cooldown = time.Duration(cooldown) * time.Millisecond
		items = append(items, item)
	}
	return items, rows.Err()
}

// BreedingRules lists parent compatibility and weighted result appearances.
func (repository *Repository) BreedingRules(ctx context.Context) ([]petrecord.BreedingRule, []petrecord.BreedingRace, error) {
	ruleRows, err := repository.executorFor(ctx).Query(ctx, `select parent_one_type_id,parent_two_type_id,result_type_id,enabled from pet_breeding_rules order by parent_one_type_id,parent_two_type_id`)
	if err != nil {
		return nil, nil, err
	}
	rules := make([]petrecord.BreedingRule, 0, 35)
	for ruleRows.Next() {
		item := petrecord.BreedingRule{}
		if err = ruleRows.Scan(&item.ParentOneTypeID, &item.ParentTwoTypeID, &item.ResultTypeID, &item.Enabled); err != nil {
			ruleRows.Close()
			return nil, nil, err
		}
		rules = append(rules, item)
	}
	if err = ruleRows.Err(); err != nil {
		ruleRows.Close()
		return nil, nil, err
	}
	ruleRows.Close()
	raceRows, err := repository.executorFor(ctx).Query(ctx, `select result_type_id,breed_id,palette_id,weight,mutation,enabled from pet_breeding_races order by result_type_id,breed_id,palette_id`)
	if err != nil {
		return nil, nil, err
	}
	defer raceRows.Close()
	races := make([]petrecord.BreedingRace, 0, 40)
	for raceRows.Next() {
		item := petrecord.BreedingRace{}
		if err = raceRows.Scan(&item.ResultTypeID, &item.BreedID, &item.PaletteID, &item.Weight, &item.Mutation, &item.Enabled); err != nil {
			return nil, nil, err
		}
		races = append(races, item)
	}
	return rules, races, raceRows.Err()
}

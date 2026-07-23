// Package recipe owns crafting matching and transactional workflows.
package recipe

import craftingrecord "github.com/niflaot/pixels/internal/realm/crafting/record"

// Match describes one recipe comparison with a definition bag.
type Match struct {
	// Recipe stores the compared recipe.
	Recipe craftingrecord.Recipe
	// Contains reports that the bag contains all required ingredients.
	Contains bool
	// Exact reports that the bag is exactly equal to the recipe.
	Exact bool
}

// MatchRecipes compares recipes without allocating per recipe.
func MatchRecipes(recipes []craftingrecord.Recipe, bag map[int64]int32) []Match {
	matches := make([]Match, 0, len(recipes))
	for _, candidate := range recipes {
		contains := true
		requiredKinds := 0
		for _, ingredient := range candidate.Ingredients {
			requiredKinds++
			if bag[ingredient.DefinitionID] < ingredient.Amount {
				contains = false
			}
		}
		exact := contains && len(bag) == requiredKinds
		if exact {
			for _, ingredient := range candidate.Ingredients {
				if bag[ingredient.DefinitionID] != ingredient.Amount {
					exact = false
					break
				}
			}
		}
		matches = append(matches, Match{Recipe: candidate, Contains: contains, Exact: exact})
	}
	return matches
}

// Hint counts partial unknown-secret matches and reports an exact one.
func Hint(recipes []craftingrecord.Recipe, bag map[int64]int32) (int32, bool) {
	var count int32
	exact := false
	for _, match := range MatchRecipes(recipes, bag) {
		partial := false
		for _, ingredient := range match.Recipe.Ingredients {
			if bag[ingredient.DefinitionID] > 0 {
				partial = true
				break
			}
		}
		if partial {
			count++
		}
		exact = exact || match.Exact
	}
	return count, exact
}

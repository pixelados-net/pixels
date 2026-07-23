package recipe

import (
	"testing"

	craftingrecord "github.com/niflaot/pixels/internal/realm/crafting/record"
)

// TestMatchRecipesDistinguishesContainsAndExact verifies partial, exact, and surplus bags.
func TestMatchRecipesDistinguishesContainsAndExact(t *testing.T) {
	recipes := []craftingrecord.Recipe{{ID: 1, Ingredients: []craftingrecord.Ingredient{{DefinitionID: 10, Amount: 2}, {DefinitionID: 20, Amount: 1}}}}
	tests := []struct {
		name     string
		bag      map[int64]int32
		contains bool
		exact    bool
	}{{"partial", map[int64]int32{10: 1}, false, false}, {"exact", map[int64]int32{10: 2, 20: 1}, true, true}, {"surplus amount", map[int64]int32{10: 3, 20: 1}, true, false}, {"surplus kind", map[int64]int32{10: 2, 20: 1, 30: 1}, true, false}}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			match := MatchRecipes(recipes, test.bag)[0]
			if match.Contains != test.contains || match.Exact != test.exact {
				t.Fatalf("contains=%v exact=%v", match.Contains, match.Exact)
			}
		})
	}
}

// TestHintCountsOnlyTouchedRecipes verifies hint aggregation semantics.
func TestHintCountsOnlyTouchedRecipes(t *testing.T) {
	recipes := []craftingrecord.Recipe{{Ingredients: []craftingrecord.Ingredient{{DefinitionID: 1, Amount: 1}}}, {Ingredients: []craftingrecord.Ingredient{{DefinitionID: 2, Amount: 1}}}}
	count, exact := Hint(recipes, map[int64]int32{1: 1})
	if count != 1 || !exact {
		t.Fatalf("count=%d exact=%v", count, exact)
	}
}

// BenchmarkRecipeMatch measures the allocation-bounded in-memory matching path.
func BenchmarkRecipeMatch(b *testing.B) {
	recipes := make([]craftingrecord.Recipe, 32)
	for index := range recipes {
		recipes[index].Ingredients = []craftingrecord.Ingredient{{DefinitionID: 1, Amount: 2}, {DefinitionID: 2, Amount: 1}}
	}
	bag := map[int64]int32{1: 2, 2: 1}
	b.ReportAllocs()
	for b.Loop() {
		_ = MatchRecipes(recipes, bag)
	}
}

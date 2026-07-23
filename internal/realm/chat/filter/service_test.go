package filter

import (
	"context"
	"testing"
)

// storeForTest stores global words in memory.
type storeForTest struct{ words []string }

// List returns stored words.
func (store *storeForTest) List(context.Context) ([]string, error) {
	return append([]string(nil), store.words...), nil
}

// Add appends one absent word.
func (store *storeForTest) Add(_ context.Context, word string) error {
	for _, current := range store.words {
		if current == word {
			return nil
		}
	}
	store.words = append(store.words, word)
	return nil
}

// Remove deletes one word.
func (store *storeForTest) Remove(_ context.Context, word string) error {
	for index, current := range store.words {
		if current == word {
			store.words = append(store.words[:index], store.words[index+1:]...)
			break
		}
	}
	return nil
}

// TestService verifies snapshot mutation and censorship.
func TestService(t *testing.T) {
	service := New(&storeForTest{})
	if err := service.Add(context.Background(), " Bad "); err != nil {
		t.Fatalf("add: %v", err)
	}
	text, changed := service.Censor("a BAD idea")
	if !changed || text != "a *** idea" || len(service.List()) != 1 {
		t.Fatalf("text=%q changed=%v words=%v", text, changed, service.List())
	}
	if err := service.Remove(context.Background(), "bad"); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if _, changed = service.Censor("bad"); changed {
		t.Fatal("removed word remained visible")
	}
}

// TestNormalizeRejectsInvalidWords verifies global dictionary validation.
func TestNormalizeRejectsInvalidWords(t *testing.T) {
	for _, word := range []string{"", "two words", "\t"} {
		if _, err := normalize(word); err == nil {
			t.Fatalf("expected %q rejection", word)
		}
	}
}

package entry

import (
	"errors"
	"os"
	"regexp"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// TestHashPasswordCreatesSaltedBcryptHashes verifies password hashing and matching.
func TestHashPasswordCreatesSaltedBcryptHashes(t *testing.T) {
	first, err := HashPassword("pixels-secret", bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	second, err := HashPassword("pixels-secret", bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash password again: %v", err)
	}
	if first == "pixels-secret" || first == second {
		t.Fatalf("expected salted non-plaintext hashes first=%q second=%q", first, second)
	}
	if !passwordMatches(&first, "pixels-secret") || passwordMatches(&first, "wrong") {
		t.Fatal("unexpected bcrypt match result")
	}
}

// TestDevelopmentSeedPasswordHashes verifies every seeded password room accepts 1234.
func TestDevelopmentSeedPasswordHashes(t *testing.T) {
	data, err := os.ReadFile("../../database/seed/development/0005_closed_rooms.sql")
	if err != nil {
		t.Fatalf("read closed room seed: %v", err)
	}
	hashes := regexp.MustCompile(`\$2a\$10\$[./A-Za-z0-9]{53}`).FindAll(data, -1)
	if len(hashes) != 2 {
		t.Fatalf("expected two password room hashes, got %d", len(hashes))
	}
	for _, hash := range hashes {
		value := string(hash)
		if !passwordMatches(&value, "1234") {
			t.Fatalf("seeded hash does not accept configured password %q", value)
		}
	}
}

// TestHashPasswordRejectsEmpty verifies empty room passwords are not hashed.
func TestHashPasswordRejectsEmpty(t *testing.T) {
	_, err := HashPassword("", bcrypt.MinCost)
	if !errors.Is(err, ErrInvalidPassword) {
		t.Fatalf("expected invalid password, got %v", err)
	}
}

package diffie

import (
	"errors"
	"testing"
)

// TestNewParameters verifies parameter validation.
func TestNewParameters(t *testing.T) {
	parameters, err := NewParameters("prime", "generator")
	if err != nil {
		t.Fatalf("new parameters: %v", err)
	}

	if parameters.EncryptedPrime != "prime" {
		t.Fatalf("expected prime, got %s", parameters.EncryptedPrime)
	}
}

// TestNewParametersRejectsEmpty verifies empty parameter validation.
func TestNewParametersRejectsEmpty(t *testing.T) {
	_, err := NewParameters("", "generator")
	if !errors.Is(err, ErrInvalidParameters) {
		t.Fatalf("expected invalid parameters, got %v", err)
	}
}

// TestNewPublicKey verifies public key validation.
func TestNewPublicKey(t *testing.T) {
	key, err := NewPublicKey("public")
	if err != nil {
		t.Fatalf("new public key: %v", err)
	}

	if key.Encrypted != "public" {
		t.Fatalf("expected public key, got %s", key.Encrypted)
	}
}

// TestNewPublicKeyRejectsEmpty verifies empty public key validation.
func TestNewPublicKeyRejectsEmpty(t *testing.T) {
	_, err := NewPublicKey("")
	if !errors.Is(err, ErrInvalidPublicKey) {
		t.Fatalf("expected invalid public key, got %v", err)
	}
}

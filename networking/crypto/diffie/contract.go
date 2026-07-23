// Package diffie defines Diffie-Hellman handshake contracts.
package diffie

import (
	"context"
	"errors"
)

var (
	// ErrInvalidParameters reports empty encrypted Diffie-Hellman parameters.
	ErrInvalidParameters = errors.New("invalid diffie parameters")

	// ErrInvalidPublicKey reports an empty encrypted Diffie-Hellman public key.
	ErrInvalidPublicKey = errors.New("invalid diffie public key")
)

// Parameters contains encrypted server Diffie-Hellman parameters.
type Parameters struct {
	// EncryptedPrime is the RSA-signed prime value.
	EncryptedPrime string
	// EncryptedGenerator is the RSA-signed generator value.
	EncryptedGenerator string
}

// PublicKey contains an encrypted Diffie-Hellman public key.
type PublicKey struct {
	// Encrypted is the RSA-protected public key value.
	Encrypted string
}

// Result contains server values produced after client key completion.
type Result struct {
	// PublicKey is the encrypted server public key.
	PublicKey PublicKey
	// ServerClientEncryption reports whether server-to-client encryption is enabled.
	ServerClientEncryption bool
}

// Provider prepares protocol Diffie-Hellman values.
type Provider interface {
	// Begin returns encrypted prime and generator values.
	Begin(context.Context) (Parameters, error)
	// Complete consumes a client public key and returns server completion values.
	Complete(context.Context, PublicKey) (Result, error)
}

// NewParameters creates encrypted Diffie-Hellman parameters.
func NewParameters(encryptedPrime string, encryptedGenerator string) (Parameters, error) {
	if encryptedPrime == "" || encryptedGenerator == "" {
		return Parameters{}, ErrInvalidParameters
	}

	return Parameters{EncryptedPrime: encryptedPrime, EncryptedGenerator: encryptedGenerator}, nil
}

// NewPublicKey creates an encrypted Diffie-Hellman public key.
func NewPublicKey(encrypted string) (PublicKey, error) {
	if encrypted == "" {
		return PublicKey{}, ErrInvalidPublicKey
	}

	return PublicKey{Encrypted: encrypted}, nil
}

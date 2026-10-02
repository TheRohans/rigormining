// Package secret generates and hashes the opaque random values used for
// sessions, OAuth state and API tokens.
package secret

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// New returns 32 bytes from crypto/rand, hex encoded.
func New() string {
	b := make([]byte, 32)
	// crypto/rand.Read never returns an error as of Go 1.24; it crashes the
	// program instead if the OS can't supply randomness.
	rand.Read(b)
	return hex.EncodeToString(b)
}

// Hash is what gets stored in the database in place of a raw secret, so a
// leaked database (or backup) can't be replayed as a login or API token.
// The inputs are 256-bit random values, so a plain SHA-256 is sufficient -
// there's nothing to brute force.
func Hash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

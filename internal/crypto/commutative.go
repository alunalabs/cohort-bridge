// commutative.go
// Package crypto provides commutative encryption using Pohlig-Hellman over Curve25519
// for secure blocking in the fuzzy matching pipeline.
package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"

	"filippo.io/edwards25519"
)

// CommutativeKey represents a private key for commutative encryption
type CommutativeKey struct {
	scalar *edwards25519.Scalar
}

// CommutativePoint represents an encrypted point on Curve25519
type CommutativePoint struct {
	point *edwards25519.Point
}

// GenerateCommutativeKey generates a new random private key for commutative encryption
func GenerateCommutativeKey() (*CommutativeKey, error) {
	// Generate random 32 bytes and reduce modulo the curve order
	var keyBytes [32]byte
	if _, err := rand.Read(keyBytes[:]); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	scalar := new(edwards25519.Scalar)
	scalar, err := scalar.SetBytesWithClamping(keyBytes[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create scalar: %w", err)
	}

	return &CommutativeKey{scalar: scalar}, nil
}

// EncryptString encrypts a string using commutative encryption
// The string is first hashed to a curve point, then multiplied by the private key
func (ck *CommutativeKey) EncryptString(input string) (*CommutativePoint, error) {
	// Hash the input string to a curve point
	point, err := hashToPoint(input)
	if err != nil {
		return nil, fmt.Errorf("failed to hash input to point: %w", err)
	}

	// Multiply the point by our scalar (private key)
	encryptedPoint := new(edwards25519.Point).ScalarMult(ck.scalar, point)

	return &CommutativePoint{point: encryptedPoint}, nil
}

// Encrypt encrypts a point that's already on the curve
func (ck *CommutativeKey) Encrypt(point *CommutativePoint) *CommutativePoint {
	encryptedPoint := new(edwards25519.Point).ScalarMult(ck.scalar, point.point)
	return &CommutativePoint{point: encryptedPoint}
}

// hashToPoint hashes a string to a point on Curve25519 using the try-and-increment method
func hashToPoint(input string) (*edwards25519.Point, error) {
	h := sha256.New()
	h.Write([]byte(input))
	hash := h.Sum(nil)

	// Try to find a valid point by incrementing a counter
	for i := 0; i < 256; i++ {
		// Modify hash slightly for each attempt
		attempt := make([]byte, len(hash))
		copy(attempt, hash)
		attempt[0] ^= byte(i)

		// Try to decode as a point
		point := new(edwards25519.Point)
		if _, err := point.SetBytes(attempt); err == nil {
			return point, nil
		}
	}

	return nil, errors.New("failed to hash string to curve point")
}

// Bytes returns the 32-byte representation of the encrypted point
func (cp *CommutativePoint) Bytes() []byte {
	return cp.point.Bytes()
}

// FromBytes creates a CommutativePoint from a 32-byte slice
func CommutativePointFromBytes(data []byte) (*CommutativePoint, error) {
	if len(data) != 32 {
		return nil, errors.New("invalid point data length")
	}

	point := new(edwards25519.Point)
	if _, err := point.SetBytes(data); err != nil {
		return nil, fmt.Errorf("failed to decode point: %w", err)
	}

	return &CommutativePoint{point: point}, nil
}

// Equal checks if two commutative points are equal
func (cp *CommutativePoint) Equal(other *CommutativePoint) bool {
	return cp.point.Equal(other.point) == 1
}

// String returns a hex representation of the point for debugging
func (cp *CommutativePoint) String() string {
	return fmt.Sprintf("%x", cp.Bytes())
}

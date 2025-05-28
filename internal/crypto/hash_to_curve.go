package crypto

import (
	"crypto/sha256"

	"filippo.io/edwards25519"
)

// HashToCurve hashes input to a point on edwards25519 by hashing to a scalar and multiplying by the base point.
func HashToCurve(input string) *edwards25519.Point {
	h := sha256.Sum256([]byte(input))
	scalar, err := new(edwards25519.Scalar).SetCanonicalBytes(h[:])
	if err != nil {
		// If the hash is not a valid scalar, use SetUniformBytes (returns (scalar, error))
		scalar2, _ := new(edwards25519.Scalar).SetUniformBytes(h[:])
		scalar = scalar2
	}
	return new(edwards25519.Point).ScalarBaseMult(scalar)
}

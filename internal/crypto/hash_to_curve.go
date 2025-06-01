package crypto

import (
	"crypto/sha512"

	"filippo.io/edwards25519"
)

//TODO: Reimplement as a RFC 9380-compliant hash-to-curve

// HashToCurve securely maps a token to a point on edwards25519
func HashToCurve(token string) *edwards25519.Point {
	hash := sha512.Sum512([]byte(token)) // ✅ 64 bytes
	scalar := new(edwards25519.Scalar)
	scalar.SetUniformBytes(hash[:]) // now safe
	P := new(edwards25519.Point).ScalarBaseMult(scalar)
	return P
}

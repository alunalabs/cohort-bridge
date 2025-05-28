package crypto

import (
	"crypto/sha256"

	"filippo.io/edwards25519"
)

// DeriveSharedKey derives a 32-byte key from point P and scalar s (ECDH)
func DeriveSharedKey(P *edwards25519.Point, s *edwards25519.Scalar) []byte {
	S := new(edwards25519.Point).ScalarMult(s, P)
	sum := sha256.Sum256(S.Bytes())
	return sum[:]
}

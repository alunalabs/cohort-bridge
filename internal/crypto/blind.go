package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"log"

	"filippo.io/edwards25519"
)

/* -------------------------------------------------------------------------- */
/*                         Scalar / Point helper utils                        */
/* -------------------------------------------------------------------------- */

// RandomScalar returns a uniformly-random scalar in edwards25519’s field.
func RandomScalar() *edwards25519.Scalar {
	buf := make([]byte, 64)
	if _, err := rand.Read(buf); err != nil {
		log.Fatalf("rand: %v", err)
	}
	s, err := new(edwards25519.Scalar).SetUniformBytes(buf)
	if err != nil {
		log.Fatalf("scalar: %v", err)
	}
	return s
}

// BlindPoint computes Q = r·P and returns (Q, r).
func BlindPoint(P *edwards25519.Point) (*edwards25519.Point, *edwards25519.Scalar) {
	r := RandomScalar()
	Q := new(edwards25519.Point).ScalarMult(r, P)
	return Q, r
}

// ReblindPoint computes Q' = s·Q   (used by the **sender**).
func ReblindPoint(Q *edwards25519.Point, s *edwards25519.Scalar) *edwards25519.Point {
	return new(edwards25519.Point).ScalarMult(s, Q)
}

// UnblindPoint computes P' = r⁻¹·Q'   (used by the **receiver**).
func UnblindPoint(Qp *edwards25519.Point, r *edwards25519.Scalar) *edwards25519.Point {
	rInv := new(edwards25519.Scalar).Invert(r)
	return new(edwards25519.Point).ScalarMult(rInv, Qp)
}

/* -------------------------------------------------------------------------- */
/*                        Deterministic key derivation                        */
/* -------------------------------------------------------------------------- */

// PointKey returns SHA-256(point.Bytes()) – 32-byte symmetric key.
func PointKey(pt *edwards25519.Point) []byte {
	h := sha256.Sum256(pt.Bytes())
	return h[:]
}

package crypto

import (
	"crypto/rand"

	"filippo.io/edwards25519"
)

// BlindPoint returns Q = r·P and r
func BlindPoint(P *edwards25519.Point) (*edwards25519.Point, *edwards25519.Scalar) {
	rBytes := make([]byte, 32)
	rand.Read(rBytes)
	r, err := new(edwards25519.Scalar).SetBytesWithClamping(rBytes)
	if err != nil {
		panic("failed to create scalar: " + err.Error())
	}
	Q := new(edwards25519.Point).ScalarMult(r, P)
	return Q, r
}

// UnblindPoint returns P = r⁻¹·Q
func UnblindPoint(Q *edwards25519.Point, r *edwards25519.Scalar) *edwards25519.Point {
	rInv := new(edwards25519.Scalar).Invert(r)
	P := new(edwards25519.Point).ScalarMult(rInv, Q)
	return P
}

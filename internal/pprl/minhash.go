// minhash.go
// Package pprl provides a simple MinHash implementation that takes the set of "1" bits
// from a Bloom filter and produces an s‐length signature vector.
// We use a family of random hash functions of the form (a*x + b) mod p, where p is prime > m.
package pprl

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"math/big"
	mathrand "math/rand"
)

// MinHash holds the parameters and signature for a given Bloom filter.
type MinHash struct {
	s         uint32   // number of hash functions / signature length
	a, b      []uint32 // random coefficients for linears hashes
	prime     uint32   // a prime > m (Bloom filter size)
	signature []uint32 // resulting signature
}

// NewMinHash returns a MinHash configured to produce an s‐length signature for filters of size m.
// It draws random (a, b) pairs from [1, prime-1]. Prime must be chosen > m (e.g., next prime after m).
func NewMinHash(m, s uint32) (*MinHash, error) {
	if m == 0 || s == 0 {
		return nil, errors.New("minhash: invalid parameters")
	}
	// Choose a prime > m. For simplicity, pick a hardcoded prime near 2^31.
	const prime uint32 = 2147483647 // Mersenne prime (2^31 - 1); must exceed m.
	if m >= prime {
		return nil, errors.New("minhash: m too large for chosen prime")
	}

	a := make([]uint32, s)
	b := make([]uint32, s)
	// Securely sample random coefficients in [1..prime-1].
	for i := uint32(0); i < s; i++ {
		a[i] = randomUint32(1, prime-1)
		b[i] = randomUint32(0, prime-1)
	}

	// Initialize signature array with "max value" sentinel (prime).
	sig := make([]uint32, s)
	for i := range sig {
		sig[i] = prime
	}

	return &MinHash{
		s:         s,
		a:         a,
		b:         b,
		prime:     prime,
		signature: sig,
	}, nil
}

// NewMinHashSeeded returns a MinHash with deterministic hash functions based on a seed.
// This ensures both parties can generate identical MinHash instances for record linkage.
func NewMinHashSeeded(m, s uint32, seed string) (*MinHash, error) {
	if m == 0 || s == 0 {
		return nil, errors.New("minhash: invalid parameters")
	}
	// Choose a prime > m. For simplicity, pick a hardcoded prime near 2^31.
	const prime uint32 = 2147483647 // Mersenne prime (2^31 - 1); must exceed m.
	if m >= prime {
		return nil, errors.New("minhash: m too large for chosen prime")
	}

	// Use deterministic random generator based on seed
	h := sha256.Sum256([]byte(seed))
	var seedInt int64
	for i := 0; i < 8; i++ {
		seedInt = (seedInt << 8) | int64(h[i])
	}
	rng := mathrand.New(mathrand.NewSource(seedInt))

	a := make([]uint32, s)
	b := make([]uint32, s)
	// Generate deterministic coefficients in [1..prime-1].
	for i := uint32(0); i < s; i++ {
		a[i] = uint32(rng.Int63n(int64(prime-2))) + 1 // [1, prime-1]
		b[i] = uint32(rng.Int63n(int64(prime)))       // [0, prime-1]
	}

	// Initialize signature array with "max value" sentinel (prime).
	sig := make([]uint32, s)
	for i := range sig {
		sig[i] = prime
	}

	return &MinHash{
		s:         s,
		a:         a,
		b:         b,
		prime:     prime,
		signature: sig,
	}, nil
}

// ComputeSignature fills mh.signature based on the set of bit‐indices where BF = 1.
// You must pass a pointer to a fully‐populated BloomFilter.
func (mh *MinHash) ComputeSignature(bf *BloomFilter) ([]uint32, error) {
	if bf == nil {
		return nil, errors.New("minhash: nil BloomFilter")
	}
	m := bf.m

	// Reset signature to initial state (prime values) before computation
	for i := uint32(0); i < mh.s; i++ {
		mh.signature[i] = mh.prime
	}

	// Iterate over all bits in bf. For any bit that's 1, record its index.
	for blockIdx, blockVal := range bf.bitArray {
		if blockVal == 0 {
			continue
		}
		for bitOff := uint32(0); bitOff < 64; bitOff++ {
			if (blockVal>>bitOff)&1 == 1 {
				idx := uint32(blockIdx)*64 + bitOff
				if idx >= m {
					break
				}
				// For each hash function i, compute h_i(idx) = (a[i]*idx + b[i]) mod prime
				for i := uint32(0); i < mh.s; i++ {
					// cast to uint64 to avoid overflow
					x := (uint64(mh.a[i])*uint64(idx) + uint64(mh.b[i])) % uint64(mh.prime)
					if uint32(x) < mh.signature[i] {
						mh.signature[i] = uint32(x)
					}
				}
			}
		}
	}
	// Return a copy
	out := make([]uint32, mh.s)
	copy(out, mh.signature)
	return out, nil
}

// randomUint32 returns a uniform random integer in [min..max], inclusive.
func randomUint32(min, max uint32) uint32 {
	if min > max {
		min, max = max, min
	}
	span := big.NewInt(int64(max) - int64(min) + 1)
	nBig, err := rand.Int(rand.Reader, span)
	if err != nil {
		panic("minhash: crypto/rand failure")
	}
	return uint32(nBig.Int64() + int64(min))
}

// JaccardSimilarity estimates Jaccard similarity between two MinHash signatures.
// It's simply the fraction of positions where sig1[i] == sig2[i].
func JaccardSimilarity(sig1, sig2 []uint32) (float64, error) {
	if len(sig1) != len(sig2) {
		return 0, errors.New("minhash: length mismatch")
	}
	var same uint32
	for i := range sig1 {
		if sig1[i] == sig2[i] {
			same++
		}
	}
	return float64(same) / float64(len(sig1)), nil
}

// MarshalBinary serializes the MinHash to a byte slice
func (mh *MinHash) MarshalBinary() ([]byte, error) {
	// Calculate buffer size: s + (s * 4) + (s * 4) + 4 + (s * 4)
	// s (4 bytes) + a array + b array + prime (4 bytes) + signature array
	bufSize := 4 + int(mh.s)*4 + int(mh.s)*4 + 4 + int(mh.s)*4
	buf := make([]byte, bufSize)

	offset := 0

	// Write s
	binary.LittleEndian.PutUint32(buf[offset:offset+4], mh.s)
	offset += 4

	// Write a array
	for i := uint32(0); i < mh.s; i++ {
		binary.LittleEndian.PutUint32(buf[offset:offset+4], mh.a[i])
		offset += 4
	}

	// Write b array
	for i := uint32(0); i < mh.s; i++ {
		binary.LittleEndian.PutUint32(buf[offset:offset+4], mh.b[i])
		offset += 4
	}

	// Write prime
	binary.LittleEndian.PutUint32(buf[offset:offset+4], mh.prime)
	offset += 4

	// Write signature array
	for i := uint32(0); i < mh.s; i++ {
		binary.LittleEndian.PutUint32(buf[offset:offset+4], mh.signature[i])
		offset += 4
	}

	return buf, nil
}

// UnmarshalBinary deserializes a MinHash from a byte slice
func (mh *MinHash) UnmarshalBinary(data []byte) error {
	if len(data) < 4 {
		return errors.New("minhash: data too short")
	}

	offset := 0

	// Read s
	s := binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	// Check expected length
	expectedLen := 4 + int(s)*4 + int(s)*4 + 4 + int(s)*4
	if len(data) != expectedLen {
		return errors.New("minhash: incorrect data length")
	}

	// Read a array
	a := make([]uint32, s)
	for i := uint32(0); i < s; i++ {
		a[i] = binary.LittleEndian.Uint32(data[offset : offset+4])
		offset += 4
	}

	// Read b array
	b := make([]uint32, s)
	for i := uint32(0); i < s; i++ {
		b[i] = binary.LittleEndian.Uint32(data[offset : offset+4])
		offset += 4
	}

	// Read prime
	prime := binary.LittleEndian.Uint32(data[offset : offset+4])
	offset += 4

	// Read signature array
	signature := make([]uint32, s)
	for i := uint32(0); i < s; i++ {
		signature[i] = binary.LittleEndian.Uint32(data[offset : offset+4])
		offset += 4
	}

	// Set fields
	mh.s = s
	mh.a = a
	mh.b = b
	mh.prime = prime
	mh.signature = signature

	return nil
}

// ToBase64 serializes the MinHash to a base64 string
func (mh *MinHash) ToBase64() (string, error) {
	data, err := mh.MarshalBinary()
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// FromBase64 deserializes a MinHash from a base64 string
func (mh *MinHash) FromBase64(encoded string) error {
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return err
	}
	return mh.UnmarshalBinary(data)
}

// MinHashToBase64 creates a base64 string from a MinHash
func MinHashToBase64(mh *MinHash) (string, error) {
	return mh.ToBase64()
}

// MinHashFromBase64 creates a MinHash from a base64 string
func MinHashFromBase64(encoded string) (*MinHash, error) {
	mh := &MinHash{}
	err := mh.FromBase64(encoded)
	if err != nil {
		return nil, err
	}
	return mh, nil
}

// GetSignature returns a copy of the computed signature
func (mh *MinHash) GetSignature() []uint32 {
	if mh.signature == nil {
		return nil
	}
	sig := make([]uint32, len(mh.signature))
	copy(sig, mh.signature)
	return sig
}

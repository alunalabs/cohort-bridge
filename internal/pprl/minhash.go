// minhash.go
// Package pprl provides a simple MinHash implementation that takes the set of “1” bits
// from a Bloom filter and produces an s‐length signature vector.
// We use a family of random hash functions of the form (a*x + b) mod p, where p is prime > m.
package pprl

import (
	"crypto/rand"
	"errors"
	"math/big"
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

	// Initialize signature array with “max value” sentinel (prime).
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
// It’s simply the fraction of positions where sig1[i] == sig2[i].
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

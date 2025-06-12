// bloom.go
// Package pprl provides a fixed-size Bloom filter with optional noise.
// Every Bloom filter is m bits long, uses k distinct hash indices per input,
// and supports serialisation to/from []byte.
package pprl

import (
	"encoding/binary"
	"errors"
	"hash/fnv"
	"math/rand"
	"time"
)

// BloomFilter is a fixed-size bitset with k hash functions.
type BloomFilter struct {
	m        uint32   // total number of bits
	k        uint32   // number of hash functions
	bitArray []uint64 // underlying bit array (length = ceil(m/64))
}

// NewBloomFilter returns an empty BloomFilter of m bits and k hashes.
func NewBloomFilter(m, k uint32) *BloomFilter {
	if m == 0 || k == 0 {
		return nil
	}
	blocks := (m + 63) / 64
	return &BloomFilter{
		m:        m,
		k:        k,
		bitArray: make([]uint64, blocks),
	}
}

// NewBloomFilterWithRandomBits returns a BloomFilter with the specified percentage of random bits set.
// randomBitsPercent should be a value between 0.0 and 1.0 representing the fraction of bits to randomly set.
func NewBloomFilterWithRandomBits(m, k uint32, randomBitsPercent float64) *BloomFilter {
	bf := NewBloomFilter(m, k)
	if bf == nil {
		return nil
	}

	// Add random bits if percentage is specified
	if randomBitsPercent > 0.0 {
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		numRandomBits := int(float64(m) * randomBitsPercent)

		for i := 0; i < numRandomBits; i++ {
			randomIdx := uint32(rng.Intn(int(m)))
			bf.setBit(randomIdx)
		}
	}

	return bf
}

// Add inserts a byte-slice (e.g. a q-gram) into the filter.
// Internally, it runs k different hash‐index computations.
func (bf *BloomFilter) Add(data []byte) {
	// For each i in [0..k), compute a hash and set bit at (h mod m).
	h1 := fnv.New64a()
	h1.Write(data)
	sum := h1.Sum64()
	seed := sum

	for i := uint32(0); i < bf.k; i++ {
		// Derive a second hash by appending the iteration index to seed.
		h2 := fnv.New64a()
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, seed^(uint64(i)))
		h2.Write(buf)
		h2.Write(data)
		idx := uint32(h2.Sum64() % uint64(bf.m))
		bf.setBit(idx)
	}
}

// setBit flips the bit at position idx to 1.
func (bf *BloomFilter) setBit(idx uint32) {
	block := idx / 64
	offset := idx % 64
	bf.bitArray[block] |= 1 << offset
}

// Test returns true if data is "probably" in the filter. False => definitely not.
func (bf *BloomFilter) Test(data []byte) bool {
	h1 := fnv.New64a()
	h1.Write(data)
	sum := h1.Sum64()
	seed := sum

	for i := uint32(0); i < bf.k; i++ {
		h2 := fnv.New64a()
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, seed^(uint64(i)))
		h2.Write(buf)
		h2.Write(data)
		idx := uint32(h2.Sum64() % uint64(bf.m))
		if !bf.getBit(idx) {
			return false
		}
	}
	return true
}

// getBit returns true if bit at idx is 1.
func (bf *BloomFilter) getBit(idx uint32) bool {
	block := idx / 64
	offset := idx % 64
	return (bf.bitArray[block] & (1 << offset)) != 0
}

// MarshalBinary serialises the BloomFilter into a byte slice:
// first 4 bytes = m, next 4 bytes = k, then the bitArray as little‐endian uint64s.
func (bf *BloomFilter) MarshalBinary() ([]byte, error) {
	totalBlocks := len(bf.bitArray)
	buf := make([]byte, 8+8*totalBlocks)
	binary.LittleEndian.PutUint32(buf[0:4], bf.m)
	binary.LittleEndian.PutUint32(buf[4:8], bf.k)
	offset := 8
	for i, blockVal := range bf.bitArray {
		binary.LittleEndian.PutUint64(buf[offset+8*i:offset+8*i+8], blockVal)
	}
	return buf, nil
}

// UnmarshalBinary populates bf from a byte slice produced by MarshalBinary.
func (bf *BloomFilter) UnmarshalBinary(data []byte) error {
	if len(data) < 8 {
		return errors.New("bloom: data too short")
	}
	m := binary.LittleEndian.Uint32(data[0:4])
	k := binary.LittleEndian.Uint32(data[4:8])
	blocks := (m + 63) / 64
	expectedLen := 8 + 8*int(blocks)
	if len(data) != expectedLen {
		return errors.New("bloom: incorrect length")
	}
	bf.m = m
	bf.k = k
	bf.bitArray = make([]uint64, blocks)
	offset := 8
	for i := 0; i < int(blocks); i++ {
		bf.bitArray[i] = binary.LittleEndian.Uint64(data[offset+8*i : offset+8*i+8])
	}
	return nil
}

// AddWithNoise flips each bit that was set by Add(data), then flips
// an additional fraction (probability p) of random bits to 1 or 0.
func (bf *BloomFilter) AddWithNoise(data []byte, p float64) {
	bf.Add(data)

	// Seed a fast PRNG for noise
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	totalBits := bf.m
	noiseCount := int(float64(totalBits) * p)

	for i := 0; i < noiseCount; i++ {
		idx := uint32(rng.Intn(int(totalBits)))
		// flip bit at idx
		block := idx / 64
		offset := idx % 64
		bf.bitArray[block] ^= 1 << offset
	}
}

// HammingDistance computes the number of differing bits between two Bloom filters.
// Returns error if they differ in size or k.
func (bf *BloomFilter) HammingDistance(other *BloomFilter) (uint32, error) {
	if bf.m != other.m || bf.k != other.k {
		return 0, errors.New("bloom: incompatible filters")
	}
	var dist uint32
	for i := range bf.bitArray {
		x := bf.bitArray[i] ^ other.bitArray[i]
		dist += uint32(popcount(x))
	}
	return dist, nil
}

// GetSize returns the size (number of bits) of the Bloom filter
func (bf *BloomFilter) GetSize() uint32 {
	return bf.m
}

// popcount returns the number of set bits in a uint64.
func popcount(x uint64) int {
	return bitsSetTable[x>>(0*16)&0xFFFF] +
		bitsSetTable[x>>(1*16)&0xFFFF] +
		bitsSetTable[x>>(2*16)&0xFFFF] +
		bitsSetTable[x>>(3*16)&0xFFFF]
}

// bitsSetTable is a 16-bit lookup table for popcount.
var bitsSetTable [1 << 16]int

func init() {
	for i := 0; i < len(bitsSetTable); i++ {
		bitsSetTable[i] = popcount16(uint16(i))
	}
}

func popcount16(x uint16) int {
	count := 0
	for x != 0 {
		x &= x - 1
		count++
	}
	return count
}

package crypto

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/hex"
)

func GenerateKey() *ecdh.PrivateKey {
	priv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}
	return priv
}

// PrivateKeyFromHex parses a hex string into an X25519 private key.
func PrivateKeyFromHex(hexStr string) (*ecdh.PrivateKey, error) {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, err
	}
	priv, err := ecdh.X25519().NewPrivateKey(bytes)
	if err != nil {
		return nil, err
	}
	return priv, nil
}

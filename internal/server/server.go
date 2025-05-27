package server

import (
	"crypto/ecdh"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strings"
)

// ListenAndServe starts a TCP server and handles a single peer connection.
func Listen(port string) error {
	addr := ":" + port
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("error listening: %w", err)
	}
	fmt.Println("Listening for peer connections on", addr)
	conn, err := ln.Accept()
	if err != nil {
		return fmt.Errorf("error accepting connection: %w", err)
	}
	fmt.Println("Peer connected:", conn.RemoteAddr())
	return nil
}

// ConnectAndServe connects to a TCP peer and starts message exchange.
func Connect(addr string) error {
	_, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("error connecting to peer: %w", err)
	}
	fmt.Println("Connected to peer at", addr)
	return nil
}

// DeriveSharedSalt creates a shared salt from private and peer public keys using X25519 and sha256.
func DeriveSharedSalt(privateKeyHex, peerPublicKey string) string {
	// Sanitize input: remove PEM headers/footers and newlines if present
	privateKeyHex = sanitizeKey(privateKeyHex)
	peerPublicKey = sanitizeKey(peerPublicKey)

	privKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid private key hex: %q (%v)\n", privateKeyHex, err)
		os.Exit(1)
	}
	peerPubBytes, err := hex.DecodeString(peerPublicKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid peer public key hex: %q (%v)\n", peerPublicKey, err)
		os.Exit(1)
	}

	curve := ecdh.X25519()
	priv, err := curve.NewPrivateKey(privKeyBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not parse private key: %v\n", err)
		os.Exit(1)
	}
	peerPub, err := curve.NewPublicKey(peerPubBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not parse peer public key: %v\n", err)
		os.Exit(1)
	}

	sharedSecret, err := priv.ECDH(peerPub)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ECDH failed: %v\n", err)
		os.Exit(1)
	}

	salt := sha256.Sum256(sharedSecret)
	return fmt.Sprintf("%x", salt[:])
}

// Helper to sanitize PEM or hex key input
func sanitizeKey(key string) string {
	key = strings.TrimSpace(key)
	if strings.HasPrefix(key, "-----BEGIN") {
		lines := strings.Split(key, "\n")
		var b strings.Builder
		for _, line := range lines {
			if strings.HasPrefix(line, "-----") {
				continue
			}
			b.WriteString(line)
		}
		return b.String()
	}
	key = strings.ReplaceAll(key, "\n", "")
	key = strings.ReplaceAll(key, "-", "")
	key = strings.TrimPrefix(key, "0x")
	return key
}

// deriveSharedSalt creates a shared salt from private and peer public keys.
func deriveSharedSalt(privateKey, peerPublicKey string) string {
	// Simple HMAC-SHA256(privateKey, peerPublicKey) for demonstration
	h := hmac.New(sha256.New, []byte(privateKey))
	h.Write([]byte(peerPublicKey))
	return fmt.Sprintf("%x", h.Sum(nil))
}

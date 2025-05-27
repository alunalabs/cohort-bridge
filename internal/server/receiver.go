package server

import (
	"fmt"
	"net"
)

// ExchangePublicKeysAndPrintSaltServer listens for a peer, then exchanges public keys and prints both keys and the derived salt.
func ExchangePublicKeysAndPrintSaltServer(port, privateKeyHex, publicKeyHex string) error {
	addr := ":" + port
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	fmt.Printf("Listening for peer connection on port %s...\n", port)
	conn, err := ln.Accept()
	if err != nil {
		return fmt.Errorf("failed to accept connection: %w", err)
	}
	defer conn.Close()

	// Now exchange public keys (receiver receives first)
	if err := exchangeAndPrint(conn, privateKeyHex, publicKeyHex, false); err != nil {
		return err
	}
	return nil
}

// ExchangePublicKeysServer listens for a peer, exchanges public keys, and returns the peer's public key.
func ExchangePublicKeysServer(port, publicKeyHex string) (string, error) {
	addr := ":" + port
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return "", fmt.Errorf("failed to listen: %w", err)
	}
	conn, err := ln.Accept()
	if err != nil {
		return "", fmt.Errorf("failed to accept connection: %w", err)
	}
	defer conn.Close()

	// Receive peer's public key
	peerPubKey, err := readMultiline(conn)
	if err != nil {
		return "", fmt.Errorf("failed to read peer public key: %w", err)
	}
	// Send our public key
	_, err = fmt.Fprintf(conn, "%s\nENDKEY\n", publicKeyHex)
	if err != nil {
		return "", fmt.Errorf("failed to send public key: %w", err)
	}
	return peerPubKey, nil
}

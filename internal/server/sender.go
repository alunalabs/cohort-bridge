package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// ExchangePublicKeysAndPrintSaltClient connects to a peer, then exchanges public keys and prints both keys and the derived salt.
func ExchangePublicKeysAndPrintSaltClient(peerAddr, privateKeyHex, publicKeyHex string) error {
	// Establish connection first
	conn, err := net.Dial("tcp", peerAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to peer: %w", err)
	}
	defer conn.Close()

	// Now exchange public keys
	if err := exchangeAndPrint(conn, privateKeyHex, publicKeyHex, true); err != nil {
		return err
	}
	return nil
}

// exchangeAndPrint handles the actual public key exchange and salt derivation.
// If isSender is true, send our key first, then receive. If false, receive first, then send.
func exchangeAndPrint(conn net.Conn, privateKeyHex, publicKeyHex string, isSender bool) error {
	if isSender {
		// Send our public key
		_, err := fmt.Fprintf(conn, "%s\nENDKEY\n", publicKeyHex)
		if err != nil {
			return fmt.Errorf("failed to send public key: %w", err)
		}
		// Receive peer's public key
		peerPubKey, err := readMultiline(conn)
		if err != nil {
			return fmt.Errorf("failed to read peer public key: %w", err)
		}
		fmt.Println("Our public key:", publicKeyHex)
		fmt.Println("Peer public key:", peerPubKey)
		sharedSalt := DeriveSharedSalt(privateKeyHex, peerPubKey)
		fmt.Println("Derived shared salt:", sharedSalt)
		fmt.Println("Done. Exiting for debugging.")
	} else {
		// Receive peer's public key
		peerPubKey, err := readMultiline(conn)
		if err != nil {
			return fmt.Errorf("failed to read peer public key: %w", err)
		}
		// Send our public key
		_, err = fmt.Fprintf(conn, "%s\nENDKEY\n", publicKeyHex)
		if err != nil {
			return fmt.Errorf("failed to send public key: %w", err)
		}
		fmt.Println("Our public key:", publicKeyHex)
		fmt.Println("Peer public key:", peerPubKey)
		sharedSalt := DeriveSharedSalt(privateKeyHex, peerPubKey)
		fmt.Println("Derived shared salt:", sharedSalt)
		fmt.Println("Done. Exiting for debugging.")
	}
	return nil
}

// ExchangePublicKeysClient connects to a peer, exchanges public keys, and returns the peer's public key.
func ExchangePublicKeysClient(peerAddr, publicKeyHex string) (string, error) {
	conn, err := net.Dial("tcp", peerAddr)
	if err != nil {
		return "", fmt.Errorf("failed to connect to peer: %w", err)
	}
	defer conn.Close()

	// Send our public key
	_, err = fmt.Fprintf(conn, "%s\nENDKEY\n", publicKeyHex)
	if err != nil {
		return "", fmt.Errorf("failed to send public key: %w", err)
	}

	// Receive peer's public key (read until a delimiter line)
	peerPubKey, err := readMultiline(conn)
	if err != nil {
		return "", fmt.Errorf("failed to read peer public key: %w", err)
	}
	return peerPubKey, nil
}

// Helper to read a multi-line public key until ENDKEY line
func readMultiline(conn net.Conn) (string, error) {
	var lines []string
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "ENDKEY" {
			break
		}
		lines = append(lines, line)
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return strings.Join(lines, "\n"), nil
}

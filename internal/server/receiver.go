package server

import (
	"bufio"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"

	"filippo.io/edwards25519"
	"github.com/auroradata-ai/cohort-bridge/internal/config"
	"github.com/auroradata-ai/cohort-bridge/internal/crypto"
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

// RunReceiverPSI executes the receiver side of the Private Set Intersection protocol.
func RunReceiverPSI(cfg *config.Config) error {
	// Open patients2.csv
	csvFile, err := os.Open("patients2.csv")
	if err != nil {
		return fmt.Errorf("open patients2.csv: %w", err)
	}
	defer csvFile.Close()
	reader := csv.NewReader(bufio.NewReader(csvFile))
	_, err = reader.Read()
	if err != nil {
		return fmt.Errorf("read header: %w", err)
	}
	idIdx := 0 // Assume first column is ID

	// Connect to sender
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.ListenPort))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	defer ln.Close()
	fmt.Printf("Waiting for sender on port %d...\n", cfg.ListenPort)
	conn, err := ln.Accept()
	if err != nil {
		return fmt.Errorf("accept: %w", err)
	}
	defer conn.Close()

	// Step 1: Hash and blind IDs, send to sender
	var blinds []*edwards25519.Scalar
	var blindedPointsHex []string
	var ids []string
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("csv read: %w", err)
		}
		id := row[idIdx]
		ids = append(ids, id)
		P := crypto.HashToCurve(id)
		Q, r := crypto.BlindPoint(P)
		blinds = append(blinds, r)
		blindedPointsHex = append(blindedPointsHex, hex.EncodeToString(Q.Bytes()))
	}
	// Send number of points
	fmt.Fprintf(conn, "%d\n", len(blindedPointsHex))
	// Send each blinded point as hex
	for _, qHex := range blindedPointsHex {
		fmt.Fprintf(conn, "%s\n", qHex)
	}

	// Step 2: Receive encrypted blobs
	respReader := bufio.NewReader(conn)
	numRespLine, err := respReader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("read num responses: %w", err)
	}
	var numResp int
	fmt.Sscanf(numRespLine, "%d", &numResp)
	results := [][]string{}
	for i := 0; i < numResp; i++ {
		encLine, err := respReader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("read enc line: %w", err)
		}
		encLine = encLine[:len(encLine)-1]
		// Format: idx:hex_nonce:hex_ciphertext
		var idx int
		var nonceHex, ctHex string
		fmt.Sscanf(encLine, "%d:%s:%s", &idx, &nonceHex, &ctHex)
		nonce, _ := hex.DecodeString(nonceHex)
		ct, _ := hex.DecodeString(ctHex)
		// Derive shared key
		P := crypto.HashToCurve(ids[idx])
		r := blinds[idx]
		key := crypto.DeriveSharedKey(P, r)
		plaintext, err := crypto.DecryptAESGCM(key, nonce, ct)
		if err != nil {
			continue
		}
		results = append(results, []string{ids[idx], string(plaintext)})
	}
	// Write results.csv
	out, err := os.Create("results.csv")
	if err != nil {
		return fmt.Errorf("create results.csv: %w", err)
	}
	defer out.Close()
	w := csv.NewWriter(out)
	w.Write([]string{"id", "metadata"})
	for _, row := range results {
		w.Write(row)
	}
	w.Flush()
	fmt.Printf("Wrote %d matches to results.csv\n", len(results))
	return nil
}

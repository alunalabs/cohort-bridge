package server

import (
	"bufio"
	"crypto/rand"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"filippo.io/edwards25519"
	"github.com/auroradata-ai/cohort-bridge/internal/config"
	"github.com/auroradata-ai/cohort-bridge/internal/crypto"
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

func RunSenderPSI(cfg *config.Config) error {
	// Open patients.csv
	csvFile, err := os.Open("patients.csv")
	if err != nil {
		return fmt.Errorf("open patients.csv: %w", err)
	}
	defer csvFile.Close()
	reader := csv.NewReader(bufio.NewReader(csvFile))
	_, err = reader.Read()
	if err != nil {
		return fmt.Errorf("read header: %w", err)
	}
	idIdx := 0   // Assume first column is ID
	metaIdx := 1 // Assume second column is metadata

	// Build map: H(y) hex -> metadata
	hashMap := make(map[string]string)
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("csv read: %w", err)
		}
		id := row[idIdx]
		meta := row[metaIdx]
		P := crypto.HashToCurve(id)
		hashMap[hex.EncodeToString(P.Bytes())] = meta
	}

	// Connect to receiver
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", cfg.Peer.Host, cfg.Peer.Port))
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()
	connReader := bufio.NewReader(conn)

	// Step 1: Receive number of blinded points
	numLine, err := connReader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("read num: %w", err)
	}
	var num int
	fmt.Sscanf(numLine, "%d", &num)
	blindedPoints := make([]*edwards25519.Point, num)
	for i := 0; i < num; i++ {
		qHex, err := connReader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("read qHex: %w", err)
		}
		qHex = qHex[:len(qHex)-1]
		qBytes, _ := hex.DecodeString(qHex)
		Q, err := new(edwards25519.Point).SetBytes(qBytes)
		if err != nil {
			return fmt.Errorf("invalid point: %w", err)
		}
		blindedPoints[i] = Q
	}

	// Step 2: For each Q_x, try to match H(y)
	type encResp struct {
		idx       int
		nonce, ct []byte
	}
	var responses []encResp
	for idx, Q := range blindedPoints {
		// Try all H(y)
		for hHex, meta := range hashMap {
			HyBytes, _ := hex.DecodeString(hHex)
			Hy, _ := new(edwards25519.Point).SetBytes(HyBytes)
			// Try to compute shared key: ECDH(Q, 1) == Q
			// In practice, sender cannot unblind, but can use Q as ECDH base
			oneScalar := new(edwards25519.Scalar)
			oneScalarBytes := [32]byte{}
			oneScalarBytes[0] = 1
			oneScalar.SetCanonicalBytes(oneScalarBytes[:])
			key := crypto.DeriveSharedKey(Q, oneScalar)
			nonce := make([]byte, 12)
			rand.Read(nonce)
			ct, _, err := crypto.EncryptAESGCM(key, []byte(meta))
			if err != nil {
				continue
			}
			// For demo, match if Q.Bytes() == Hy.Bytes()
			if string(Q.Bytes()) == string(Hy.Bytes()) {
				responses = append(responses, encResp{idx, nonce, ct})
				break
			}
		}
	}
	// Send number of responses
	fmt.Fprintf(conn, "%d\n", len(responses))
	for _, r := range responses {
		fmt.Fprintf(conn, "%d:%s:%s\n", r.idx, hex.EncodeToString(r.nonce), hex.EncodeToString(r.ct))
	}
	fmt.Printf("Sent %d matches\n", len(responses))
	return nil
}

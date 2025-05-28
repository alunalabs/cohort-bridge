package server

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"os"

	"filippo.io/edwards25519"
	"github.com/auroradata-ai/cohort-bridge/internal/config"
	"github.com/auroradata-ai/cohort-bridge/internal/crypto"
)

// Receiver: blinds, sends, receives, decrypts, writes results.csv
func RunPSIReceiver(cfg *config.Config, tokens []string) ([]string, error) {
	fmt.Printf("[Receiver] Listening for PSI on port %d...\n", cfg.ListenPort)
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.ListenPort))
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}
	defer ln.Close()
	fmt.Printf("[Receiver] Waiting for sender to connect...\n")
	conn, err := ln.Accept()
	if err != nil {
		return nil, fmt.Errorf("accept: %w", err)
	}
	defer conn.Close()
	fmt.Printf("[Receiver] Sender connected.\n")

	var blinds []*edwards25519.Scalar
	var blindedPointsHex []string
	fmt.Printf("[Receiver] Blinding %d tokens and sending to sender...\n", len(tokens))
	for _, id := range tokens {
		P := crypto.HashToCurve(id)
		Q, r := crypto.BlindPoint(P)
		blinds = append(blinds, r)
		blindedPointsHex = append(blindedPointsHex, hex.EncodeToString(Q.Bytes()))
	}
	fmt.Fprintf(conn, "%d\n", len(blindedPointsHex))
	for _, qHex := range blindedPointsHex {
		fmt.Fprintf(conn, "%s\n", qHex)
	}
	fmt.Printf("[Receiver] Sent all blinded tokens to sender.\n")

	respReader := bufio.NewReader(conn)
	numRespLine, err := respReader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("read num responses: %w", err)
	}
	var numResp int
	fmt.Sscanf(numRespLine, "%d", &numResp)
	fmt.Printf("[Receiver] Expecting %d encrypted responses from sender.\n", numResp)
	results := [][]string{}
	var intersection []string
	for i := 0; i < numResp; i++ {
		encLine, err := respReader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("read enc line: %w", err)
		}
		encLine = encLine[:len(encLine)-1]
		var idx int
		var nonceHex, ctHex string
		fmt.Sscanf(encLine, "%d:%s:%s", &idx, &nonceHex, &ctHex)
		nonce, _ := hex.DecodeString(nonceHex)
		ct, _ := hex.DecodeString(ctHex)
		P := crypto.HashToCurve(tokens[idx])
		r := blinds[idx]
		key := crypto.DeriveSharedKey(P, r)
		plaintext, err := crypto.DecryptAESGCM(key, nonce, ct)
		if err != nil {
			fmt.Printf("[Receiver] Warning: failed to decrypt response for token %s: %v\n", tokens[idx], err)
			continue
		}
		results = append(results, []string{tokens[idx], string(plaintext)})
		intersection = append(intersection, tokens[idx])
		fmt.Printf("[Receiver] Decrypted match for token %s\n", tokens[idx])
	}
	out, err := os.Create("results.csv")
	if err != nil {
		return intersection, fmt.Errorf("create results.csv: %w", err)
	}
	defer out.Close()
	w := bufio.NewWriter(out)
	w.WriteString("id,metadata\n")
	for _, row := range results {
		w.WriteString(fmt.Sprintf("%s,%s\n", row[0], row[1]))
	}
	w.Flush()
	fmt.Printf("[Receiver] Wrote %d matches to results.csv\n", len(results))
	return intersection, nil
}

// Sender: matches, encrypts, sends
func RunPSISender(cfg *config.Config, tokens map[string]string) ([]string, error) {
	fmt.Printf("[Sender] Connecting to receiver at %s:%d for PSI...\n", cfg.Peer.Host, cfg.Peer.Port)
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", cfg.Peer.Host, cfg.Peer.Port))
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()
	connReader := bufio.NewReader(conn)

	fmt.Println("[Sender] Waiting to receive number of blinded points...")
	numLine, err := connReader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("read num: %w", err)
	}
	var num int
	fmt.Sscanf(numLine, "%d", &num)
	fmt.Printf("[Sender] Expecting %d blinded points from receiver.\n", num)
	blindedPoints := make([]*edwards25519.Point, num)
	for i := 0; i < num; i++ {
		qHex, err := connReader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("read blinded point: %w", err)
		}
		qHex = qHex[:len(qHex)-1]
		qBytes, _ := hex.DecodeString(qHex)
		Q, err := new(edwards25519.Point).SetBytes(qBytes)
		if err != nil || Q == nil {
			fmt.Printf("[Sender] Warning: failed to parse blinded point at index %d\n", i)
			continue
		}
		blindedPoints[i] = Q
	}

	type encResp struct {
		idx       int
		nonce, ct []byte
	}
	var responses []encResp
	var intersection []string
	for idx, Q := range blindedPoints {
		if Q == nil {
			continue
		}
		for hHex, meta := range tokens {
			HyBytes, _ := hex.DecodeString(hHex)
			Hy, err := new(edwards25519.Point).SetBytes(HyBytes)
			if err != nil || Hy == nil {
				fmt.Printf("[Sender] Warning: failed to parse token point for %s\n", hHex)
				continue
			}
			if string(Q.Bytes()) == string(Hy.Bytes()) {
				oneScalar := new(edwards25519.Scalar)
				oneScalarBytes := [32]byte{1}
				oneScalar.SetCanonicalBytes(oneScalarBytes[:])
				key := crypto.DeriveSharedKey(Q, oneScalar)
				nonce := make([]byte, 12)
				rand.Read(nonce)
				ct, nonce, err := crypto.EncryptAESGCM(key, []byte(meta))
				if err != nil {
					continue
				}
				responses = append(responses, encResp{idx, nonce, ct})
				intersection = append(intersection, hHex)
				break
			}
		}
	}
	fmt.Fprintf(conn, "%d\n", len(responses))
	for _, r := range responses {
		fmt.Fprintf(conn, "%d:%s:%s\n", r.idx, hex.EncodeToString(r.nonce), hex.EncodeToString(r.ct))
	}
	fmt.Printf("Sent %d matches\n", len(responses))
	return intersection, nil
}

// internal/server/psi.go
package server

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"

	"filippo.io/edwards25519"
	"github.com/auroradata-ai/cohort-bridge/internal/config"
	"github.com/auroradata-ai/cohort-bridge/internal/crypto"
)

/* -------------------------------------------------------------------------- */
/*                              Receiver (Party A)                            */
/* -------------------------------------------------------------------------- */

func RunPSIReceiver(cfg *config.Config, tokenToID map[string]string) ([]string, error) {
	fmt.Printf("[Receiver] Listening on :%d …\n", cfg.ListenPort)
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.ListenPort))
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}
	defer ln.Close()

	conn, err := ln.Accept()
	if err != nil {
		return nil, fmt.Errorf("accept: %w", err)
	}
	defer conn.Close()
	fmt.Println("[Receiver] Sender connected")

	/* ---------- 1 ─ blind tokens and send them ----------------------------- */

	var (
		blinds       []*edwards25519.Scalar
		blindedHex   []string
		tokenOrdered []string
	)
	for tok := range tokenToID {
		P := crypto.HashToCurve(tok)
		Q, r := crypto.BlindPoint(P)
		blinds = append(blinds, r)
		blindedHex = append(blindedHex, hex.EncodeToString(Q.Bytes()))
		tokenOrdered = append(tokenOrdered, tok)
	}
	fmt.Printf("[Receiver] Blinded %d tokens\n", len(blindedHex))

	fmt.Fprintf(conn, "%d\n", len(blindedHex))
	for _, h := range blindedHex {
		fmt.Fprintf(conn, "%s\n", h)
	}
	fmt.Println("[Receiver] Sent blinded list")

	/* ---------- 2 ─ read encrypted matches --------------------------------- */

	rdr := bufio.NewReader(conn)
	var numResp int
	if _, err := fmt.Fscanf(rdr, "%d\n", &numResp); err != nil {
		return nil, fmt.Errorf("read resp-count: %w", err)
	}
	fmt.Printf("[Receiver] Expecting %d encrypted rows\n", numResp)

	type row struct{ id, meta string }
	var (
		results      []row
		intersection []string
	)

	for i := 0; i < numResp; i++ {
		line, _ := rdr.ReadString('\n')
		var idx int
		var qPrimeHex, nonceHex, ctHex string
		fmt.Sscanf(line, "%d:%s:%s:%s", &idx, &qPrimeHex, &nonceHex, &ctHex)

		r := blinds[idx]
		QpB, _ := hex.DecodeString(qPrimeHex)
		Qp, _ := new(edwards25519.Point).SetBytes(QpB)

		Pp := crypto.UnblindPoint(Qp, r) // r⁻¹·Q' = s·P
		key := sha256.Sum256(Pp.Bytes())

		nonce, _ := hex.DecodeString(nonceHex)
		ct, _ := hex.DecodeString(ctHex)
		_, err := crypto.DecryptAESGCM(key[:], nonce, ct)
		if err != nil {
			fmt.Println("decrypt fail:", err)
			continue
		}

		id := tokenToID[tokenOrdered[idx]]
		fmt.Printf("[Recv] ok  %-2d key %x… %s\n", idx, key[:4], id)
	}

	out, err := os.Create("results.csv")
	if err != nil {
		return intersection, fmt.Errorf("csv: %w", err)
	}
	defer out.Close()
	w := bufio.NewWriter(out)
	w.WriteString("id,metadata\n")
	for _, r := range results {
		w.WriteString(fmt.Sprintf("%s,%s\n", r.id, r.meta))
	}
	w.Flush()
	fmt.Printf("[Receiver] MATCHES = %d\n", len(results))
	return intersection, nil
}

/* -------------------------------------------------------------------------- */
/*                               Sender (Party B)                             */
/* -------------------------------------------------------------------------- */

func RunPSISender(cfg *config.Config, tokenToID map[string]string) ([]string, error) {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", cfg.Peer.Host, cfg.Peer.Port))
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	rdr := bufio.NewReader(conn)

	var n int
	fmt.Fscanf(rdr, "%d\n", &n)

	blinded := make([]*edwards25519.Point, n)
	for i := range blinded {
		h, _ := rdr.ReadString('\n')
		h = h[:len(h)-1]
		b, _ := hex.DecodeString(h)
		blinded[i], _ = new(edwards25519.Point).SetBytes(b)
	}

	s := crypto.RandomScalar() // one secret
	pointKey := crypto.PointKey

	/* build map: Hash(s·P) -> id  (for quick “which id does this belong to?”) */
	idx := make(map[string]string)
	for tok, id := range tokenToID {
		P := crypto.HashToCurve(tok)
		sP := new(edwards25519.Point).ScalarMult(s, P)
		idx[hex.EncodeToString(pointKey(sP))] = id
	}

	type row struct {
		idx       int
		qPrimeHex string
		nonce, ct []byte
	}
	var rows []row
	for i, Q := range blinded {
		Qp := crypto.ReblindPoint(Q, s)               // Q' = s·Q
		k := pointKey(Qp)                             // same key both sides
		if id, ok := idx[hex.EncodeToString(k)]; ok { // match found
			nonce, ct, _ := crypto.EncryptAESGCM(k, []byte(id))
			rows = append(rows, row{
				idx:       i,
				qPrimeHex: hex.EncodeToString(Qp.Bytes()),
				nonce:     nonce, ct: ct,
			})
			fmt.Printf("[Sender] hit  %-2d key %x… %s\n", i, k[:4], id)
		}
	}

	/* --- send rows: idx|Q'|nonce|ct -------------------------------------- */
	fmt.Fprintf(conn, "%d\n", len(rows))
	for _, r := range rows {
		fmt.Fprintf(conn, "%d:%s:%s:%s\n",
			r.idx, r.qPrimeHex,
			hex.EncodeToString(r.nonce),
			hex.EncodeToString(r.ct))
	}
	fmt.Printf("[Sender] rows sent = %d\n", len(rows))
	return nil, nil
}

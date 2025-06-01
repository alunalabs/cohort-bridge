package main

import (
	"fmt"
	"os"
	"time"

	"github.com/auroradata-ai/cohort-bridge/internal/config"
	"github.com/auroradata-ai/cohort-bridge/internal/crypto"
	"github.com/auroradata-ai/cohort-bridge/internal/db"
	"github.com/auroradata-ai/cohort-bridge/internal/server"
)

func RunAsSender(cfg *config.Config) {
	peerAddr := fmt.Sprintf("%s:%d", cfg.Peer.Host, cfg.Peer.Port)

	fmt.Printf("[Sender] Connecting to peer at %s...\n", peerAddr)
	peerPubKey, err := server.ExchangePublicKeysClient(peerAddr, cfg.PublicKey)
	if err != nil {
		fmt.Println("[Sender] Error exchanging public keys:", err)
		os.Exit(1)
	}
	fmt.Println("[Sender] Peer public key:", peerPubKey)

	sharedSalt := server.DeriveSharedSalt(cfg.PrivateKey, peerPubKey)
	fmt.Println("[Sender] Derived shared salt:", sharedSalt)

	// --- Tokenize all patients using the shared salt ---
	fmt.Println("[Sender] Tokenizing all patients...")
	tokenMap, err := getTokenMapSender(cfg, sharedSalt) // id -> token
	if err != nil {
		fmt.Println("[Sender] Error tokenizing patients:", err)
		os.Exit(1)
	}
	fmt.Printf("[Sender] Tokenized %d patients.\n", len(tokenMap))

	// --- Run PSI protocol as sender ---
	psiCfg := *cfg
	psiCfg.Peer.Port = cfg.Peer.Port + 1
	fmt.Printf("[Sender] Connecting to PSI on port %d...\n", psiCfg.Peer.Port)

	// Wait for a second
	time.Sleep(1 * time.Second)

	// Wait for receiver to be ready before connecting (retry loop)
	maxAttempts := 3
	var mapped_tokens []string
	var psiErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		mapped_tokens, psiErr = server.RunPSISender(&psiCfg, tokenMap)
		if psiErr == nil {
			break
		}
		fmt.Printf("[Sender] PSI connect attempt %d failed: %v\n", attempt, psiErr)
		if attempt == maxAttempts {
			fmt.Println("[Sender] PSI sender error:", psiErr)
			os.Exit(1)
		}
		time.Sleep(1 * time.Second)
	}
	fmt.Printf("[Sender] PSI protocol complete.\n")
	fmt.Printf("[Sender] %d tokens mapped to receiver's tokens.\n", len(mapped_tokens))

	// // --- Compute intersection with results.csv ---
	// fmt.Println("[Sender] Computing intersection with results.csv...")
	// intersection, err := computeIntersection(mapped_tokens, "results.csv")
	// if err != nil {
	// 	fmt.Println("[Sender] Error computing intersection:", err)
	// 	os.Exit(1)
	// }
	// fmt.Println("[Sender] Intersection tokens:")
	// for _, t := range intersection {
	// 	fmt.Println(t)
	// }
	// fmt.Printf("[Sender] Intersection size: %d\n", len(intersection))
}

// getTokenMapSender retrieves and tokenizes records from a database based on the provided configuration and salt.
// Returns a map[id] = token
func getTokenMapSender(cfg *config.Config, salt string) (map[string]string, error) {
	dbase, err := db.GetDatabaseFromConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}
	// List all keys
	data, err := dbase.List(0, 1000000)
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}
	// Read all records
	var records []map[string]any
	for _, record := range data {
		fields := cfg.Database.Fields
		rec := make(map[string]any)
		for _, f := range fields {
			rec[f] = record[f]
		}
		records = append(records, rec)
	}
	// Tokenize
	tokens, err := crypto.TokenizeRecords(records, cfg.Database.Fields, salt)
	if err != nil {
		return nil, fmt.Errorf("tokenization failed: %w", err)
	}
	// Build map[token] = id
	tokenToID := make(map[string]string)
	for i, record := range data {
		id, ok := record["id"]
		if !ok {
			continue
		}
		tokenToID[tokens[i]] = id
	}
	return tokenToID, nil
}

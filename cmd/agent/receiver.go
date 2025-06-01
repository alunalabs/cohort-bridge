package main

import (
	"fmt"
	"os"

	"github.com/auroradata-ai/cohort-bridge/internal/config"
	"github.com/auroradata-ai/cohort-bridge/internal/crypto"
	"github.com/auroradata-ai/cohort-bridge/internal/db"
	"github.com/auroradata-ai/cohort-bridge/internal/server"
)

func RunAsReceiver(cfg *config.Config) {
	port := fmt.Sprintf("%d", cfg.ListenPort)

	fmt.Printf("[Receiver] Listening for peer connection on port %s...\n", port)
	peerPubKey, err := server.ExchangePublicKeysServer(port, cfg.PublicKey)
	if err != nil {
		fmt.Println("[Receiver] Error exchanging public keys:", err)
		os.Exit(1)
	}
	fmt.Println("[Receiver] Peer public key:", peerPubKey)

	sharedSalt := server.DeriveSharedSalt(cfg.PrivateKey, peerPubKey)
	fmt.Println("[Receiver] Derived shared salt:", sharedSalt)

	// --- Tokenize all patients using the shared salt ---
	fmt.Println("[Receiver] Tokenizing all patients...")
	tokens, err := getTokenMapReceiver(cfg, sharedSalt) // token -> id
	if err != nil {
		fmt.Println("[Receiver] Error tokenizing patients:", err)
		os.Exit(1)
	}
	fmt.Printf("[Receiver] Tokenized %d patients.\n", len(tokens))

	// --- Run PSI protocol as receiver ---
	psiPort := cfg.ListenPort + 1
	psiCfg := *cfg
	psiCfg.ListenPort = psiPort
	fmt.Printf("[Receiver] Listening for PSI on port %d...\n", psiPort)
	mapped_tokens, err := server.RunPSIReceiver(&psiCfg, tokens)
	if err != nil {
		fmt.Println("[Receiver] PSI receiver error:", err)
		os.Exit(1)
	}
	fmt.Printf("[Receiver] PSI protocol complete. %d mapped tokens returned.\n", len(mapped_tokens))
}

// getTokenMapReceiver retrieves and tokenizes records from a database based on the provided configuration and salt.
// Returns a map[token] = id
func getTokenMapReceiver(cfg *config.Config, salt string) (map[string]string, error) {
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

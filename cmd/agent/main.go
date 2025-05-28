package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/auroradata-ai/cohort-bridge/internal/config"
	"github.com/auroradata-ai/cohort-bridge/internal/crypto"
	"github.com/auroradata-ai/cohort-bridge/internal/db"
	"github.com/auroradata-ai/cohort-bridge/internal/server"
	_ "github.com/lib/pq"
	"github.com/manifoldco/promptui"
)

func main() {
	fmt.Println("🩺 Welcome to CohortBridge CLI")

	// --- Add CLI flags ---
	modeFlag := flag.String("mode", "", "Mode: send or receive")
	configFlag := flag.String("config", "", "Path to config YAML file")
	flag.Parse()

	var mode string
	var configPath string
	var err error

	// --- Mode selection ---
	if *modeFlag != "" && (*modeFlag == "send" || *modeFlag == "receive") {
		mode = *modeFlag
	} else {
		modePrompt := promptui.Select{
			Label: "Select mode",
			Items: []string{"send", "receive"},
		}
		_, mode, err = modePrompt.Run()
		if err != nil {
			fmt.Println("Prompt failed:", err)
			os.Exit(1)
		}
	}

	// --- Config file selection ---
	if *configFlag != "" {
		configPath = *configFlag
	} else {
		// List .yaml config files in current directory
		yamlFiles := []string{}
		_ = filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".yaml") {
				yamlFiles = append(yamlFiles, path)
			}
			return nil
		})
		if len(yamlFiles) == 0 {
			yamlFiles = append(yamlFiles, "config.yaml")
		}

		configPrompt := promptui.Select{
			Label: "Select config file",
			Items: yamlFiles,
		}
		_, configPath, err = configPrompt.Run()
		if err != nil {
			fmt.Println("Prompt failed:", err)
			os.Exit(1)
		}
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Println("Error loading config:", err)
		os.Exit(1)
	}

	// --- Handle private key logic ---
	if cfg.PrivateKey == "" {
		fmt.Println("No private key found in config. Generating a new one...")
		priv := crypto.GenerateKey()
		privBytes := priv.Bytes()
		cfg.PrivateKey = fmt.Sprintf("%x", privBytes)
		fmt.Printf("Generated private key (hex): %s\n", cfg.PrivateKey)
		pubBytes := priv.PublicKey().Bytes()
		cfg.PublicKey = fmt.Sprintf("%x", pubBytes)
		fmt.Printf("Derived public key (hex): %s\n", cfg.PublicKey)
	} else {
		priv, err := crypto.PrivateKeyFromHex(cfg.PrivateKey)
		if err != nil {
			fmt.Println("Invalid private key in config:", err)
			os.Exit(1)
		}
		pubBytes := priv.PublicKey().Bytes()
		cfg.PublicKey = fmt.Sprintf("%x", pubBytes)
	}

	switch mode {
	case "send":
		runAsSender(cfg)
	case "receive":
		runAsReceiver(cfg)
	}
}

func runAsSender(cfg *config.Config) {
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
	tokens, err := tokenizeAllPatientsReturnTokens(cfg, sharedSalt)
	if err != nil {
		fmt.Println("[Sender] Error tokenizing patients:", err)
		os.Exit(1)
	}
	fmt.Printf("[Sender] Tokenized %d patients.\n", len(tokens))

	// --- Run PSI protocol as sender ---
	psiCfg := *cfg
	psiCfg.Peer.Port = cfg.Peer.Port + 1
	fmt.Printf("[Sender] Connecting to PSI on port %d...\n", psiCfg.Peer.Port)

	// Wait for receiver to be ready before connecting (retry loop)
	maxAttempts := 10
	var mapped_tokens []string
	var psiErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		mapped_tokens, psiErr = server.RunPSISender(&psiCfg, buildTokenMapForPSI(cfg))
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

	// --- Compute intersection with results.csv ---
	fmt.Println("[Sender] Computing intersection with results.csv...")
	intersection, err := computeIntersection(mapped_tokens, "results.csv")
	if err != nil {
		fmt.Println("[Sender] Error computing intersection:", err)
		os.Exit(1)
	}
	fmt.Println("[Sender] Intersection tokens:")
	for _, t := range intersection {
		fmt.Println(t)
	}
	fmt.Printf("[Sender] Intersection size: %d\n", len(intersection))
}

// Helper to build tokenMap for PSI sender
func buildTokenMapForPSI(cfg *config.Config) map[string]string {
	tokenMap := make(map[string]string)
	for _, rec := range getPatientRecords(cfg) {
		idVal, ok := rec["id"]
		if !ok || idVal == nil {
			continue
		}
		id, ok := idVal.(string)
		if !ok {
			id = fmt.Sprintf("%v", idVal)
		}
		P := crypto.HashToCurve(id)
		tokenHex := fmt.Sprintf("%x", P.Bytes())
		tokenMap[tokenHex] = id
	}
	return tokenMap
}

func runAsReceiver(cfg *config.Config) {
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
	tokens, err := tokenizeAllPatientsReturnTokens(cfg, sharedSalt)
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
	mapped_tokens, err := server.RunPSIReceiver(&psiCfg, getCurveTokens(cfg))
	if err != nil {
		fmt.Println("[Receiver] PSI receiver error:", err)
		os.Exit(1)
	}
	fmt.Printf("[Receiver] PSI protocol complete. %d mapped tokens returned.\n", len(mapped_tokens))

	// --- Compute intersection with results.csv ---
	fmt.Println("[Receiver] Computing intersection with results.csv...")
	intersection, err := computeIntersection(mapped_tokens, "results.csv")
	if err != nil {
		fmt.Println("[Receiver] Error computing intersection:", err)
		os.Exit(1)
	}
	fmt.Println("[Receiver] Intersection tokens:")
	for _, t := range intersection {
		fmt.Println(t)
	}
	fmt.Printf("[Receiver] Intersection size: %d\n", len(intersection))
}

// tokenizeAllPatients loads the database, reads all patient records, and tokenizes them.
func tokenizeAllPatients(cfg *config.Config, salt string) error {
	if cfg.Database.Type != "csv" {
		return fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
	}
	csvPath := cfg.Database.Host
	if csvPath == "" {
		csvPath = cfg.Database.Table
	}
	if csvPath == "" && cfg.Database.Type == "csv" {
		csvPath = cfg.Database.Filename
	}
	if csvPath == "" {
		// Try to get from config field "filename"
		if v, ok := any(cfg.Database).(map[string]interface{})["filename"]; ok {
			csvPath, _ = v.(string)
		}
	}
	if csvPath == "" {
		return fmt.Errorf("CSV filename not specified in config")
	}
	// Try all possible fields for filename
	if cfg.Database.Filename != "" {
		csvPath = cfg.Database.Filename
	}
	dbase, err := db.NewCSVDatabase(csvPath)
	if err != nil {
		return fmt.Errorf("failed to open CSV database: %w", err)
	}
	// List all keys
	keys, err := dbase.List(0, 1000000)
	if err != nil {
		return fmt.Errorf("failed to list keys: %w", err)
	}
	// Read all records
	var records []map[string]interface{}
	for _, key := range keys {
		val, err := dbase.Get(key)
		if err != nil {
			continue
		}
		// Assume CSV value is a comma-separated string of fields
		fields := cfg.Database.Fields
		vals := strings.Split(val, ",")
		rec := make(map[string]interface{})
		for i, f := range fields {
			if i < len(vals) {
				rec[f] = vals[i]
			}
		}
		records = append(records, rec)
	}
	// Tokenize
	tokens, err := crypto.TokenizeRecords(records, cfg.Database.Fields, salt)
	if err != nil {
		return fmt.Errorf("tokenization failed: %w", err)
	}
	fmt.Println("Patient tokens:")
	for _, t := range tokens {
		fmt.Println(t)
	}
	return nil
}

// tokenizeAllPatientsReturnTokens is like tokenizeAllPatients but returns the tokens as []string.
func tokenizeAllPatientsReturnTokens(cfg *config.Config, salt string) ([]string, error) {
	if cfg.Database.Type != "csv" {
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Database.Type)
	}
	csvPath := cfg.Database.Host
	if csvPath == "" {
		csvPath = cfg.Database.Table
	}
	if csvPath == "" && cfg.Database.Type == "csv" {
		csvPath = cfg.Database.Filename
	}
	if csvPath == "" {
		// Try to get from config field "filename"
		if v, ok := any(cfg.Database).(map[string]interface{})["filename"]; ok {
			csvPath, _ = v.(string)
		}
	}
	if csvPath == "" {
		return nil, fmt.Errorf("CSV filename not specified in config")
	}
	// Try all possible fields for filename
	if cfg.Database.Filename != "" {
		csvPath = cfg.Database.Filename
	}
	dbase, err := db.NewCSVDatabase(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV database: %w", err)
	}
	// List all keys
	keys, err := dbase.List(0, 1000000)
	if err != nil {
		return nil, fmt.Errorf("failed to list keys: %w", err)
	}
	// Read all records
	var records []map[string]interface{}
	for _, key := range keys {
		val, err := dbase.Get(key)
		if err != nil {
			continue
		}
		// Assume CSV value is a comma-separated string of fields
		fields := cfg.Database.Fields
		vals := strings.Split(val, ",")
		rec := make(map[string]interface{})
		for i, f := range fields {
			if i < len(vals) {
				rec[f] = vals[i]
			}
		}
		records = append(records, rec)
	}
	// Tokenize
	tokens, err := crypto.TokenizeRecords(records, cfg.Database.Fields, salt)
	if err != nil {
		return nil, fmt.Errorf("tokenization failed: %w", err)
	}
	return tokens, nil
}

// computeIntersection loads tokens from results.csv and returns the intersection with localTokens.
func computeIntersection(localTokens []string, resultsCSV string) ([]string, error) {
	file, err := os.Open(resultsCSV)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	lines := []string{}
	// Skip header
	scanner := bufio.NewScanner(file)
	if scanner.Scan() {
		// skip header
	}
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		if len(parts) > 0 {
			lines = append(lines, parts[0])
		}
	}
	// Compute intersection
	tokenSet := make(map[string]struct{}, len(localTokens))
	for _, t := range localTokens {
		tokenSet[t] = struct{}{}
	}
	var intersection []string
	for _, t := range lines {
		if _, ok := tokenSet[t]; ok {
			intersection = append(intersection, t)
		}
	}
	return intersection, nil
}

// Helper to get patient records for PSI token map
func getPatientRecords(cfg *config.Config) []map[string]interface{} {
	csvPath := cfg.Database.Host
	if csvPath == "" {
		csvPath = cfg.Database.Table
	}
	if csvPath == "" && cfg.Database.Type == "csv" {
		csvPath = cfg.Database.Filename
	}
	if csvPath == "" {
		if v, ok := any(cfg.Database).(map[string]interface{})["filename"]; ok {
			csvPath, _ = v.(string)
		}
	}
	if csvPath == "" && cfg.Database.Filename != "" {
		csvPath = cfg.Database.Filename
	}
	dbase, err := db.NewCSVDatabase(csvPath)
	if err != nil {
		return nil
	}
	keys, err := dbase.List(0, 1000000)
	if err != nil {
		return nil
	}
	var records []map[string]interface{}
	for _, key := range keys {
		val, err := dbase.Get(key)
		if err != nil {
			continue
		}
		fields := cfg.Database.Fields
		vals := strings.Split(val, ",")
		rec := make(map[string]interface{})
		for i, f := range fields {
			if i < len(vals) {
				rec[f] = vals[i]
			}
		}
		records = append(records, rec)
	}
	return records
}

// Helper to get curve-encoded tokens for PSI and intersection
func getCurveTokens(cfg *config.Config) []string {
	tokens := []string{}
	for _, rec := range getPatientRecords(cfg) {
		idVal, ok := rec["id"]
		if !ok || idVal == nil {
			continue
		}
		id, ok := idVal.(string)
		if !ok {
			id = fmt.Sprintf("%v", idVal)
		}
		P := crypto.HashToCurve(id)
		tokenHex := fmt.Sprintf("%x", P.Bytes())
		tokens = append(tokens, tokenHex)
	}
	return tokens
}

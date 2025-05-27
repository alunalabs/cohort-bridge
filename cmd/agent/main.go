package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

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

	fmt.Printf("Connecting to peer at %s...\n", peerAddr)
	peerPubKey, err := server.ExchangePublicKeysClient(peerAddr, cfg.PublicKey)
	if err != nil {
		fmt.Println("Error exchanging public keys:", err)
		os.Exit(1)
	}
	fmt.Println("Peer public key:", peerPubKey)

	sharedSalt := server.DeriveSharedSalt(cfg.PrivateKey, peerPubKey)
	fmt.Println("Derived shared salt:", sharedSalt)

	// --- Tokenize all patients using the shared salt ---
	if err := tokenizeAllPatients(cfg, sharedSalt); err != nil {
		fmt.Println("Error tokenizing patients:", err)
		os.Exit(1)
	}
	fmt.Println("Done. Exiting for debugging.")
}

func runAsReceiver(cfg *config.Config) {
	port := fmt.Sprintf("%d", cfg.ListenPort)

	fmt.Printf("Listening for peer connection on port %s...\n", port)
	peerPubKey, err := server.ExchangePublicKeysServer(port, cfg.PublicKey)
	if err != nil {
		fmt.Println("Error exchanging public keys:", err)
		os.Exit(1)
	}
	fmt.Println("Peer public key:", peerPubKey)

	sharedSalt := server.DeriveSharedSalt(cfg.PrivateKey, peerPubKey)
	fmt.Println("Derived shared salt:", sharedSalt)

	// --- Tokenize all patients using the shared salt ---
	if err := tokenizeAllPatients(cfg, sharedSalt); err != nil {
		fmt.Println("Error tokenizing patients:", err)
		os.Exit(1)
	}
	fmt.Println("Done. Exiting for debugging.")
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

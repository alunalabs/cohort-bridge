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
		RunAsSender(cfg)
	case "receive":
		RunAsReceiver(cfg)
	}
}

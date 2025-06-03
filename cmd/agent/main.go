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
	"github.com/auroradata-ai/cohort-bridge/internal/server"
	_ "github.com/lib/pq"
	"github.com/manifoldco/promptui"
)

func main() {
	fmt.Println("🩺 Welcome to CohortBridge CLI")

	// --- Add CLI flags ---
	mode := ""
	configFile := "config.yaml"
	flag.StringVar(&mode, "mode", "", "Mode: send, receive, or shutdown")
	flag.StringVar(&configFile, "config", "config.yaml", "Configuration file path")
	flag.Parse()

	if mode == "" {
		fmt.Println("Please specify mode: -mode=send, -mode=receive, or -mode=shutdown")
		os.Exit(1)
	}

	// --- Config file selection ---
	if configFile != "" {
		configPath := configFile
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
			server.RunAsSender(cfg)
		case "receive":
			server.RunAsReceiver(cfg)
		case "shutdown":
			server.SendShutdown(cfg)
		default:
			fmt.Printf("Unknown mode: %s\n", mode)
			os.Exit(1)
		}
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
		_, configFile, err := configPrompt.Run()
		if err != nil {
			fmt.Println("Prompt failed:", err)
			os.Exit(1)
		}

		cfg, err := config.Load(configFile)
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
			server.RunAsSender(cfg)
		case "receive":
			server.RunAsReceiver(cfg)
		case "shutdown":
			server.SendShutdown(cfg)
		default:
			fmt.Printf("Unknown mode: %s\n", mode)
			os.Exit(1)
		}
	}
}

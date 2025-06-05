package main

import (
	"bufio"
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
	configFile := ""
	flag.StringVar(&mode, "mode", "", "Mode: send, receive, shutdown, or view")
	flag.StringVar(&configFile, "config", "", "Configuration file path")
	flag.Parse()

	// If no mode is provided, use interactive selection
	if mode == "" {
		fmt.Println("\n📋 Interactive Mode Selection")
		fmt.Println("=============================")

		modePrompt := promptui.Select{
			Label: "Select operation mode",
			Items: []string{
				"send - Send data to receiver for matching",
				"receive - Receive data and perform matching",
				"shutdown - Send shutdown signal to receiver",
				"view - View previous match results",
			},
		}

		modeIndex, _, err := modePrompt.Run()
		if err != nil {
			fmt.Printf("Mode selection failed: %v\n", err)
			os.Exit(1)
		}

		switch modeIndex {
		case 0:
			mode = "send"
		case 1:
			mode = "receive"
		case 2:
			mode = "shutdown"
		case 3:
			mode = "view"
		default:
			fmt.Println("Invalid mode selected")
			os.Exit(1)
		}
	}

	// Config file selection (except for view mode which doesn't need config)
	if mode != "view" {
		if configFile == "" {
			// List .yaml config files in current directory
			yamlFiles := []string{}
			_ = filepath.Walk(".", func(path string, info fs.FileInfo, err error) error {
				if err == nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".yaml") {
					yamlFiles = append(yamlFiles, path)
				}
				return nil
			})
			if len(yamlFiles) == 0 {
				fmt.Println("❌ No .yaml config files found in current directory")
				fmt.Println("💡 Please create a config file based on config.example.yaml")
				os.Exit(1)
			}

			if len(yamlFiles) == 1 {
				configFile = yamlFiles[0]
				fmt.Printf("📁 Using config file: %s\n", configFile)
			} else {
				configPrompt := promptui.Select{
					Label: "Select config file",
					Items: yamlFiles,
				}
				var err error
				_, configFile, err = configPrompt.Run()
				if err != nil {
					fmt.Println("Config selection failed:", err)
					os.Exit(1)
				}
			}
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
	} else {
		// View mode doesn't need config
		viewMatchResults()
	}
}

// viewMatchResults lists and displays saved match result files
func viewMatchResults() {
	fmt.Println("📋 Viewing Match Results")
	fmt.Println("========================")

	// Find all match result CSV files in out/ directory
	resultFiles := []string{}
	_ = filepath.Walk("out", func(path string, info fs.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.Contains(info.Name(), "matches_") && strings.HasSuffix(info.Name(), ".csv") {
			resultFiles = append(resultFiles, path)
		}
		return nil
	})

	if len(resultFiles) == 0 {
		fmt.Println("❌ No match result files found in out/ directory.")
		fmt.Println("   Match result files are saved with names like: out/matches_YYYYMMDD_HHMMSS.csv")
		os.Exit(0)
	}

	fmt.Printf("📁 Found %d result file(s):\n", len(resultFiles))
	for i, file := range resultFiles {
		fmt.Printf("  %d. %s\n", i+1, file)
	}

	// Prompt user to select file
	filePrompt := promptui.Select{
		Label: "Select result file to view",
		Items: resultFiles,
	}
	_, selectedFile, err := filePrompt.Run()
	if err != nil {
		fmt.Println("Selection cancelled:", err)
		os.Exit(0)
	}

	// Display the selected file
	displayCSVFile(selectedFile)
}

// displayCSVFile reads and displays a CSV file in a formatted way
func displayCSVFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("❌ Failed to open file %s: %v\n", filename, err)
		return
	}
	defer file.Close()

	fmt.Printf("\n📄 Contents of %s:\n", filename)
	fmt.Println("=" + strings.Repeat("=", len(filename)+15))

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++

		if lineNum == 1 {
			// Format header
			fmt.Printf("%-12s %-12s %-12s %-15s %-18s %s\n",
				"Receiver_ID", "Sender_ID", "Match_Score", "Hamming_Dist", "Jaccard_Sim", "Is_Match")
			fmt.Println(strings.Repeat("-", 80))
		} else {
			// Format data rows
			parts := strings.Split(line, ",")
			if len(parts) >= 6 {
				fmt.Printf("%-12s %-12s %-12s %-15s %-18s %s\n",
					parts[0], parts[1], parts[2], parts[3], parts[4], parts[5])
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("❌ Error reading file: %v\n", err)
	}

	fmt.Printf("\n📊 Total matches: %d\n", lineNum-1)
}

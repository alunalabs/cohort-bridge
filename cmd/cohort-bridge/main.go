package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/auroradata-ai/cohort-bridge/internal/config"
	"github.com/manifoldco/promptui"
)

func main() {
	// ASCII art header
	fmt.Println("🤖 CohortBridge - PPRL Orchestrator")
	fmt.Println("====================================")
	fmt.Println("Privacy-Preserving Record Linkage System")
	fmt.Println()

	// Check for subcommands
	if len(os.Args) > 1 {
		subcommand := os.Args[1]
		switch subcommand {
		case "tokenize":
			runTokenizeCommand(os.Args[2:])
			return
		case "intersect":
			runIntersectCommand(os.Args[2:])
			return
		case "send":
			runSendCommand(os.Args[2:])
			return
		case "validate":
			runValidateCommand(os.Args[2:])
			return
		case "orchestrate":
			// Continue with orchestrate functionality using remaining args
			os.Args = os.Args[1:]
		case "-help", "--help", "help":
			showMainHelp()
			return
		default:
			// Check if it's a flag - if so, use legacy agent mode
			if strings.HasPrefix(subcommand, "-") {
				runLegacyMode()
				return
			}
			fmt.Printf("Unknown subcommand: %s\n", subcommand)
			showMainHelp()
			os.Exit(1)
		}
	}

	// No arguments or 'orchestrate' subcommand - run interactive mode
	runInteractiveMode()
}

func showMainHelp() {
	fmt.Println("CohortBridge - Privacy-Preserving Record Linkage System")
	fmt.Println("========================================================")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  cohort-bridge [SUBCOMMAND] [OPTIONS]")
	fmt.Println("  cohort-bridge                              # Interactive mode")
	fmt.Println()
	fmt.Println("SUBCOMMANDS:")
	fmt.Println("  tokenize     Convert raw data to privacy-preserving tokens")
	fmt.Println("  intersect    Find matches between tokenized datasets")
	fmt.Println("  send         Send intersection results to another party")
	fmt.Println("  validate     End-to-end validation against ground truth")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Interactive mode")
	fmt.Println("  cohort-bridge")
	fmt.Println()
	fmt.Println("  # Tokenize data")
	fmt.Println("  cohort-bridge tokenize -input data.csv -output tokens.csv")
	fmt.Println()
	fmt.Println("  # Find intersections")
	fmt.Println("  cohort-bridge intersect -dataset1 tokens1.csv -dataset2 tokens2.csv")
	fmt.Println()
	fmt.Println("  # Complete workflow (legacy mode)")
	fmt.Println("  cohort-bridge -mode=orchestrate -config=config.yaml")
	fmt.Println()
	fmt.Println("Get help for specific subcommands:")
	fmt.Println("  cohort-bridge tokenize -help")
	fmt.Println("  cohort-bridge intersect -help")
	fmt.Println("  cohort-bridge send -help")
	fmt.Println("  cohort-bridge validate -help")
}

func runLegacyMode() {
	// Keep the old command-line interface for backwards compatibility
	var (
		mode       = flag.String("mode", "", "Mode: sender, receiver, or orchestrate")
		configFile = flag.String("config", "", "Configuration file path")
		workflow   = flag.Bool("workflow", false, "Run complete PPRL workflow (orchestrate mode)")
	)
	flag.Parse()

	if *workflow {
		*mode = "orchestrate"
	}

	if *mode == "" || *configFile == "" {
		fmt.Println("Usage: cohort-bridge -mode=<sender|receiver|orchestrate> -config=<config.yaml>")
		fmt.Println("   or: cohort-bridge (for interactive mode)")
		os.Exit(1)
	}

	runWithConfig(*mode, *configFile)
}

func runInteractiveMode() {
	// Welcome message
	fmt.Println("🚀 Welcome to CohortBridge Interactive Mode")
	fmt.Println("This tool will guide you through privacy-preserving record linkage.")
	fmt.Println()

	// Step 1: Select operation mode
	modePrompt := promptui.Select{
		Label: "Select Operation Mode",
		Items: []string{
			"🛡️  Receiver - Listen for sender connections (Party A)",
			"📡 Sender - Connect to receiver (Party B)",
			"🔄 Orchestrate - Complete local workflow (single machine)",
		},
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}:",
			Active:   "▶ {{ . | cyan }}",
			Inactive: "  {{ . | white }}",
			Selected: "✓ {{ . | green }}",
		},
		Size: 3,
	}

	modeIndex, _, err := modePrompt.Run()
	if err != nil {
		fmt.Printf("❌ Selection cancelled: %v\n", err)
		os.Exit(1)
	}

	var mode string
	switch modeIndex {
	case 0:
		mode = "receiver"
	case 1:
		mode = "sender"
	case 2:
		mode = "orchestrate"
	}

	fmt.Printf("\n🎯 Selected Mode: %s\n\n", strings.Title(mode))

	// Step 2: Select configuration file
	configFiles := findConfigFiles()
	if len(configFiles) == 0 {
		// Manual input if no configs found
		prompt := promptui.Prompt{
			Label: "No config files found. Enter config file path",
			Validate: func(input string) error {
				if strings.TrimSpace(input) == "" {
					return fmt.Errorf("config file path cannot be empty")
				}
				if _, err := os.Stat(input); os.IsNotExist(err) {
					return fmt.Errorf("file does not exist: %s", input)
				}
				return nil
			},
		}

		configFile, err := prompt.Run()
		if err != nil {
			fmt.Printf("❌ Input cancelled: %v\n", err)
			os.Exit(1)
		}

		runWithConfig(mode, configFile)
		return
	}

	// Add descriptions to config files
	var configOptions []string
	for _, file := range configFiles {
		description := getConfigDescription(file)
		configOptions = append(configOptions, fmt.Sprintf("%s - %s", file, description))
	}

	configPrompt := promptui.Select{
		Label: "Select Configuration File",
		Items: configOptions,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}:",
			Active:   "▶ {{ . | cyan }}",
			Inactive: "  {{ . | white }}",
			Selected: "✓ {{ . | green }}",
		},
		Size: 10,
	}

	configIndex, _, err := configPrompt.Run()
	if err != nil {
		fmt.Printf("❌ Selection cancelled: %v\n", err)
		os.Exit(1)
	}

	configFile := configFiles[configIndex]
	fmt.Printf("\n📁 Selected Config: %s\n", configFile)

	// Step 3: Show operation summary and confirm
	fmt.Println("\n📋 Operation Summary:")
	fmt.Printf("   Mode: %s\n", strings.Title(mode))
	fmt.Printf("   Config: %s\n", configFile)
	fmt.Println()

	confirmPrompt := promptui.Select{
		Label: "Ready to proceed?",
		Items: []string{"✅ Yes, start the operation", "❌ No, exit"},
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}:",
			Active:   "▶ {{ . | cyan }}",
			Inactive: "  {{ . | white }}",
			Selected: "✓ {{ . | green }}",
		},
	}

	confirmIndex, _, err := confirmPrompt.Run()
	if err != nil || confirmIndex != 0 {
		fmt.Println("\n👋 Operation cancelled. Goodbye!")
		os.Exit(0)
	}

	fmt.Println("\n🔥 Starting operation...\n")
	runWithConfig(mode, configFile)
}

func runWithConfig(mode, configFile string) {
	fmt.Printf("🎯 Mode: %s\n", mode)
	fmt.Printf("📁 Config: %s\n\n", configFile)

	// Load config to validate
	cfg, err := config.Load(configFile)
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Run based on mode - these will be implemented in workflows.go
	switch mode {
	case "sender":
		runSenderWorkflow(cfg)
	case "receiver":
		runReceiverWorkflow(cfg)
	case "orchestrate":
		runOrchestrationWorkflow(cfg)
	default:
		fmt.Printf("Unknown mode: %s\n", mode)
		os.Exit(1)
	}

	fmt.Printf("\n✅ %s workflow completed successfully!\n", strings.Title(mode))
}

// Helper functions
func findConfigFiles() []string {
	var configs []string
	patterns := []string{"*.yaml", "*.yml", "config*.yaml", "config*.yml"}

	for _, pattern := range patterns {
		if matches, err := filepath.Glob(pattern); err == nil {
			for _, match := range matches {
				// Skip if already in list
				found := false
				for _, existing := range configs {
					if existing == match {
						found = true
						break
					}
				}
				if !found {
					configs = append(configs, match)
				}
			}
		}
	}

	return configs
}

func getConfigDescription(filename string) string {
	// Try to give meaningful descriptions based on filename patterns
	lower := strings.ToLower(filename)

	if strings.Contains(lower, "_a") || strings.Contains(lower, "party_a") {
		return "Party A configuration"
	}
	if strings.Contains(lower, "_b") || strings.Contains(lower, "party_b") {
		return "Party B configuration"
	}
	if strings.Contains(lower, "receiver") {
		return "Receiver configuration"
	}
	if strings.Contains(lower, "sender") {
		return "Sender configuration"
	}
	if strings.Contains(lower, "postgres") {
		return "PostgreSQL database configuration"
	}
	if strings.Contains(lower, "secure") {
		return "Secure/encrypted configuration"
	}
	if strings.Contains(lower, "tokenized") {
		return "Pre-tokenized data configuration"
	}
	if strings.Contains(lower, "example") {
		return "Example configuration"
	}

	return "General configuration"
}

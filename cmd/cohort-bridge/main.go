package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/manifoldco/promptui"
)

func main() {
	// Print banner
	fmt.Println("🤖 CohortBridge - PPRL Orchestrator")
	fmt.Println("====================================")
	fmt.Println("Privacy-Preserving Record Linkage System")
	fmt.Println()

	// Handle command line arguments
	if len(os.Args) > 1 {
		// Handle subcommands
		subcommand := os.Args[1]
		args := os.Args[2:]

		switch subcommand {
		case "tokenize":
			runTokenizeCommand(args)
		case "decrypt":
			runDecryptCommand(args)
		case "intersect":
			runIntersectCommand(args)
		case "validate":
			runValidateCommand(args)
		case "workflows":
			runWorkflowsCommand(args)
		case "-help", "--help", "help":
			showMainHelp()
		case "-version", "--version", "version":
			showVersion()
		default:
			fmt.Printf("❌ Unknown subcommand: %s\n\n", subcommand)
			showMainHelp()
			os.Exit(1)
		}
		return
	}

	// Interactive mode - no arguments provided
	runInteractiveMode()
}

func runInteractiveMode() {
	fmt.Println("🎯 Interactive Mode")

	options := []string{
		"🔐 Tokenize - Convert PHI data to privacy-preserving tokens",
		"🔓 Decrypt - Decrypt encrypted tokenized files",
		"🔍 Intersect - Find matches between tokenized datasets",
		"🔬 Validate - Test results against ground truth",
		"⚙️  Workflows - Orchestrate complex PPRL operations",
		"❓ Help - Show detailed help information",
		"🚪 Exit",
	}

	choice := promptForChoice("Choose what you'd like to do:", options)

	switch choice {
	case 0: // Tokenize
		runTokenizeCommand([]string{"-interactive"})
	case 1: // Decrypt
		runDecryptCommand([]string{"-interactive"})
	case 2: // Intersect
		runIntersectCommand([]string{"-interactive"})
	case 3: // Validate
		runValidateCommand([]string{"-interactive"})
	case 4: // Workflows
		runWorkflowsCommand([]string{"-interactive"})
	case 5: // Help
		showMainHelp()
	case 6: // Exit
		fmt.Println("👋 Goodbye!")
		os.Exit(0)
	}
}

// Simple text input with arrow key support for text editing
func promptForInput(message, defaultValue string) string {
	var prompt string
	if defaultValue != "" {
		prompt = fmt.Sprintf("%s (default: %s): ", message, defaultValue)
	} else {
		prompt = fmt.Sprintf("%s: ", message)
	}
	fmt.Fprint(os.Stderr, prompt)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		// Handle EOF or other read errors
		if defaultValue != "" {
			fmt.Printf("\nNo input received, using default: %s\n", defaultValue)
			return defaultValue
		}
		fmt.Printf("\nNo input received and no default available\n")
		return ""
	}

	input = strings.TrimSpace(input)
	if input == "" && defaultValue != "" {
		return defaultValue
	}
	return input
}

// Use promptui for menu selection with arrow keys
func promptForChoice(message string, options []string) int {
	prompt := promptui.Select{
		Label: message,
		Items: options,
		Size:  10, // Show up to 10 items at once
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "▶ {{ . | cyan }}",
			Inactive: "  {{ . }}",
			Selected: "✓ {{ . | green }}",
		},
	}

	index, _, err := prompt.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	return index
}

func showMainHelp() {
	fmt.Println("🤖 CohortBridge - Privacy-Preserving Record Linkage")
	fmt.Println("====================================================")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  cohort-bridge                     # Interactive mode")
	fmt.Println("  cohort-bridge <subcommand>        # Direct subcommand")
	fmt.Println("  cohort-bridge -mode=<mode>        # Legacy mode")
	fmt.Println()
	fmt.Println("SUBCOMMANDS:")
	fmt.Println("  tokenize    🔐 Convert PHI data to privacy-preserving tokens")
	fmt.Println("  decrypt     🔓 Decrypt encrypted tokenized files")
	fmt.Println("  intersect   🔍 Find matches between tokenized datasets")
	fmt.Println("  send        📡 Network operations for secure communication")
	fmt.Println("  validate    🔬 Test results against ground truth")
	fmt.Println("  workflows   ⚙️  Orchestrate complex PPRL operations")
	fmt.Println()
	fmt.Println()
	fmt.Println("GLOBAL OPTIONS:")
	fmt.Println("  -help, --help    Show this help message")
	fmt.Println("  -version         Show version information")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Interactive mode")
	fmt.Println("  cohort-bridge")
	fmt.Println()
	fmt.Println("  # Direct subcommands")
	fmt.Println("  cohort-bridge tokenize -input data.csv -output tokens.csv.enc")
	fmt.Println("  cohort-bridge decrypt -input tokens.csv.enc -key tokens.key")
	fmt.Println("  cohort-bridge intersect -dataset1 tokens1.csv -dataset2 tokens2.csv")
	fmt.Println("  cohort-bridge validate -config1 config_a.yaml -config2 config_b.yaml")
	fmt.Println()
	fmt.Println("  # Legacy mode")
	fmt.Println("  cohort-bridge -mode=sender -config=config.yaml")
	fmt.Println()
	fmt.Println("For detailed help on any subcommand, use:")
	fmt.Println("  cohort-bridge <subcommand> -help")
}

func showVersion() {
	fmt.Println("🤖 CohortBridge v1.0.0")
	fmt.Println("Privacy-Preserving Record Linkage System")
	fmt.Println("Built with ❤️  for secure healthcare data collaboration")
}

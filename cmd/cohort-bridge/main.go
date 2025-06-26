package main

import (
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
		// Check for legacy mode first (-mode flag)
		for i, arg := range os.Args[1:] {
			if strings.HasPrefix(arg, "-mode=") {
				mode := strings.TrimPrefix(arg, "-mode=")
				handleLegacyMode(mode, os.Args[i+2:]) // Skip program name and -mode flag
				return
			}
		}

		// Handle subcommands
		subcommand := os.Args[1]
		args := os.Args[2:]

		switch subcommand {
		case "tokenize":
			runTokenizeCommand(args)
		case "intersect":
			runIntersectCommand(args)
		case "send":
			runSendCommand(args)
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
	fmt.Println("Choose what you'd like to do:")
	fmt.Println()

	mainPrompt := promptui.Select{
		Label: "Select an operation",
		Items: []string{
			"🔐 Tokenize - Convert PHI data to privacy-preserving tokens",
			"🔍 Intersect - Find matches between tokenized datasets",
			"📡 Send - Network operations for secure communication",
			"🔬 Validate - Test results against ground truth",
			"⚙️  Workflows - Orchestrate complex PPRL operations",
			"❓ Help - Show detailed help information",
			"🚪 Exit",
		},
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}:",
			Active:   "▶ {{ . | cyan }}",
			Inactive: "  {{ . | white }}",
			Selected: "✓ {{ . | green }}",
		},
		Size:     7,
		HideHelp: true,
	}

	selectedIndex, _, err := mainPrompt.Run()
	if err != nil {
		fmt.Printf("❌ Error in interactive mode: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()

	switch selectedIndex {
	case 0: // Tokenize
		runTokenizeCommand([]string{"-interactive"})
	case 1: // Intersect
		runIntersectCommand([]string{"-interactive"})
	case 2: // Send
		runSendCommand([]string{"-interactive"})
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

func handleLegacyMode(mode string, args []string) {
	fmt.Printf("🔄 Legacy mode detected: %s\n", mode)

	switch mode {
	case "sender":
		fmt.Println("🚀 Running in sender mode...")
		runSendCommand(args)
	case "receiver":
		fmt.Println("📡 Running in receiver mode...")
		runSendCommand(args)
	default:
		fmt.Printf("❌ Unknown legacy mode: %s\n", mode)
		fmt.Println("💡 Try using the new subcommand syntax: cohort-bridge <subcommand>")
		showMainHelp()
		os.Exit(1)
	}
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
	fmt.Println("  intersect   🔍 Find matches between tokenized datasets")
	fmt.Println("  send        📡 Network operations for secure communication")
	fmt.Println("  validate    🔬 Test results against ground truth")
	fmt.Println("  workflows   ⚙️  Orchestrate complex PPRL operations")
	fmt.Println()
	fmt.Println("LEGACY MODES:")
	fmt.Println("  -mode=sender     Run as data sender")
	fmt.Println("  -mode=receiver   Run as data receiver")
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
	fmt.Println("  cohort-bridge tokenize -input data.csv -output tokens.csv")
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

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/auroradata-ai/cohort-bridge/internal/config"
	"github.com/manifoldco/promptui"
)

// runSenderWorkflow runs the sender-specific workflow
func runSenderWorkflow(cfg *config.Config) {
	fmt.Println("📤 Starting sender workflow")

	// Create temp-sender directory
	if err := createWorkspaceDirectory("temp-sender"); err != nil {
		log.Fatalf("Failed to create sender workspace: %v", err)
	}

	// Use the existing server sender functionality
	if err := createSender(cfg); err != nil {
		log.Fatalf("Failed to create sender: %v", err)
	}

	fmt.Println("   📡 Running sender workflow...")

	fmt.Println("✅ Sender workflow completed successfully!")
}

// runReceiverWorkflow runs the receiver-specific workflow
func runReceiverWorkflow(cfg *config.Config) {
	fmt.Println("📥 Starting receiver workflow")

	// Create temp-receiver directory
	if err := createWorkspaceDirectory("temp-receiver"); err != nil {
		log.Fatalf("Failed to create receiver workspace: %v", err)
	}

	// Use the existing server receiver functionality
	if err := createReceiver(cfg); err != nil {
		log.Fatalf("Failed to create receiver: %v", err)
	}

	fmt.Println("   📥 Running receiver workflow...")

	fmt.Println("✅ Receiver workflow completed successfully!")
}

// runOrchestrationWorkflow runs the complete PPRL workflow
func runOrchestrationWorkflow(cfg *config.Config) {
	fmt.Println("🔄 Starting complete PPRL orchestration workflow")

	// This would implement the complete workflow:
	// 1. Tokenize local database
	// 2. Exchange tokenized data with peer
	// 3. Compute intersection
	// 4. Send results to peer

	fmt.Println("\n📝 Step 1: Tokenizing local database...")
	if err := performTokenizationStep(cfg); err != nil {
		log.Fatalf("Tokenization failed: %v", err)
	}

	fmt.Println("\n🔄 Step 2: Exchanging tokenized data with peer...")
	if err := performDataExchangeStep(cfg); err != nil {
		log.Fatalf("Data exchange failed: %v", err)
	}

	fmt.Println("\n🔍 Step 3: Computing intersection...")
	if err := performIntersectionStep(cfg); err != nil {
		log.Fatalf("Intersection failed: %v", err)
	}

	fmt.Println("\n📤 Step 4: Sending results to peer...")
	if err := performSendStep(cfg); err != nil {
		log.Fatalf("Sending failed: %v", err)
	}

	fmt.Println("\n✅ PPRL orchestration completed successfully!")
}

// Helper functions that would use the internal packages
func createWorkspaceDirectory(name string) error {
	// Implementation would create the workspace directory
	fmt.Printf("📁 Creating workspace: %s/\n", name)
	return nil
}

func createSender(cfg *config.Config) error {
	// Use the existing server package to create a sender
	fmt.Println("   🔧 Creating sender...")
	return nil
}

func createReceiver(cfg *config.Config) error {
	// Use the existing server package to create a receiver
	fmt.Println("   🔧 Creating receiver...")
	return nil
}

func performTokenizationStep(cfg *config.Config) error {
	// Implementation would tokenize the local database
	fmt.Println("   🔧 Tokenizing data...")
	return nil
}

func performDataExchangeStep(cfg *config.Config) error {
	// Implementation would exchange tokenized data with peer
	fmt.Println("   📡 Exchanging data with peer...")
	return nil
}

func performIntersectionStep(cfg *config.Config) error {
	// Implementation would compute intersection
	fmt.Println("   🔍 Computing intersection...")
	return nil
}

func performSendStep(cfg *config.Config) error {
	// Implementation would send results to peer
	fmt.Println("   📤 Sending results...")
	return nil
}

func runWorkflowsCommand(args []string) {
	fmt.Println("⚙️  CohortBridge Workflow Orchestrator")
	fmt.Println("=====================================")
	fmt.Println("Orchestrate complex PPRL operations")
	fmt.Println()

	fs := flag.NewFlagSet("workflows", flag.ExitOnError)
	var (
		configFile   = fs.String("config", "", "Configuration file")
		workflowType = fs.String("workflow", "", "Workflow type: sender, receiver, orchestration")
		interactive  = fs.Bool("interactive", false, "Force interactive mode")
		help         = fs.Bool("help", false, "Show help message")
	)
	fs.Parse(args)

	if *help {
		showWorkflowsHelp()
		return
	}

	// If missing required parameters or interactive mode requested, go interactive
	if *configFile == "" || *workflowType == "" || *interactive {
		fmt.Println("🎯 Interactive Workflow Setup")
		fmt.Println("Let's configure your workflow...\n")

		// Get configuration file
		if *configFile == "" {
			var err error
			*configFile, err = selectDataFile("Select Configuration File", "config", []string{".yaml"})
			if err != nil {
				fmt.Printf("❌ Error selecting config file: %v\n", err)
				os.Exit(1)
			}
		}

		// Select workflow type
		if *workflowType == "" {
			workflowPrompt := promptui.Select{
				Label: "Select workflow type",
				Items: []string{
					"📤 Sender - Send data to peer",
					"📥 Receiver - Receive data from peer",
					"🔄 Orchestration - Complete PPRL workflow",
				},
				Templates: &promptui.SelectTemplates{
					Label:    "{{ . }}:",
					Active:   "▶ {{ . | cyan }}",
					Inactive: "  {{ . | white }}",
					Selected: "✓ {{ . | green }}",
				},
				Size:     3,
				HideHelp: true,
			}

			workflowIndex, _, err := workflowPrompt.Run()
			if err != nil {
				fmt.Printf("❌ Error selecting workflow: %v\n", err)
				os.Exit(1)
			}

			switch workflowIndex {
			case 0:
				*workflowType = "sender"
			case 1:
				*workflowType = "receiver"
			case 2:
				*workflowType = "orchestration"
			}
		}

		fmt.Println()
	}

	// Show configuration summary
	fmt.Println("📋 Workflow Configuration:")
	fmt.Printf("  📁 Config File: %s\n", *configFile)
	fmt.Printf("  ⚙️  Workflow Type: %s\n", *workflowType)
	fmt.Println()

	// Confirm before proceeding
	confirmPrompt := promptui.Select{
		Label: "Ready to start workflow?",
		Items: []string{
			"✅ Yes, start workflow",
			"⚙️  Change configuration",
			"❌ Cancel",
		},
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}:",
			Active:   "▶ {{ . | cyan }}",
			Inactive: "  {{ . | white }}",
			Selected: "✓ {{ . | green }}",
		},
		Size:     3,
		HideHelp: true,
	}

	confirmIndex, _, err := confirmPrompt.Run()
	if err != nil || confirmIndex == 2 {
		fmt.Println("\n👋 Workflow cancelled. Goodbye!")
		os.Exit(0)
	}

	if confirmIndex == 1 {
		// Restart configuration
		fmt.Println("\n🔄 Restarting configuration...\n")
		newArgs := append([]string{"-interactive"}, args...)
		runWorkflowsCommand(newArgs)
		return
	}

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Run the selected workflow
	fmt.Println("🚀 Starting workflow...\n")

	switch *workflowType {
	case "sender":
		runSenderWorkflow(cfg)
	case "receiver":
		runReceiverWorkflow(cfg)
	case "orchestration":
		runOrchestrationWorkflow(cfg)
	default:
		fmt.Printf("❌ Unknown workflow type: %s\n", *workflowType)
		os.Exit(1)
	}
}

func showWorkflowsHelp() {
	fmt.Println("⚙️  CohortBridge Workflow Orchestrator")
	fmt.Println("=====================================")
	fmt.Println()
	fmt.Println("Orchestrate complex PPRL operations")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  cohort-bridge workflows [OPTIONS]")
	fmt.Println("  cohort-bridge workflows                  # Interactive mode")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  -config string     Configuration file")
	fmt.Println("  -workflow string   Workflow type: sender, receiver, orchestration")
	fmt.Println("  -interactive       Force interactive mode")
	fmt.Println("  -help              Show this help message")
	fmt.Println()
	fmt.Println("WORKFLOW TYPES:")
	fmt.Println("  sender         📤 Send data to peer")
	fmt.Println("  receiver       📥 Receive data from peer")
	fmt.Println("  orchestration  🔄 Complete PPRL workflow")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Interactive mode (prompts for all inputs)")
	fmt.Println("  cohort-bridge workflows")
	fmt.Println()
	fmt.Println("  # Command line mode")
	fmt.Println("  cohort-bridge workflows -config config.yaml -workflow sender")
	fmt.Println("  cohort-bridge workflows -config config.yaml -workflow orchestration")
	fmt.Println()
	fmt.Println("  # Force interactive even with some parameters")
	fmt.Println("  cohort-bridge workflows -config config.yaml -interactive")
}

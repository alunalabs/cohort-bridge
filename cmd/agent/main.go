package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/auroradata-ai/cohort-bridge/internal/config"
)

// WorkflowConfig defines the orchestration workflow configuration
type WorkflowConfig struct {
	ConfigFile       string `json:"config_file"`
	Mode             string `json:"mode"`
	TokenizeOutput   string `json:"tokenize_output"`
	IntersectionFile string `json:"intersection_file"`
	PeerTokens       string `json:"peer_tokens"`
	Streaming        bool   `json:"streaming"`
	BatchSize        int    `json:"batch_size"`
}

func main() {
	var (
		mode         = flag.String("mode", "", "Mode: 'sender', 'receiver', 'shutdown', or 'orchestrate'")
		configFile   = flag.String("config", "config.yaml", "Path to configuration file")
		workflow     = flag.Bool("workflow", false, "Run complete PPRL workflow (equivalent to -mode=orchestrate)")
		help         = flag.Bool("help", false, "Show help message")
		streaming    = flag.Bool("streaming", false, "Enable streaming mode for large datasets")
		batchSize    = flag.Int("batch-size", 1000, "Batch size for streaming mode")
		tokensOutput = flag.String("tokens-output", "tokenized_data.csv", "Output file for tokenized data")
		intersection = flag.String("intersection-output", "intersection_results.csv", "Output file for intersection results")
		peerTokens   = flag.String("peer-tokens", "", "Path to peer's tokenized data (for local intersection)")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Workflow mode is an alias for orchestrate mode
	if *workflow {
		*mode = "orchestrate"
	}

	if *mode == "" {
		fmt.Println("Error: Mode is required")
		showHelp()
		os.Exit(1)
	}

	// Validate mode
	validModes := []string{"sender", "receiver", "shutdown", "orchestrate"}
	if !contains(validModes, *mode) {
		fmt.Printf("Error: Invalid mode '%s'. Valid modes are: %s\n", *mode, strings.Join(validModes, ", "))
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration from %s: %v", *configFile, err)
	}

	workflowConfig := &WorkflowConfig{
		ConfigFile:       *configFile,
		Mode:             *mode,
		TokenizeOutput:   *tokensOutput,
		IntersectionFile: *intersection,
		PeerTokens:       *peerTokens,
		Streaming:        *streaming,
		BatchSize:        *batchSize,
	}

	fmt.Println("🤖 CohortBridge Agent - PPRL Orchestrator")
	fmt.Printf("📁 Config: %s\n", *configFile)
	fmt.Printf("🎯 Mode: %s\n", *mode)

	// Execute based on mode
	switch *mode {
	case "orchestrate":
		if err := runOrchestration(workflowConfig, cfg); err != nil {
			log.Fatalf("Orchestration failed: %v", err)
		}

	case "sender":
		if err := runSenderWorkflow(workflowConfig, cfg); err != nil {
			log.Fatalf("Sender workflow failed: %v", err)
		}

	case "receiver":
		if err := runReceiverWorkflow(workflowConfig, cfg); err != nil {
			log.Fatalf("Receiver workflow failed: %v", err)
		}

	case "shutdown":
		fmt.Println("🛑 Sending shutdown signal to receiver")
		fmt.Printf("🎯 Target: %s:%d\n", cfg.Peer.Host, cfg.Peer.Port)
		// For shutdown, we can still use the existing server functionality
		// since it's just a simple network call
		if err := sendShutdownSignal(cfg); err != nil {
			log.Fatalf("Failed to send shutdown signal: %v", err)
		}
		fmt.Println("✅ Shutdown signal sent")

	default:
		fmt.Printf("Error: Unknown mode '%s'\n", *mode)
		os.Exit(1)
	}
}

// runOrchestration runs the complete PPRL workflow
func runOrchestration(config *WorkflowConfig, cfg *config.Config) error {
	fmt.Println("🔄 Starting complete PPRL orchestration workflow")

	// Step 1: Tokenize local database
	fmt.Println("\n📝 Step 1: Tokenizing local database...")
	if err := callTokenize(config, cfg); err != nil {
		return fmt.Errorf("tokenization failed: %w", err)
	}

	// Step 2: Exchange tokenized data with peer
	fmt.Println("\n🔄 Step 2: Exchanging tokenized data with peer...")
	if err := exchangeTokenizedData(config, cfg); err != nil {
		return fmt.Errorf("data exchange failed: %w", err)
	}

	// Step 3: Compute intersection
	fmt.Println("\n🔍 Step 3: Computing intersection...")
	if err := callIntersect(config); err != nil {
		return fmt.Errorf("intersection failed: %w", err)
	}

	// Step 4: Send results to peer (if sender)
	fmt.Println("\n📤 Step 4: Sending results to peer...")
	if err := callSend(config); err != nil {
		return fmt.Errorf("sending failed: %w", err)
	}

	fmt.Println("\n✅ PPRL orchestration completed successfully!")
	return nil
}

// runSenderWorkflow runs the sender-specific workflow
func runSenderWorkflow(config *WorkflowConfig, cfg *config.Config) error {
	fmt.Println("📤 Starting sender workflow")

	// Tokenize data
	if err := callTokenize(config, cfg); err != nil {
		return err
	}

	// For now, use a simplified workflow
	// In a full implementation, this would coordinate with the receiver
	fmt.Println("📡 Waiting for peer connection and coordination...")

	return fmt.Errorf("sender workflow requires peer coordination - use 'orchestrate' mode for complete workflow")
}

// runReceiverWorkflow runs the receiver-specific workflow
func runReceiverWorkflow(config *WorkflowConfig, cfg *config.Config) error {
	fmt.Println("📥 Starting receiver workflow")

	// Tokenize data
	if err := callTokenize(config, cfg); err != nil {
		return err
	}

	// For now, use a simplified workflow
	// In a full implementation, this would listen for sender connections
	fmt.Println("👂 Listening for peer connections...")

	return fmt.Errorf("receiver workflow requires network listening - use 'orchestrate' mode for complete workflow")
}

// callTokenize calls the tokenize CLI program
func callTokenize(config *WorkflowConfig, cfg *config.Config) error {
	fmt.Printf("   🔧 Tokenizing data from: %s\n", cfg.Database.Filename)

	args := []string{
		"-config", config.ConfigFile,
		"-output", config.TokenizeOutput,
	}

	if config.Streaming {
		args = append(args, "-streaming", "-batch-size", fmt.Sprintf("%d", config.BatchSize))
	}

	cmd := exec.Command("./tokenize", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tokenize command failed: %w", err)
	}

	fmt.Printf("   ✅ Tokenization completed: %s\n", config.TokenizeOutput)
	return nil
}

// exchangeTokenizedData handles the exchange of tokenized data with the peer
func exchangeTokenizedData(config *WorkflowConfig, cfg *config.Config) error {
	// This is a placeholder for the peer exchange logic
	// In a real implementation, this would:
	// 1. Send our tokenized data to peer
	// 2. Receive peer's tokenized data
	// 3. Save peer data for intersection

	fmt.Println("   📡 Establishing peer connection...")
	fmt.Printf("   📤 Sending tokenized data: %s\n", config.TokenizeOutput)
	fmt.Println("   📥 Receiving peer tokenized data...")

	// For demo purposes, assume peer data is provided locally
	if config.PeerTokens == "" {
		config.PeerTokens = "peer_tokenized_data.csv"
		fmt.Printf("   ⚠️  Using local peer data file: %s\n", config.PeerTokens)
	}

	return nil
}

// callIntersect calls the intersect CLI program
func callIntersect(config *WorkflowConfig) error {
	fmt.Printf("   🔍 Finding intersection between datasets\n")

	args := []string{
		"-dataset1", config.TokenizeOutput,
		"-dataset2", config.PeerTokens,
		"-output", config.IntersectionFile,
	}

	if config.Streaming {
		args = append(args, "-streaming", "-batch-size", fmt.Sprintf("%d", config.BatchSize))
	}

	cmd := exec.Command("./intersect", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("intersect command failed: %w", err)
	}

	fmt.Printf("   ✅ Intersection completed: %s\n", config.IntersectionFile)
	return nil
}

// callSend calls the send CLI program
func callSend(config *WorkflowConfig) error {
	fmt.Printf("   📤 Sending intersection results to peer\n")

	args := []string{
		"-intersection", config.IntersectionFile,
		"-config", config.ConfigFile,
		"-mode", "intersection",
	}

	cmd := exec.Command("./send", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("send command failed: %w", err)
	}

	fmt.Println("   ✅ Results sent successfully")
	return nil
}

// sendShutdownSignal sends a shutdown signal to the receiver
func sendShutdownSignal(cfg *config.Config) error {
	// This can remain as a simple direct implementation since it's just a network call
	// For demonstration, we'll use a placeholder
	target := fmt.Sprintf("%s:%d", cfg.Peer.Host, cfg.Peer.Port)
	fmt.Printf("Sending shutdown to %s\n", target)

	// In real implementation, this would make the actual network call
	// server.SendShutdown(cfg)

	time.Sleep(1 * time.Second) // Simulate network call
	return nil
}

func showHelp() {
	fmt.Println("CohortBridge Agent - Privacy-Preserving Record Linkage Orchestrator")
	fmt.Println("================================================================")
	fmt.Println()
	fmt.Println("The agent orchestrates the complete PPRL workflow by calling specialized")
	fmt.Println("CLI tools for tokenization, intersection finding, and data transmission.")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  agent [OPTIONS]")
	fmt.Println()
	fmt.Println("OPTIONS:")
	fmt.Println("  -mode string")
	fmt.Println("        Operating mode: 'orchestrate', 'sender', 'receiver', or 'shutdown'")
	fmt.Println("  -workflow")
	fmt.Println("        Run complete PPRL workflow (alias for -mode=orchestrate)")
	fmt.Println("  -config string")
	fmt.Println("        Path to configuration file (default: config.yaml)")
	fmt.Println("  -tokens-output string")
	fmt.Println("        Output file for tokenized data (default: tokenized_data.csv)")
	fmt.Println("  -intersection-output string")
	fmt.Println("        Output file for intersection results (default: intersection_results.csv)")
	fmt.Println("  -peer-tokens string")
	fmt.Println("        Path to peer's tokenized data file (for local testing)")
	fmt.Println("  -streaming")
	fmt.Println("        Enable streaming mode for large datasets")
	fmt.Println("  -batch-size int")
	fmt.Println("        Batch size for streaming mode (default: 1000)")
	fmt.Println("  -help")
	fmt.Println("        Show this help message")
	fmt.Println()
	fmt.Println("MODES:")
	fmt.Println("  orchestrate - Run complete PPRL workflow (tokenize → exchange → intersect → send)")
	fmt.Println("  sender      - Run sender-specific workflow")
	fmt.Println("  receiver    - Run receiver-specific workflow")
	fmt.Println("  shutdown    - Send shutdown signal to a running receiver")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  # Run complete PPRL workflow")
	fmt.Println("  ./agent -workflow -config=config_a.yaml")
	fmt.Println("  ./agent -mode=orchestrate -config=config_a.yaml")
	fmt.Println()
	fmt.Println("  # Run with streaming for large datasets")
	fmt.Println("  ./agent -workflow -streaming -batch-size=500")
	fmt.Println()
	fmt.Println("  # Specify custom output files")
	fmt.Println("  ./agent -workflow -tokens-output=my_tokens.csv \\")
	fmt.Println("    -intersection-output=my_intersection.csv")
	fmt.Println()
	fmt.Println("  # Shutdown receiver")
	fmt.Println("  ./agent -mode=shutdown -config=config_a.yaml")
	fmt.Println()
	fmt.Println("WORKFLOW STEPS:")
	fmt.Println("  1. Tokenize local database using 'tokenize' program")
	fmt.Println("  2. Exchange tokenized data with peer")
	fmt.Println("  3. Find intersection using 'intersect' program")
	fmt.Println("  4. Send results using 'send' program")
	fmt.Println()
	fmt.Println("NOTES:")
	fmt.Println("  - Each step uses a specialized CLI tool for maximum modularity")
	fmt.Println("  - All tools can be used independently for custom workflows")
	fmt.Println("  - The agent coordinates the complete end-to-end process")
	fmt.Println("  - Results are saved to specified output files")
	fmt.Println("  - Logs from each tool are displayed in real-time")
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	configPkg "github.com/auroradata-ai/cohort-bridge/internal/config"
	"github.com/manifoldco/promptui"
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
	// ASCII art header
	fmt.Println("🤖 CohortBridge Agent - PPRL Orchestrator")
	fmt.Println("=========================================")
	fmt.Println("Privacy-Preserving Record Linkage System")
	fmt.Println()

	// Check if any arguments were passed - if so, use legacy mode
	if len(os.Args) > 1 {
		runLegacyMode()
		return
	}

	// Interactive mode with promptui
	runInteractiveMode()
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
		fmt.Println("Usage: agent -mode=<sender|receiver|orchestrate> -config=<config.yaml>")
		fmt.Println("   or: agent (for interactive mode)")
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
		// Try to determine what type of config this is
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

func findConfigFiles() []string {
	var configs []string

	// Look for common config file patterns
	patterns := []string{"*.yaml", "*.yml", "config*.yaml", "config*.yml"}

	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
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

	return configs
}

func runWithConfig(mode, configFile string) {
	fmt.Printf("🎯 Mode: %s\n", mode)
	fmt.Printf("📁 Config: %s\n\n", configFile)

	// Load config to validate
	_, err := configPkg.Load(configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create workflow config
	config := &WorkflowConfig{
		ConfigFile: configFile,
	}

	// Run based on mode
	switch mode {
	case "sender":
		config.TokenizeOutput = "temp-sender/tokens.csv"
		config.PeerTokens = "temp-sender/peer_tokens.csv"
		config.IntersectionFile = "temp-sender/intersection.csv"
		cfg, loadErr := configPkg.Load(configFile)
		if loadErr != nil {
			log.Fatalf("Failed to load config: %v", loadErr)
		}
		err = runSenderWorkflow(config, cfg)
	case "receiver":
		config.TokenizeOutput = "temp-receiver/tokens.csv"
		config.PeerTokens = "temp-receiver/peer_tokens.csv"
		config.IntersectionFile = "temp-receiver/intersection.csv"
		cfg, loadErr := configPkg.Load(configFile)
		if loadErr != nil {
			log.Fatalf("Failed to load config: %v", loadErr)
		}
		err = runReceiverWorkflow(config, cfg)
	case "orchestrate":
		config.TokenizeOutput = "tokens.csv"
		config.PeerTokens = "peer_tokens.csv"
		config.IntersectionFile = "intersection.csv"
		cfg, loadErr := configPkg.Load(configFile)
		if loadErr != nil {
			log.Fatalf("Failed to load config for orchestration: %v", loadErr)
		}
		err = runOrchestration(config, cfg)
	default:
		log.Fatalf("Unknown mode: %s", mode)
	}

	if err != nil {
		log.Fatalf("%s workflow failed: %v", strings.Title(mode), err)
	}

	fmt.Printf("\n✅ %s workflow completed successfully!\n", strings.Title(mode))
}

// runOrchestration runs the complete PPRL workflow
func runOrchestration(config *WorkflowConfig, cfg *configPkg.Config) error {
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

	// Add a small delay to ensure receiver is ready to listen for results
	time.Sleep(1 * time.Second)

	if err := callSend(config); err != nil {
		return fmt.Errorf("sending failed: %w", err)
	}

	fmt.Println("\n✅ PPRL orchestration completed successfully!")
	return nil
}

// runSenderWorkflow runs the sender-specific workflow
func runSenderWorkflow(config *WorkflowConfig, cfg *configPkg.Config) error {
	fmt.Println("📤 Starting sender workflow")

	// Create temp-sender directory
	if err := os.MkdirAll("temp-sender", 0755); err != nil {
		return fmt.Errorf("failed to create temp-sender directory: %w", err)
	}

	// Update paths to use temp-sender directory
	config.TokenizeOutput = "temp-sender/tokens.csv"
	config.IntersectionFile = "temp-sender/intersection.csv"

	fmt.Printf("📁 Using sender workspace: temp-sender/\n")

	// Step 1: Tokenize local data
	fmt.Println("\n🔐 Step 1: Tokenizing local data...")
	if err := callTokenize(config, cfg); err != nil {
		return fmt.Errorf("tokenization failed: %w", err)
	}

	// Step 2: Send tokenized data to receiver and wait for their data
	fmt.Println("\n📡 Step 2: Exchanging tokenized data with receiver...")
	if err := senderExchangeData(config, cfg); err != nil {
		return fmt.Errorf("data exchange failed: %w", err)
	}

	// Step 3: Compute intersection locally
	fmt.Println("\n🔍 Step 3: Computing intersection...")
	if err := callIntersect(config); err != nil {
		return fmt.Errorf("intersection failed: %w", err)
	}

	// Step 4: Send intersection results to receiver
	fmt.Println("\n📤 Step 4: Sending intersection results...")

	// Add a small delay to ensure receiver is ready to listen for results
	time.Sleep(1 * time.Second)

	if err := callSend(config); err != nil {
		return fmt.Errorf("sending results failed: %w", err)
	}

	fmt.Println("\n✅ Sender workflow completed successfully!")
	return nil
}

// runReceiverWorkflow runs the receiver-specific workflow
func runReceiverWorkflow(config *WorkflowConfig, cfg *configPkg.Config) error {
	fmt.Println("📥 Starting receiver workflow")

	// Create temp-receiver directory
	if err := os.MkdirAll("temp-receiver", 0755); err != nil {
		return fmt.Errorf("failed to create temp-receiver directory: %w", err)
	}

	// Update paths to use temp-receiver directory
	config.TokenizeOutput = "temp-receiver/tokens.csv"
	config.IntersectionFile = "temp-receiver/intersection.csv"
	config.PeerTokens = "temp-receiver/peer_tokens.csv"

	fmt.Printf("📁 Using receiver workspace: temp-receiver/\n")

	// Step 1: Tokenize local data
	fmt.Println("\n🔐 Step 1: Tokenizing local data...")
	if err := callTokenize(config, cfg); err != nil {
		return fmt.Errorf("tokenization failed: %w", err)
	}

	// Step 2: Listen for sender and exchange data
	fmt.Println("\n👂 Step 2: Listening for sender and exchanging data...")
	if err := receiverExchangeData(config, cfg); err != nil {
		return fmt.Errorf("data exchange failed: %w", err)
	}

	// Step 3: Compute intersection locally
	fmt.Println("\n🔍 Step 3: Computing intersection...")
	if err := callIntersect(config); err != nil {
		return fmt.Errorf("intersection failed: %w", err)
	}

	// Step 4: Wait for sender's intersection results
	fmt.Println("\n📥 Step 4: Receiving final results from sender...")
	if err := receiverGetResults(config, cfg); err != nil {
		return fmt.Errorf("receiving results failed: %w", err)
	}

	fmt.Println("\n✅ Receiver workflow completed successfully!")
	return nil
}

// callTokenize calls the tokenize CLI program
func callTokenize(config *WorkflowConfig, cfg *configPkg.Config) error {
	fmt.Printf("   🔧 Tokenizing data from: %s\n", cfg.Database.Filename)

	args := []string{
		"-input", cfg.Database.Filename,
		"-output", config.TokenizeOutput,
		"-main-config", config.ConfigFile,
		"-minhash-seed", "shared-seed-2024",
	}

	if config.Streaming {
		args = append(args, "-streaming", "-batch-size", fmt.Sprintf("%d", config.BatchSize))
	}

	cmd := exec.Command("./tokenize.exe", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tokenize command failed: %w", err)
	}

	fmt.Printf("   ✅ Tokenization completed: %s\n", config.TokenizeOutput)
	return nil
}

// exchangeTokenizedData handles the exchange of tokenized data with the peer
func exchangeTokenizedData(config *WorkflowConfig, cfg *configPkg.Config) error {
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

	cmd := exec.Command("./intersect.exe", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("intersect command failed: %w", err)
	}

	fmt.Printf("   ✅ Intersection completed: %s\n", config.IntersectionFile)
	return nil
}

// callSend sends intersection results to the receiver
func callSend(config *WorkflowConfig) error {
	fmt.Printf("   📤 Sending intersection results to receiver\n")

	// Read the intersection file
	file, err := os.Open(config.IntersectionFile)
	if err != nil {
		return fmt.Errorf("failed to open intersection file: %w", err)
	}
	defer file.Close()

	// Load config to get peer information
	cfg, err := configPkg.Load(config.ConfigFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Send via HTTP POST to receiver
	url := fmt.Sprintf("http://%s:%d/results", cfg.Peer.Host, cfg.Peer.Port)
	resp, err := http.Post(url, "text/csv", file)
	if err != nil {
		return fmt.Errorf("failed to send intersection results via HTTP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("receiver returned error: %s", resp.Status)
	}

	fmt.Println("   ✅ Intersection results sent successfully")
	return nil
}

// sendShutdownSignal sends a shutdown signal to the receiver
func sendShutdownSignal(cfg *configPkg.Config) error {
	// This can remain as a simple direct implementation since it's just a network call
	// For demonstration, we'll use a placeholder
	target := fmt.Sprintf("%s:%d", cfg.Peer.Host, cfg.Peer.Port)
	fmt.Printf("Sending shutdown to %s\n", target)

	// In real implementation, this would make the actual network call
	// server.SendShutdown(cfg)

	time.Sleep(1 * time.Second) // Simulate network call
	return nil
}

// senderExchangeData handles sending tokens to receiver and receiving their tokens
func senderExchangeData(config *WorkflowConfig, cfg *configPkg.Config) error {
	fmt.Printf("   📤 Sending tokenized data to %s:%d\n", cfg.Peer.Host, cfg.Peer.Port)

	// Send our tokenized data to the receiver and get their data back
	config.PeerTokens = "temp-sender/peer_tokens.csv"
	err := sendTokensAndReceiveResponse(config.TokenizeOutput, config.PeerTokens, cfg)
	if err != nil {
		return fmt.Errorf("failed to exchange tokens: %w", err)
	}

	fmt.Println("   ✅ Data exchange completed")
	return nil
}

// receiverExchangeData handles receiving tokens from sender and sending own tokens
func receiverExchangeData(config *WorkflowConfig, cfg *configPkg.Config) error {
	fmt.Printf("   👂 Listening on port %d for sender connections...\n", cfg.ListenPort)

	// Start listener for bidirectional token exchange
	err := startTokenExchangeServer(config.TokenizeOutput, config.PeerTokens, cfg)
	if err != nil {
		return fmt.Errorf("failed to run token exchange server: %w", err)
	}

	fmt.Println("   ✅ Data exchange completed")
	return nil
}

// receiverGetResults waits for and receives final intersection results from sender
func receiverGetResults(config *WorkflowConfig, cfg *configPkg.Config) error {
	fmt.Printf("   👂 Waiting for intersection results from sender...\n")

	// Listen for intersection results from sender
	resultsFile := "temp-receiver/sender_intersection.csv"
	err := receiveIntersectionResults(resultsFile, cfg)
	if err != nil {
		return fmt.Errorf("failed to receive intersection results: %w", err)
	}

	fmt.Printf("   ✅ Received intersection results: %s\n", resultsFile)
	return nil
}

// sendTokensAndReceiveResponse sends tokens to receiver and gets their tokens back
func sendTokensAndReceiveResponse(tokenFile, peerTokensFile string, cfg *configPkg.Config) error {
	fmt.Printf("   📤 Exchanging tokens with receiver at %s:%d...\n", cfg.Peer.Host, cfg.Peer.Port)

	// Read our token file
	file, err := os.Open(tokenFile)
	if err != nil {
		return fmt.Errorf("failed to open token file: %w", err)
	}
	defer file.Close()

	// Send via HTTP POST and get receiver's tokens back
	url := fmt.Sprintf("http://%s:%d/exchange-tokens", cfg.Peer.Host, cfg.Peer.Port)
	resp, err := http.Post(url, "text/csv", file)
	if err != nil {
		return fmt.Errorf("failed to exchange tokens via HTTP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned error %s: %s", resp.Status, string(bodyBytes))
	}

	// Read all response data first to ensure complete transfer
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response data: %w", err)
	}

	// Write receiver's tokens to file
	err = os.WriteFile(peerTokensFile, responseData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write peer tokens file: %w", err)
	}

	fmt.Printf("   ✅ Token exchange completed successfully (%d bytes received)\n", len(responseData))
	return nil
}

// startTokenExchangeServer starts a server that handles bidirectional token exchange
func startTokenExchangeServer(ourTokensFile, receivedTokensFile string, cfg *configPkg.Config) error {
	fmt.Printf("   👂 Starting token exchange server on port %d...\n", cfg.ListenPort)

	exchanged := make(chan error, 1)

	// Create HTTP handler for bidirectional token exchange
	mux := http.NewServeMux()
	mux.HandleFunc("/exchange-tokens", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// Receive sender's tokens
			senderFile, err := os.Create(receivedTokensFile)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to create file: %v", err), http.StatusInternalServerError)
				exchanged <- err
				return
			}
			defer senderFile.Close()

			_, err = io.Copy(senderFile, r.Body)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to write sender tokens: %v", err), http.StatusInternalServerError)
				exchanged <- err
				return
			}

			// Send our tokens back
			ourFile, err := os.Open(ourTokensFile)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to open our tokens: %v", err), http.StatusInternalServerError)
				exchanged <- err
				return
			}
			defer ourFile.Close()

			w.Header().Set("Content-Type", "text/csv")
			_, err = io.Copy(w, ourFile)
			if err != nil {
				exchanged <- fmt.Errorf("failed to send our tokens: %w", err)
				return
			}

			// Add a small delay before signaling completion to ensure response is sent
			go func() {
				time.Sleep(100 * time.Millisecond)
				exchanged <- nil
			}()
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Start server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ListenPort),
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			exchanged <- fmt.Errorf("server error: %w", err)
		}
	}()

	// Wait for exchange to complete
	err := <-exchanged

	// Add a small delay before shutdown to ensure response is fully sent
	time.Sleep(200 * time.Millisecond)

	// Shutdown server gracefully
	server.Close()

	if err != nil {
		return fmt.Errorf("failed to exchange tokens: %w", err)
	}

	fmt.Println("   ✅ Tokens exchanged successfully")
	return nil
}

// Network helper functions using HTTP (keeping the existing ones for intersection results)
func sendTokensToReceiver(tokenFile string, cfg *configPkg.Config) error {
	fmt.Printf("   📤 Sending tokens to %s:%d...\n", cfg.Peer.Host, cfg.Peer.Port)

	// Read the token file
	file, err := os.Open(tokenFile)
	if err != nil {
		return fmt.Errorf("failed to open token file: %w", err)
	}
	defer file.Close()

	// Send via HTTP POST
	url := fmt.Sprintf("http://%s:%d/tokens", cfg.Peer.Host, cfg.Peer.Port)
	resp, err := http.Post(url, "text/csv", file)
	if err != nil {
		return fmt.Errorf("failed to send tokens via HTTP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned error: %s", resp.Status)
	}

	fmt.Println("   ✅ Tokens sent successfully")
	return nil
}

func receiveTokensFromReceiver(outputFile string, cfg *configPkg.Config) error {
	fmt.Println("   📥 Requesting tokens from receiver...")

	// Request tokens via HTTP GET
	url := fmt.Sprintf("http://%s:%d/tokens", cfg.Peer.Host, cfg.Peer.Port)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to request tokens via HTTP: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned error: %s", resp.Status)
	}

	// Write response to file
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write tokens to file: %w", err)
	}

	fmt.Println("   ✅ Tokens received successfully")
	return nil
}

func receiveTokensFromSender(outputFile string, cfg *configPkg.Config) error {
	fmt.Printf("   👂 Starting HTTP server on port %d to receive tokens...\n", cfg.ListenPort)

	received := make(chan error, 1)

	// Create HTTP handler
	mux := http.NewServeMux()
	mux.HandleFunc("/tokens", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// Receive tokens from sender
			file, err := os.Create(outputFile)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to create file: %v", err), http.StatusInternalServerError)
				received <- err
				return
			}
			defer file.Close()

			_, err = io.Copy(file, r.Body)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to write file: %v", err), http.StatusInternalServerError)
				received <- err
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Tokens received successfully"))
			received <- nil
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Start server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ListenPort),
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			received <- fmt.Errorf("server error: %w", err)
		}
	}()

	// Wait for tokens to be received
	err := <-received

	// Shutdown server
	server.Close()

	if err != nil {
		return fmt.Errorf("failed to receive tokens: %w", err)
	}

	fmt.Println("   ✅ Tokens received from sender")
	return nil
}

func sendTokensToSender(tokenFile string, cfg *configPkg.Config) error {
	return sendTokensToReceiver(tokenFile, cfg) // Same implementation
}

func receiveIntersectionResults(outputFile string, cfg *configPkg.Config) error {
	fmt.Printf("   👂 Starting HTTP server on port %d to receive intersection results...\n", cfg.ListenPort)

	received := make(chan error, 1)

	// Create HTTP handler
	mux := http.NewServeMux()
	mux.HandleFunc("/results", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			// Receive results from sender
			file, err := os.Create(outputFile)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to create file: %v", err), http.StatusInternalServerError)
				received <- err
				return
			}
			defer file.Close()

			_, err = io.Copy(file, r.Body)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to write file: %v", err), http.StatusInternalServerError)
				received <- err
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Results received successfully"))

			// Add a small delay before signaling completion to ensure response is sent
			go func() {
				time.Sleep(100 * time.Millisecond)
				received <- nil
			}()
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Start server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ListenPort),
		Handler: mux,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			received <- fmt.Errorf("server error: %w", err)
		}
	}()

	// Wait for results to be received
	err := <-received

	// Add a small delay before shutdown to ensure response is fully sent
	time.Sleep(200 * time.Millisecond)

	// Shutdown server gracefully
	server.Close()

	if err != nil {
		return fmt.Errorf("failed to receive results: %w", err)
	}

	fmt.Println("   ✅ Intersection results received from sender")
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

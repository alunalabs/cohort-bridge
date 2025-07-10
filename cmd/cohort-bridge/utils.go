package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
)

// promptForInput reads text input from user with optional default value
func promptForInput(message, defaultValue string) string {
	time.Sleep(time.Millisecond)
	if defaultValue != "" {
		fmt.Printf("%s (default: %s): ", message, defaultValue)
	} else {
		fmt.Printf("%s: ", message)
	}

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		return defaultValue
	}

	input = strings.TrimSpace(input)
	if input == "" && defaultValue != "" {
		return defaultValue
	}
	return input
}

// promptForChoice uses promptui for menu selection with arrow keys
func promptForChoice(message string, options []string) int {
	prompt := promptui.Select{
		Label: message,
		Items: options,
		Size:  10, // Show up to 10 items at once
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "> {{ . | cyan }}",
			Inactive: "  {{ . }}",
			Selected: "Selected: {{ . | green }}",
		},
	}

	index, _, err := prompt.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	return index
}

// selectDataFile helps user select a data file from a directory with specific extensions
func selectDataFile(label, context string, extensions []string) (string, error) {
	// Look for files in data directory and current directory
	var dataFiles []string

	// Check data directory
	dataDir := "data"
	if _, err := os.Stat(dataDir); err == nil {
		matches, _ := filepath.Glob(filepath.Join(dataDir, "*"))
		for _, match := range matches {
			if isFileWithExtensions(match, extensions) {
				dataFiles = append(dataFiles, match)
			}
		}
	}

	// Check current directory
	matches, _ := filepath.Glob("*")
	for _, match := range matches {
		if isFileWithExtensions(match, extensions) {
			dataFiles = append(dataFiles, match)
		}
	}

	if len(dataFiles) == 0 {
		// Manual input if no files found
		return promptForInput(label+" (enter file path)", ""), nil
	}

	// Add manual entry option
	dataFiles = append(dataFiles, "Enter custom path...")

	choice := promptForChoice(label+":", dataFiles)

	if choice == len(dataFiles)-1 {
		// Manual entry
		return promptForInput("Enter file path", ""), nil
	}

	return dataFiles[choice], nil
}

// selectConfigFile helps user select a configuration file
func selectConfigFile(label string) (string, error) {
	// Find YAML config files in current directory
	var configFiles []string

	matches, _ := filepath.Glob("*.yaml")
	for _, match := range matches {
		if strings.Contains(strings.ToLower(match), "example") {
			continue // Skip example files
		}
		configFiles = append(configFiles, match)
	}

	if len(configFiles) == 0 {
		// Manual input if no files found
		return promptForInput(label+" (enter .yaml file path)", ""), nil
	}

	// Add descriptions and manual entry option
	var options []string
	for _, file := range configFiles {
		desc := getConfigDescription(file)
		if desc != "" {
			options = append(options, fmt.Sprintf("%s - %s", file, desc))
		} else {
			options = append(options, file)
		}
	}
	options = append(options, "Enter custom path...")

	choice := promptForChoice(label+":", options)

	if choice == len(options)-1 {
		// Manual entry
		return promptForInput("Enter config file path", ""), nil
	}

	return configFiles[choice], nil
}

// selectGroundTruthFile helps user select a ground truth file from data directory
func selectGroundTruthFile() (string, error) {
	dataDir := "data"
	var gtFiles []string

	// Look for files with "ground" or "truth" in the name
	if _, err := os.Stat(dataDir); err == nil {
		matches, _ := filepath.Glob(filepath.Join(dataDir, "*"))
		for _, match := range matches {
			filename := strings.ToLower(filepath.Base(match))
			if strings.Contains(filename, "ground") || strings.Contains(filename, "truth") {
				gtFiles = append(gtFiles, match)
			}
		}
	}

	// Also check current directory
	matches, _ := filepath.Glob("*")
	for _, match := range matches {
		filename := strings.ToLower(filepath.Base(match))
		if strings.Contains(filename, "ground") || strings.Contains(filename, "truth") {
			gtFiles = append(gtFiles, match)
		}
	}

	if len(gtFiles) == 0 {
		return promptForInput("Ground truth file path", ""), nil
	}

	// Add manual entry option
	gtFiles = append(gtFiles, "Enter custom path...")

	choice := promptForChoice("Select ground truth file:", gtFiles)

	if choice == len(gtFiles)-1 {
		return promptForInput("Enter ground truth file path", ""), nil
	}

	return gtFiles[choice], nil
}

// ifDefault returns a default indicator string
func ifDefault(isDefault bool) string {
	if isDefault {
		return "(default)"
	}
	return ""
}

// copyToOutput copies a file from source to output directory
func copyToOutput(srcFile, dstFile string) error {
	// Ensure output directory exists
	outputDir := "out"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create destination path
	dstPath := filepath.Join(outputDir, dstFile)

	// Read source file
	srcData, err := os.ReadFile(srcFile)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Write to destination
	if err := os.WriteFile(dstPath, srcData, 0644); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}

// copyToAbsolutePath copies a file to an absolute destination path
func copyToAbsolutePath(srcFile, dstPath string) error {
	// Ensure destination directory exists
	dstDir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Read source file
	srcData, err := os.ReadFile(srcFile)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	// Write to destination
	if err := os.WriteFile(dstPath, srcData, 0644); err != nil {
		return fmt.Errorf("failed to write destination file: %w", err)
	}

	return nil
}

// confirmStep prompts user for confirmation unless force mode is enabled
func confirmStep(message string, force bool) bool {
	if force {
		return true
	}

	choice := promptForChoice(message, []string{
		"Yes, continue",
		"No, cancel",
	})

	return choice == 0
}

// generateOutputName creates standardized output file names
func generateOutputName(prefix string, inputs ...string) string {
	// Clean and combine input names
	var parts []string
	for _, input := range inputs {
		if input != "" {
			base := strings.TrimSuffix(filepath.Base(input), filepath.Ext(input))
			// Replace common separators with underscores
			base = strings.ReplaceAll(base, "-", "_")
			base = strings.ReplaceAll(base, " ", "_")
			parts = append(parts, base)
		}
	}

	filename := prefix
	if len(parts) > 0 {
		filename += "_" + strings.Join(parts, "_vs_")
	}

	return filepath.Join("out", filename+".csv")
}

// isFileWithExtensions checks if file has one of the specified extensions
func isFileWithExtensions(filename string, extensions []string) bool {
	if len(extensions) == 0 {
		return true
	}

	ext := strings.ToLower(filepath.Ext(filename))
	for _, validExt := range extensions {
		if ext == strings.ToLower(validExt) {
			return true
		}
	}
	return false
}

// getConfigDescription returns a description for a config file based on its name
func getConfigDescription(filename string) string {
	name := strings.ToLower(filename)
	switch {
	case strings.Contains(name, "basic"):
		return "Basic CSV file configuration"
	case strings.Contains(name, "postgres"):
		return "PostgreSQL database configuration"
	case strings.Contains(name, "tokenized"):
		return "Pre-tokenized data configuration"
	case strings.Contains(name, "test"):
		return "Test configuration"
	default:
		return ""
	}
}

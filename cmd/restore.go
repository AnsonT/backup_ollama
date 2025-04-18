package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"backup_ollama/internal/utils"

	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore [model name]",
	Short: "Restore a model from backup",
	Long:  `Restore a model from a specified backup directory.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		modelName := args[0]
		backupDir, _ := cmd.Flags().GetString("backup-dir")

		err := restoreModel(modelName, backupDir)
		if err != nil {
			fmt.Printf("Error restoring model: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Model '%s' restored successfully from '%s'.\n", modelName, backupDir)
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
	restoreCmd.Flags().StringP("backup-dir", "d", "", "Directory to restore from")
}

func restoreModel(modelName string, backupDir string) error {
	if backupDir == "" {
		return fmt.Errorf("backup directory must be specified")
	}

	// Check if the backup directory exists
	if _, err := os.Stat(backupDir); os.IsNotExist(err) {
		return fmt.Errorf("backup directory '%s' does not exist", backupDir)
	}

	// Parse the model name to separate model and version if provided
	parts := strings.Split(modelName, ":")
	model := parts[0]
	var version string
	if len(parts) > 1 {
		version = parts[1]
	} else {
		// If version is not specified, use "latest"
		version = "latest"
	}

	// Find the appropriate backup directory for the model
	backupModelDir := ""
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), model+"--") {
			backupModelDir = filepath.Join(backupDir, entry.Name())
			break
		}
	}

	if backupModelDir == "" {
		return fmt.Errorf("no backup found for model '%s'", model)
	}

	fmt.Printf("Found backup directory: %s\n", backupModelDir)

	// Check if the manifest directory exists in the backup
	manifestsDir := filepath.Join(backupModelDir, "library", "manifests")
	if _, err := os.Stat(manifestsDir); os.IsNotExist(err) {
		return fmt.Errorf("invalid backup structure: manifests directory not found")
	}

	// Find the registry from the backup
	registryDirs, err := os.ReadDir(manifestsDir)
	if err != nil || len(registryDirs) == 0 {
		return fmt.Errorf("failed to find registry in backup")
	}
	registry := registryDirs[0].Name()

	// Read the manifest
	manifestPath := filepath.Join(manifestsDir, registry, "library", model, version)
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest map[string]interface{}
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Get the Ollama directory
	ollamaDir := utils.GetOllamaDirectory()
	if ollamaDir == "" {
		return fmt.Errorf("failed to determine Ollama directory")
	}

	// Create the blobs directory if it doesn't exist
	ollamaBlobsDir := filepath.Join(ollamaDir, "models", "blobs")
	if err := os.MkdirAll(ollamaBlobsDir, 0755); err != nil {
		return fmt.Errorf("failed to create Ollama blobs directory: %w", err)
	}

	// Create the manifests directory structure
	ollemaManifestsDir := filepath.Join(ollamaDir, "models", "manifests", registry, "library", model)
	if err := os.MkdirAll(ollemaManifestsDir, 0755); err != nil {
		return fmt.Errorf("failed to create Ollama manifests directory: %w", err)
	}

	// Extract and copy blob files from the backup
	if layers, ok := manifest["layers"].([]interface{}); ok {
		for _, layer := range layers {
			if layerMap, ok := layer.(map[string]interface{}); ok {
				var sourceFile string
				var destFile string

				// Check if the layer has a 'from' field
				if from, ok := layerMap["from"].(string); ok {
					// Use the 'from' field as the full path to the blob
					sourceFile = filepath.Join(backupModelDir, "blobs", filepath.Base(from))
					destFile = filepath.Join(ollamaDir, from)
				} else if digest, ok := layerMap["digest"].(string); ok {
					// Use the digest as the file name in the ~/.ollama/models/blobs/ directory
					blobFileName := strings.TrimPrefix(digest, "sha256:")
					sourceFile = filepath.Join(backupModelDir, "blobs", blobFileName)
					destFile = filepath.Join(ollamaBlobsDir, blobFileName)
				} else {
					fmt.Printf("Warning: Layer without 'from' or 'digest' field found, skipping.\n")
					continue
				}

				// Create the parent directory for the destination file
				if err := os.MkdirAll(filepath.Dir(destFile), 0755); err != nil {
					return fmt.Errorf("failed to create directory for blob: %w", err)
				}

				// Copy the blob file
				if err := copyFile(sourceFile, destFile); err != nil {
					return fmt.Errorf("failed to copy blob file: %w", err)
				}

				fmt.Printf("Copied blob to: %s\n", destFile)
			}
		}
	} else {
		return fmt.Errorf("no layers found in manifest")
	}

	// Copy the manifest
	destManifestPath := filepath.Join(ollemaManifestsDir, version)
	if err := copyFile(manifestPath, destManifestPath); err != nil {
		return fmt.Errorf("failed to copy manifest: %w", err)
	}

	fmt.Printf("Copied manifest to: %s\n", destManifestPath)

	return nil
}

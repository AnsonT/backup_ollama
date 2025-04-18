package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"backup_ollama/internal/utils"

	"github.com/spf13/cobra"
)

var backupDir string

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup [model name]",
	Short: "Backup a specified model",
	Long:  `This command allows you to backup a specified model to a designated directory.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		modelName := args[0]
		if err := backupModel(modelName, backupDir); err != nil {
			fmt.Fprintf(os.Stderr, "Error backing up model: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Model '%s' backed up successfully to '%s'\n", modelName, backupDir)
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
	backupCmd.Flags().StringVarP(&backupDir, "dir", "d", "./backup", "Directory to save the backup")
}

// validateModelName validates the model name string against the enumerated models.
// Model name is expected in the format '{model}:{version}'.
// If version is not specified and there's only one version in the directory, it uses the available version.
// Returns the model name, model version, registry, manifest details and the full path to the manifest if successful, or an error if validation fails.
func validateModelName(modelName string) (string, string, string, map[string]interface{}, string, error) {
	// Split the model name into model and version parts
	parts := strings.Split(modelName, ":")
	model := parts[0]
	var version string
	if len(parts) > 1 {
		version = parts[1]
	}

	// Get all Ollama models
	modelList, err := utils.EnumerateOllamaModels()
	if err != nil {
		return "", "", "", nil, "", fmt.Errorf("failed to enumerate models: %w", err)
	}

	// Look for the specified model in the list
	var targetModel *utils.Model
	var registry string
	for _, reg := range modelList.Registries {
		for i, m := range reg.Models {
			if m.Name == model {
				targetModel = &reg.Models[i]
				registry = reg.Name
				break
			}
		}
		if targetModel != nil {
			break
		}
	}

	if targetModel == nil {
		return "", "", "", nil, "", fmt.Errorf("model '%s' not found", model)
	}

	// If no version is specified
	if version == "" {
		if len(targetModel.Versions) == 0 {
			return "", "", "", nil, "", fmt.Errorf("no versions found for model '%s'", model)
		} else if len(targetModel.Versions) == 1 {
			// If there's only one version, use it
			version = targetModel.Versions[0].Name
			fmt.Printf("Using version '%s' for model '%s'\n", version, model)
		} else {
			// If there are multiple versions, list them and ask the user to specify
			fmt.Printf("Multiple versions found for model '%s'. Please specify a version using the format 'model:version'.\n", model)
			fmt.Println("Available versions:")
			for _, v := range targetModel.Versions {
				fmt.Printf("- %s\n", v.Name)
			}
			return "", "", "", nil, "", fmt.Errorf("version not specified")
		}
	}

	// Look for the specified version
	var targetVersion *utils.ModelVersion
	for i, v := range targetModel.Versions {
		if v.Name == version {
			targetVersion = &targetModel.Versions[i]
			break
		}
	}

	if targetVersion == nil {
		return "", "", "", nil, "", fmt.Errorf("version '%s' not found for model '%s'", version, model)
	}

	// Return the model name, version, registry, manifest details, and full path to the manifest
	return model, version, registry, targetVersion.Details, targetVersion.Path, nil
}

// backupModel performs the actual backup operation
func backupModel(modelName, dir string) error {
	// Validate the model name and get the model information
	model, version, registry, manifest, manifestPath, err := validateModelName(modelName)
	if err != nil {
		return err
	}

	fmt.Printf("Found model: %s, version: %s, registry: %s, manifest path: %s\n", model, version, registry, manifestPath)

	// Generate a backup version based on timestamp
	backupVersion := fmt.Sprintf("backup-%d", time.Now().Unix())

	// Create the destination directory with format {backup directory}/{model}--{version}--{backup version}/
	backupPath := filepath.Join(dir, fmt.Sprintf("%s--%s--%s", model, version, backupVersion))
	if err := os.MkdirAll(backupPath, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Create the blobs directory
	blobsDir := filepath.Join(backupPath, "blobs")
	if err := os.MkdirAll(blobsDir, 0755); err != nil {
		return fmt.Errorf("failed to create blobs directory: %w", err)
	}

	// Get the Ollama directory for accessing the actual blobs
	ollamaDir := utils.GetOllamaDirectory()
	ollamaBlobsDir := filepath.Join(ollamaDir, "models", "blobs")

	// Extract and copy the blob files from the manifest
	if layers, ok := manifest["layers"].([]interface{}); ok {
		for _, layer := range layers {
			if layerMap, ok := layer.(map[string]interface{}); ok {
				var sourcePath string
				var fileName string

				if from, ok := layerMap["from"].(string); ok {
					// The 'from' field exists, use it as the full path
					sourcePath = filepath.Join(ollamaDir, from)
					fileName = filepath.Base(sourcePath)
				} else if digest, ok := layerMap["digest"].(string); ok {
					// The 'from' field doesn't exist, use the digest to construct the file name
					// Convert digest format from "sha256:123abc..." to "sha256-123abc..."
					digestName := strings.Replace(digest, ":", "-", 1)
					fileName = digestName
					sourcePath = filepath.Join(ollamaBlobsDir, fileName)
				} else {
					// Skip this layer if neither from nor digest is present
					fmt.Printf("Skipping layer: no 'from' or 'digest' field found\n")
					continue
				}

				destPath := filepath.Join(blobsDir, fileName)

				// Copy the file
				if err := copyFile(sourcePath, destPath); err != nil {
					return fmt.Errorf("failed to copy blob file: %w", err)
				}

				fmt.Printf("Copied blob: %s\n", fileName)
			}
		}
	}

	// Create the directory structure for the manifest
	// Format: {backup directory}/{model}--{version}--{backup version}/library/manifests/{registry}/library/{model}/{model version}

	// Use the registry from validateModelName instead of extracting it from the manifest
	manifestDir := filepath.Join(backupPath, "library", "manifests", registry, "library", model)
	if err := os.MkdirAll(manifestDir, 0755); err != nil {
		return fmt.Errorf("failed to create manifest directory: %w", err)
	}

	// Marshal the manifest to JSON and save it
	manifestData, err := json.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	// Save the manifest to {backup directory}/{model}--{version}--{backup version}/library/manifests/{registry}/library/{model}/{model version}
	manifestOutputPath := filepath.Join(manifestDir, version)
	if err := os.WriteFile(manifestOutputPath, manifestData, 0644); err != nil {
		return fmt.Errorf("failed to write manifest file: %w", err)
	}

	fmt.Printf("Saved manifest for %s:%s from registry %s in backup %s\n", model, version, registry, backupVersion)

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return nil
}

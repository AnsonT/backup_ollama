package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ModelVersion represents information about a specific model version
type ModelVersion struct {
	Name    string
	Path    string
	Digest  string
	Size    int64
	Details map[string]interface{} // Parsed content of the version JSON file
}

// Model represents a model with its versions
type Model struct {
	Name     string
	Registry string
	Path     string
	Versions []ModelVersion
}

// Registry represents a registry with its models
type Registry struct {
	Name   string
	Path   string
	Models []Model
}

// OllamaModelList contains all registries, models and versions
type OllamaModelList struct {
	Registries []Registry
}

// GetOllamaDirectory returns the path to the Ollama data directory
func GetOllamaDirectory() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".ollama")
}

// GetManifestsDirectory returns the path to the Ollama manifests directory
func GetManifestsDirectory() string {
	return filepath.Join(GetOllamaDirectory(), "models", "manifests")
}

// GetBackupDirectory returns the full path of the backup directory.
// If a custom directory is provided, it returns that; otherwise, it returns the default backup directory.
func GetBackupDirectory(customDir string) string {
	if customDir != "" {
		return filepath.Clean(customDir)
	}
	return filepath.Join("default", "backup", "directory")
}

// GetModelPath constructs the full path for a given model name in the backup directory.
func GetModelPath(modelName string, backupDir string) string {
	return filepath.Join(backupDir, sanitizeModelName(modelName))
}

// sanitizeModelName sanitizes the model name to ensure it is a valid directory name.
func sanitizeModelName(modelName string) string {
	return strings.ReplaceAll(modelName, " ", "_")
}

// EnumerateOllamaModels scans the Ollama directory structure and returns information
// about all registries, models, and versions found.
// The structure is expected to be:
// ~/.ollama/models/manifests/{registry}/library/{model}/{version}
func EnumerateOllamaModels() (*OllamaModelList, error) {
	manifestsDir := GetManifestsDirectory()
	if _, err := os.Stat(manifestsDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("manifests directory does not exist: %s", manifestsDir)
	}

	result := &OllamaModelList{
		Registries: []Registry{},
	}

	// Step 1: List registry directories
	registryEntries, err := os.ReadDir(manifestsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifests directory: %w", err)
	}

	for _, registryEntry := range registryEntries {
		if !registryEntry.IsDir() {
			continue
		}

		registryName := registryEntry.Name()
		registryPath := filepath.Join(manifestsDir, registryName)

		registry := Registry{
			Name:   registryName,
			Path:   registryPath,
			Models: []Model{},
		}

		// Step 2: Look for the "library" directory
		libraryPath := filepath.Join(registryPath, "library")
		if _, err := os.Stat(libraryPath); os.IsNotExist(err) {
			continue
		}

		// Step 3: List model directories
		modelEntries, err := os.ReadDir(libraryPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read library directory: %w", err)
		}

		for _, modelEntry := range modelEntries {
			if !modelEntry.IsDir() {
				continue
			}

			modelName := modelEntry.Name()
			modelPath := filepath.Join(libraryPath, modelName)

			model := Model{
				Name:     modelName,
				Registry: registryName,
				Path:     modelPath,
				Versions: []ModelVersion{},
			}

			// Step 4: List version files
			versionEntries, err := os.ReadDir(modelPath)
			if err != nil {
				return nil, fmt.Errorf("failed to read model directory: %w", err)
			}

			for _, versionEntry := range versionEntries {
				if versionEntry.IsDir() {
					continue // Skip directories, we're looking for JSON files
				}

				versionPath := filepath.Join(modelPath, versionEntry.Name())

				// Parse the version JSON file
				versionInfo, err := parseVersionFile(versionPath)
				if err != nil {
					// Skip this version if we can't parse it
					continue
				}

				fileInfo, err := os.Stat(versionPath)
				if err != nil {
					continue
				}

				version := ModelVersion{
					Name:    versionEntry.Name(),
					Path:    versionPath,
					Size:    fileInfo.Size(),
					Details: versionInfo,
				}

				// Extract digest if available
				if digest, ok := versionInfo["digest"].(string); ok {
					version.Digest = digest
				}

				model.Versions = append(model.Versions, version)
			}

			if len(model.Versions) > 0 {
				registry.Models = append(registry.Models, model)
			}
		}

		if len(registry.Models) > 0 {
			result.Registries = append(result.Registries, registry)
		}
	}

	return result, nil
}

// parseVersionFile reads and parses a version JSON file
func parseVersionFile(path string) (map[string]interface{}, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read version file: %w", err)
	}

	var versionInfo map[string]interface{}
	if err := json.Unmarshal(content, &versionInfo); err != nil {
		return nil, fmt.Errorf("failed to parse version file: %w", err)
	}

	return versionInfo, nil
}

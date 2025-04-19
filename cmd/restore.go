package cmd

import (
	"archive/zip"
	"fmt"
	"io"
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
		ollamaDir, _ := cmd.Flags().GetString("ollama-dir")
		overwrite, _ := cmd.Flags().GetBool("overwrite")

		err := restoreModel(modelName, backupDir, ollamaDir, overwrite)
		if err != nil {
			fmt.Printf("Error restoring model: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Model '%s' restored successfully from '%s'.\n", modelName, backupDir)
	},
}

func init() {
	rootCmd.AddCommand(restoreCmd)
	restoreCmd.Flags().StringP("backup-dir", "d", "./backup", "Directory to restore from")
	restoreCmd.Flags().String("ollama-dir", utils.GetOllamaDirectory(), "Ollama directory to restore data to")
	restoreCmd.Flags().BoolP("overwrite", "o", false, "Overwrite existing files during restore")
}

// restoreModel handles restoring a model from a backup directory or zip file
func restoreModel(modelName string, backupDir string, ollamaDir string, overwrite bool) error {
	// Construct the full source path
	sourcePath := filepath.Join(backupDir, modelName)

	// Check if the source is a zip file
	if strings.HasSuffix(modelName, ".zip") {
		// Extract the zip file to the backup directory
		extractedDir, err := unzipBackup(sourcePath, backupDir)
		if err != nil {
			return fmt.Errorf("failed to unzip backup: %w", err)
		}

		// Update the source path to the extracted directory
		sourcePath = extractedDir
		fmt.Printf("Unzipped backup to: %s\n", extractedDir)
	}

	// Check if the source directory exists
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("backup not found: %w", err)
	}

	// Ensure source is a directory
	if !sourceInfo.IsDir() {
		return fmt.Errorf("backup source is not a directory: %s", sourcePath)
	}

	// Get paths to blobsDir and libraryDir
	sourceBlobsDir := filepath.Join(sourcePath, "blobs")
	sourceLibraryDir := filepath.Join(sourcePath, "library")

	// Make sure these directories exist
	if _, err := os.Stat(sourceBlobsDir); os.IsNotExist(err) {
		return fmt.Errorf("blobs directory missing in backup: %s", sourceBlobsDir)
	}

	if _, err := os.Stat(sourceLibraryDir); os.IsNotExist(err) {
		return fmt.Errorf("library directory missing in backup: %s", sourceLibraryDir)
	}

	// Get target directories
	ollamaBlobsDir := filepath.Join(ollamaDir, "models", "blobs")
	ollamaManifestsDir := filepath.Join(ollamaDir, "models", "manifests")

	// Create target directories if they don't exist
	if err := os.MkdirAll(ollamaBlobsDir, 0755); err != nil {
		return fmt.Errorf("failed to create ollama blobs directory: %w", err)
	}

	if err := os.MkdirAll(ollamaManifestsDir, 0755); err != nil {
		return fmt.Errorf("failed to create ollama manifests directory: %w", err)
	}

	// First check if files already exist (if not overwriting)
	if !overwrite {
		// Check blobs directory
		blobsExist, err := checkFilesExist(sourceBlobsDir, ollamaBlobsDir)
		if err != nil {
			return err
		}
		if blobsExist {
			return fmt.Errorf("some blob files already exist; use --overwrite to force restore")
		}

		// Check library/manifests directory
		manifestsExist, err := checkManifestsExist(sourceLibraryDir, ollamaManifestsDir)
		if err != nil {
			return err
		}
		if manifestsExist {
			return fmt.Errorf("some manifest files already exist; use --overwrite to force restore")
		}
	}

	// Copy blob files
	if err := copyDirectory(sourceBlobsDir, ollamaBlobsDir, overwrite); err != nil {
		return fmt.Errorf("failed to copy blob files: %w", err)
	}
	fmt.Println("Copied blob files successfully")

	// Copy library/manifests files
	sourceManifestsDir := filepath.Join(sourceLibraryDir, "manifests")
	if err := copyDirectory(sourceManifestsDir, ollamaManifestsDir, overwrite); err != nil {
		return fmt.Errorf("failed to copy manifest files: %w", err)
	}
	fmt.Println("Copied manifest files successfully")

	return nil
}

// unzipBackup extracts a zip file to the specified directory and returns the path to the extracted directory
func unzipBackup(zipFile string, destDir string) (string, error) {
	// Open the zip file
	reader, err := zip.OpenReader(zipFile)
	if err != nil {
		return "", fmt.Errorf("failed to open zip file: %w", err)
	}
	defer reader.Close()

	// Create a directory for the extracted contents
	// Remove .zip extension to get the base name
	baseName := strings.TrimSuffix(filepath.Base(zipFile), filepath.Ext(zipFile))
	extractDir := filepath.Join(destDir, baseName)

	// Check if the directory already exists
	if _, err := os.Stat(extractDir); err == nil {
		// Directory exists, remove it
		if err := os.RemoveAll(extractDir); err != nil {
			return "", fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}

	// Create the directory
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Extract each file
	for _, file := range reader.File {
		path := filepath.Join(extractDir, file.Name)

		// Check if it's a directory
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(path, file.Mode()); err != nil {
				return "", fmt.Errorf("failed to create directory: %w", err)
			}
			continue
		}

		// Create the file
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return "", fmt.Errorf("failed to create directory: %w", err)
		}

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return "", fmt.Errorf("failed to create file: %w", err)
		}

		rc, err := file.Open()
		if err != nil {
			outFile.Close()
			return "", fmt.Errorf("failed to open file in zip: %w", err)
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return "", fmt.Errorf("failed to copy file content: %w", err)
		}
	}

	return extractDir, nil
}

// checkFilesExist checks if any file in sourcePath already exists in destPath
func checkFilesExist(sourcePath, destPath string) (bool, error) {
	// Walk through the source directory
	return walkAndCheck(sourcePath, func(relPath string) (bool, error) {
		// Check if file exists in destination
		destFile := filepath.Join(destPath, relPath)
		if _, err := os.Stat(destFile); err == nil {
			fmt.Printf("File already exists: %s\n", destFile)
			return true, nil
		} else if !os.IsNotExist(err) {
			return false, err
		}
		return false, nil
	})
}

// checkManifestsExist checks if any manifest file in sourceLibraryDir already exists in ollamaManifestsDir
func checkManifestsExist(sourceLibraryDir, ollamaManifestsDir string) (bool, error) {
	// Walk through the source manifests directory
	sourceManifestsDir := filepath.Join(sourceLibraryDir, "manifests")
	if _, err := os.Stat(sourceManifestsDir); os.IsNotExist(err) {
		return false, nil
	}

	return walkAndCheck(sourceManifestsDir, func(relPath string) (bool, error) {
		// Check if file exists in destination
		destFile := filepath.Join(ollamaManifestsDir, relPath)
		if _, err := os.Stat(destFile); err == nil {
			fmt.Printf("Manifest already exists: %s\n", destFile)
			return true, nil
		} else if !os.IsNotExist(err) {
			return false, err
		}
		return false, nil
	})
}

// walkAndCheck walks a directory and applies a check function to each file
func walkAndCheck(rootPath string, checkFn func(string) (bool, error)) (bool, error) {
	exists := false

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}

		// Apply check function
		fileExists, err := checkFn(relPath)
		if err != nil {
			return err
		}

		if fileExists {
			exists = true
			// Don't stop the walk, continue checking other files
		}

		return nil
	})

	return exists, err
}

// copyDirectory copies files from src to dst
func copyDirectory(src, dst string, overwrite bool) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Construct destination path
		dstPath := filepath.Join(dst, relPath)

		// Create destination directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}

		// Check if destination file exists
		if !overwrite {
			if _, err := os.Stat(dstPath); err == nil {
				return fmt.Errorf("destination file already exists: %s", dstPath)
			} else if !os.IsNotExist(err) {
				return err
			}
		}

		// Copy the file
		return copyFile(path, dstPath)
	})
}

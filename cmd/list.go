package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"backup_ollama/internal/utils"
)

var (
	outputFormat string
	showDetails  bool
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available Ollama models",
	Long: `This command lists all models available in the Ollama directory 
(~/.ollama/models/manifests/{registry}/library/{model}/{version}).

It provides information about registries, models, versions, and optionally
detailed information from the version JSON files.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := listModels(outputFormat, showDetails); err != nil {
			fmt.Fprintf(os.Stderr, "Error listing models: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	listCmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text, json)")
	listCmd.Flags().BoolVarP(&showDetails, "details", "d", false, "Show detailed information from version files")
}

// listModels enumerates and displays Ollama models
func listModels(format string, details bool) error {
	modelList, err := utils.EnumerateOllamaModels()
	if err != nil {
		return fmt.Errorf("failed to enumerate models: %w", err)
	}

	// Count totals for summary
	var totalRegistries, totalModels, totalVersions int
	totalRegistries = len(modelList.Registries)
	for _, registry := range modelList.Registries {
		totalModels += len(registry.Models)
		for _, model := range registry.Models {
			totalVersions += len(model.Versions)
		}
	}

	switch strings.ToLower(format) {
	case "json":
		return outputJSON(modelList)
	case "text":
		fallthrough
	default:
		return outputText(modelList, details, totalRegistries, totalModels, totalVersions)
	}
}

// outputJSON formats the model list as JSON
func outputJSON(modelList *utils.OllamaModelList) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(modelList); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	return nil
}

// outputText formats the model list as text with tables
func outputText(modelList *utils.OllamaModelList, details bool, totalRegs, totalMods, totalVers int) error {
	// Create a new tabwriter
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	// Print summary
	fmt.Fprintf(w, "Found %d registries, %d models, %d versions\n\n", totalRegs, totalMods, totalVers)

	if totalRegs == 0 {
		fmt.Fprintf(w, "No models found in ~/.ollama/models/manifests/\n")
		return nil
	}

	for _, registry := range modelList.Registries {
		fmt.Fprintf(w, "Registry: %s\n", registry.Name)
		fmt.Fprintf(w, "  %-30s\t%-15s\t%-40s\n", "MODEL", "VERSIONS", "DIGEST")
		fmt.Fprintf(w, "  %-30s\t%-15s\t%-40s\n", "-----", "--------", "------")

		for _, model := range registry.Models {
			// Print the first version with the model name
			firstVersion := model.Versions[0]
			fmt.Fprintf(w, "  %-30s\t%-15s\t%-40s\n",
				model.Name,
				firstVersion.Name,
				truncateString(firstVersion.Digest, 40))

			// Print the rest of the versions with indentation
			for i := 1; i < len(model.Versions); i++ {
				version := model.Versions[i]
				fmt.Fprintf(w, "  %-30s\t%-15s\t%-40s\n",
					"",
					version.Name,
					truncateString(version.Digest, 40))
			}

			// If details are requested, print them for each version
			if details {
				for _, version := range model.Versions {
					fmt.Fprintf(w, "    Version: %s\n", version.Name)
					fmt.Fprintf(w, "    Path: %s\n", version.Path)
					fmt.Fprintf(w, "    Size: %d bytes\n", version.Size)
					fmt.Fprintf(w, "    Digest: %s\n", version.Digest)

					// Print any other interesting details from the version file
					if family, ok := version.Details["family"].(string); ok {
						fmt.Fprintf(w, "    Family: %s\n", family)
					}
					if license, ok := version.Details["license"].(string); ok {
						fmt.Fprintf(w, "    License: %s\n", license)
					}
					fmt.Fprintf(w, "\n")
				}
			}

			fmt.Fprintf(w, "\n")
		}
		fmt.Fprintf(w, "\n")
	}

	return w.Flush()
}

// truncateString truncates a string to maxLen and adds "..." if necessary
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

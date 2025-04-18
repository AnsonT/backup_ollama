# Contributing to backup_ollama

Thank you for your interest in contributing to the backup_ollama project! This document provides guidelines and instructions for contributing to the project.

## Project Overview

`backup_ollama` is a command-line tool written in Go that provides functionality to back up and restore models. The project structure is organized as follows:

- `main.go`: Entry point of the application
- `cmd/`: Contains the command implementations using Cobra
  - `root.go`: Defines the root command
  - `backup.go`: Implements the backup command
  - `restore.go`: Implements the restore command
- `internal/`: Contains internal packages
  - `utils/`: Utility functions
    - `paths.go`: Path handling utilities

## Development Setup

### Prerequisites

1. Go 1.16 or higher
2. Git

### Setting Up the Development Environment

1. Clone the repository:
   ```
   git clone https://github.com/your-username/backup_ollama.git
   cd backup_ollama
   ```

2. Install dependencies:
   ```
   go mod download
   ```

## Building and Testing

### Building the Project

To build the project locally:

```
go build -o backup_ollama
```

This will create an executable named `backup_ollama` in your current directory.

### Running the Application

After building, you can run the application:

```
./backup_ollama [command] [args]
```

Examples:
```
./backup_ollama backup my_model
./backup_ollama restore my_model --backup-dir /path/to/backup
```

### Running Tests

To run tests:

```
go test ./...
```

## Development Workflow

1. Create a new branch for your feature or bug fix:
   ```
   git checkout -b feature/your-feature-name
   ```
   or
   ```
   git checkout -b fix/issue-description
   ```

2. Make your changes and commit them with clear, descriptive commit messages:
   ```
   git commit -m "Add feature: description of the feature"
   ```

3. Push your branch to the remote repository:
   ```
   git push origin feature/your-feature-name
   ```

4. Create a pull request against the main branch.

## Code Style and Guidelines

1. Follow standard Go coding conventions and best practices.
2. Use `gofmt` to format your code before committing.
3. Add appropriate comments and documentation for functions and packages.
4. Update documentation when changing functionality.

## Creating a Release

To create a new release:

1. Update the version number in relevant files.
2. Create a tag for the new version:
   ```
   git tag -a v1.0.0 -m "Release version 1.0.0"
   ```
3. Push the tag:
   ```
   git push origin v1.0.0
   ```

4. Build binaries for different platforms:
   ```
   GOOS=linux GOARCH=amd64 go build -o backup_ollama-linux-amd64
   GOOS=darwin GOARCH=amd64 go build -o backup_ollama-darwin-amd64
   GOOS=windows GOARCH=amd64 go build -o backup_ollama-windows-amd64.exe
   ```

## Adding New Commands

The project uses [Cobra](https://github.com/spf13/cobra) for command-line functionality. To add a new command:

1. Create a new file in the `cmd` directory (e.g., `cmd/newcommmand.go`).
2. Define your command structure and functionality.
3. Add your command to the root command in `cmd/root.go`.

Example:
```go
var newCmd = &cobra.Command{
    Use:   "new [argument]",
    Short: "Short description",
    Long:  `Longer description about what the command does.`,
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        // Implementation
    },
}

func init() {
    rootCmd.AddCommand(newCmd)
    // Add any command-specific flags
    newCmd.Flags().StringP("flag-name", "f", "default", "Description of flag")
}
```

## Questions and Support

If you have questions or need help, please open an issue on the repository.

Thank you for contributing to backup_ollama!
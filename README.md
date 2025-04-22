# backup_ollama

## Overview

`backup_ollama` is a command-line application designed to facilitate the backup and restoration of Ollama models. It provides commands for backing up, restoring, and listing models, allowing users to manage their models efficiently.

## Commands

### List

The `list` command displays all available Ollama models in your installation.

**Usage:**

``` bash
backup_ollama list [flags]
```

**Flags:**

- `--output`, `-o` - Output format (text, json) [default: "text"]
- `--details`, `-d` - Show detailed information from version files [default: false]

### Backup

The `backup` command creates a backup of a specified Ollama model.

**Usage:**

``` bash
backup_ollama backup [model name] [flags]
```

**Arguments:**

- `[model name]` - Name of the model to back up (format: model:version, version is optional if only one exists)

**Flags:**

- `--dir`, `-d` - Directory to save the backup [default: "./backup"]
- `--zip`, `-z` - Create a zip file of the backup and delete the original directory [default: false]

### Restore

The `restore` command restores an Ollama model from a backup.

**Usage:**

``` bash
backup_ollama restore [model name] [flags]
```

**Arguments:**

- `[model name]` - Name of the backup directory or zip file to restore from

**Flags:**

- `--backup-dir`, `-d` - Directory containing the backup [default: "./backup"]
- `--ollama-dir` - Ollama directory to restore data to [default: user's Ollama directory]
- `--overwrite`, `-o` - Overwrite existing files during restore [default: false]

## Installation

To install the application, clone the repository and run the following command in the project directory:

``` bash
go build
```

This will create a binary named `backup_ollama` that you can run directly.

## Examples

### Listing Models

1. List all available models in text format:

   ``` bash
   backup_ollama list
   ```

2. List models with detailed information:

   ``` bash
   backup_ollama list --details
   ```

3. List models in JSON format:

   ``` bash
   backup_ollama list --output json
   ```

### Backing Up Models

1. To back up a model named `llama2` to the default directory:

   ``` bash
   backup_ollama backup llama2
   ```

2. To back up a specific version of a model:

   ``` bash
   backup_ollama backup llama2:7b
   ```

3. To back up a model to a specific directory:

   ``` bash
   backup_ollama backup llama2 --dir /path/to/backup
   ```

4. To back up a model and create a zip file:

   ``` bash
   backup_ollama backup llama2 --zip
   ```

### Restoring Models

1. To restore a model from the default backup directory:

   ``` bash
   backup_ollama restore llama2--7b--backup-1714404783
   ```

2. To restore a model from a specific backup directory:

   ``` bash
   backup_ollama restore llama2--7b--backup-1714404783 --backup-dir /path/to/backup
   ```

3. To restore from a zip file:

   ``` bash
   backup_ollama restore llama2--7b--backup-1714404783.zip --backup-dir /path/to/backup
   ```

4. To overwrite existing files when restoring:

   ``` bash
   backup_ollama restore llama2--7b--backup-1714404783 --overwrite
   ```

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue for any enhancements or bug fixes.

## License

This project is licensed under the MIT License. See the LICENSE file for more details.

> **Note**: This repository was entirely generated using VSCode AI Agent powered by Claude 3.7 Sonnet as an experiment.

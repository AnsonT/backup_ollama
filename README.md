# backup_ollama

## Overview
`backup_ollama` is a command-line application designed to facilitate the backup and restoration of models. It provides two primary commands: `backup` and `restore`, allowing users to manage their models efficiently.

## Commands

### Backup
The `backup` command is used to create a backup of a specified model.

**Usage:**
```
backup_ollama backup <model_name> [--backup-dir <directory>]
```

- `<model_name>`: The name of the model to back up.
- `--backup-dir <directory>`: (Optional) The directory where the backup will be stored. If not specified, the default backup directory will be used.

### Restore
The `restore` command is used to restore a specified model from a backup.

**Usage:**
```
backup_ollama restore <model_name> [--backup-dir <directory>]
```

- `<model_name>`: The name of the model to restore.
- `--backup-dir <directory>`: (Optional) The directory from which the backup will be restored. If not specified, the default backup directory will be used.

## Installation
To install the application, clone the repository and run the following command in the project directory:

```
go build
```

## Examples

1. To back up a model named `my_model` to the default directory:
   ```
   backup_ollama backup my_model
   ```

2. To back up a model named `my_model` to a specific directory:
   ```
   backup_ollama backup my_model --backup-dir /path/to/backup
   ```

3. To restore a model named `my_model` from the default directory:
   ```
   backup_ollama restore my_model
   ```

4. To restore a model named `my_model` from a specific directory:
   ```
   backup_ollama restore my_model --backup-dir /path/to/backup
   ```

## Contributing
Contributions are welcome! Please feel free to submit a pull request or open an issue for any enhancements or bug fixes.

## License
This project is licensed under the MIT License. See the LICENSE file for more details.
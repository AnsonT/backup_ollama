# GitHub Actions Release Setup

This document explains how to set up GitHub Actions for building, signing, and releasing the `backup_ollama` tool.

## Overview

The GitHub Actions workflow in `.github/workflows/release.yml` automates the process of:

1. Building the application for Linux, macOS, and Windows
2. Code signing the binaries (optional)
3. Generating checksums
4. Attaching all artifacts to GitHub releases

## Triggering a Release

The workflow automatically runs when you:

1. Create a new release through the GitHub web interface
2. Manually trigger the workflow using the "Run workflow" button on the Actions page

## Setting Up Code Signing

To enable code signing for your releases, you'll need to add the following secrets to your GitHub repository:

### macOS Signing

For signing macOS binaries, add these secrets:

- `APPLE_CERTIFICATE_BASE64`: Your Apple Developer certificate in Base64-encoded format
- `APPLE_CERTIFICATE_PASSWORD`: The password for your certificate
- `KEYCHAIN_PASSWORD`: A password for the temporary keychain (can be any secure string)
- `APPLE_IDENTITY`: The identity string from your certificate (e.g., "Developer ID Application: Your Name (TEAM_ID)")

To convert your certificate to Base64:

```bash
base64 -i YourCertificate.p12 -o certificate-base64.txt
```

### Windows Signing

For signing Windows binaries, add these secrets:

- `WINDOWS_CERTIFICATE_BASE64`: Your code signing certificate in Base64-encoded format
- `WINDOWS_CERTIFICATE_PASSWORD`: The password for your certificate

To convert your certificate to Base64:

```bash
certutil -encode YourCertificate.pfx certificate-base64.txt
```

Then open the file and remove the header and footer lines, keeping only the Base64 content.

### Linux Signing (GPG)

For signing Linux binaries with GPG, add these secrets:

- `GPG_PRIVATE_KEY`: Your exported GPG private key
- `GPG_PASSPHRASE`: The passphrase for your GPG key

To export your GPG private key:

```bash
gpg --export-secret-keys YOUR_KEY_ID | base64 > gpg-private-key.txt
```

## Generate a Release

To create a new release:

1. Go to your repository's "Releases" section
2. Click "Draft a new release"
3. Create a new tag (e.g., `v1.0.0`)
4. Fill in the release title and description
5. Click "Publish release"

The GitHub Actions workflow will automatically build and attach the binaries and checksums to the release.

## Manual Workflow Execution

You can manually trigger the workflow:

1. Go to the "Actions" tab in your repository
2. Select the "Build and Release" workflow
3. Click "Run workflow"
4. Enter the tag name and click "Run workflow"

## Verification

After the workflow completes:

1. Check your release to verify all artifacts are attached
2. Use the checksums to verify file integrity
3. For signed binaries:
   - macOS: Run `codesign -vvv --deep --strict backup_ollama`
   - Windows: Use tools like "sigcheck" from Sysinternals
   - Linux: Use `gpg --verify backup_ollama.asc backup_ollama`
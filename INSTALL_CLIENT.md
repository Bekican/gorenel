# Client Install Guide

## Magic Install (Recommended)

To install the Gorenel CLI, we recommend our fetch-inspect-run sequence. This gives you complete transparency over the script before execution.

### Windows (PowerShell)
Run the following in PowerShell to download, view, and run the installer:
```powershell
iwr -useb https://gorenel.site/install.ps1 -OutFile install.ps1
Get-Content install.ps1
.\install.ps1
```

### Linux / macOS (Bash)
Run the following in Terminal to download, view, and run the installer:
```bash
curl -fsSL https://gorenel.site/install.sh -o install.sh
cat install.sh
bash install.sh
```

*Note: The scripts automatically verify the binary's SHA256 checksum during execution.*

## Manual Configuration

If you prefer to configure the CLI manually:
```bash
gorenel config set api_key gk_xxx
gorenel config set port 3000
gorenel config set server wss://gorenel.site/tunnel/connect
```

Then connect:
```bash
gorenel connect
```

## Build & Packaging

### Windows Installer (Inno Setup)
1. Ensure client binary exists at `bin/gorenel-windows-amd64.exe`
2. Open `packaging/windows/gorenel.iss`
3. Build installer

### macOS Homebrew
1. Update `packaging/homebrew/gorenel.rb` with current SHA256 hashes.
2. Users install with: `brew install <your-tap>/gorenel`

## Security Verification
All official binaries are signed with SHA256. You can manually verify them using:
- **Windows**: `Get-FileHash -Algorithm SHA256 ./gorenel.exe`
- **Linux/macOS**: `sha256sum ./gorenel`

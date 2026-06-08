# ==============================================================================
#           Gorenel CLI Client Windows Installation Script
# ==============================================================================
# This script securely downloads and installs the official Gorenel CLI client
# for Windows. It configures the user's PATH environment variable for seamless use.
# ==============================================================================

$ErrorActionPreference = "Stop"

$VERSION = "1.2.5"
$REPO = "Bekican/gorenel"
$INSTALL_DIR = Join-Path $env:USERPROFILE ".gorenel\bin"
$BINARY_NAME = "gorenel-client-windows-amd64.exe"
$DOWNLOAD_URL = "https://github.com/$REPO/releases/download/v$VERSION/$BINARY_NAME"

Write-Host "🔷 Starting Gorenel CLI installation v$VERSION..." -ForegroundColor Cyan

# 1. Create installation directory
if (-not (Test-Path $INSTALL_DIR)) {
    New-Item -ItemType Directory -Path $INSTALL_DIR | Out-Null
    Write-Host "📁 Created installation directory at $INSTALL_DIR" -ForegroundColor Gray
}

# 2. Download CLI binary
$TEMP_FILE = [System.IO.Path]::GetTempFileName()
Write-Host "📡 Downloading Gorenel CLI from GitHub..." -ForegroundColor Gray
try {
    Invoke-WebRequest -Uri $DOWNLOAD_URL -OutFile $TEMP_FILE -UseBasicParsing
} catch {
    Write-Error "❌ Error: Failed to download binary from $DOWNLOAD_URL"
    exit 1
}

# 3. Move binary to installation folder
$FINAL_PATH = Join-Path $INSTALL_DIR "gorenel.exe"
Move-Item -Path $TEMP_FILE -Destination $FINAL_PATH -Force
Write-Host "🛡️ Installed Gorenel CLI executable to $FINAL_PATH" -ForegroundColor Gray

# 4. Add to PATH for the current user session
$USER_PATH = [Environment]::GetEnvironmentVariable("Path", "User")
if ($USER_PATH -notlike "*$INSTALL_DIR*") {
    $NEW_PATH = "$USER_PATH;$INSTALL_DIR"
    [Environment]::SetEnvironmentVariable("Path", $NEW_PATH, "User")
    Write-Host "⚙️ Added $INSTALL_DIR to User PATH." -ForegroundColor Gray
    Write-Host "🔔 Note: Please restart your terminal/PowerShell window to refresh environment variables." -ForegroundColor Yellow
}

Write-Host "✅ Gorenel CLI successfully installed!" -ForegroundColor Green
Write-Host "🚀 Run 'gorenel login' or 'gorenel http <port>' to get started!" -ForegroundColor Green

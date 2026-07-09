# Bullang Installer Bootstrap — Windows (PowerShell)
# Downloads Go if needed, clones the installer, compiles and launches it.

Set-ExecutionPolicy Bypass -Scope Process -Force

$ErrorActionPreference = "Stop"
$GoVersion    = "1.22.3"
$InstallerRepo = "https://github.com/The-Bullang-Foundation/bullang-installer"
$InstallDir   = "$env:USERPROFILE\.bullang-installer"

Write-Host ""
Write-Host "   ____        _ _"
Write-Host "  |  _ \      | | |"
Write-Host "  | |_) |_   _| | | __ _ _ __   __ _"
Write-Host "  |  _ <| | | | | |/ _`` | '_ \ / _`` |"
Write-Host "  | |_) | |_| | | | (_| | | | | (_| |"
Write-Host "  |____/ \__,_|_|_|\__,_|_| |_|\__, |"
Write-Host "                                 __/ |"
Write-Host "                                |___/"
Write-Host ""
Write-Host "  Bullang Ecosystem Installer"
Write-Host "  ─────────────────────────────────────────"
Write-Host ""

# ── Step 1: Install Go if missing ─────────────────────────────────────────────

$goCmd = Get-Command go -ErrorAction SilentlyContinue

if ($goCmd) {
    Write-Host "  ✓ Go is already installed: $(go version)"
} else {
    Write-Host "  → Installing Go $GoVersion..."

    $arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
    $goMsi = "go$GoVersion.windows-$arch.msi"
    $goUrl = "https://go.dev/dl/$goMsi"
    $tmpMsi = "$env:TEMP\$goMsi"

    Write-Host "  → Downloading $goUrl..."
    Invoke-WebRequest -Uri $goUrl -OutFile $tmpMsi

    Write-Host "  → Running Go installer..."
    Start-Process -Wait -FilePath "msiexec.exe" -ArgumentList "/i `"$tmpMsi`" /quiet /norestart"
    Remove-Item $tmpMsi

    # Refresh PATH
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path", "Machine") + ";" +
                [System.Environment]::GetEnvironmentVariable("Path", "User")

    Write-Host "  ✓ Go $GoVersion installed."
}

# ── Step 2: Install git if missing ────────────────────────────────────────────

$gitCmd = Get-Command git -ErrorAction SilentlyContinue
if (-not $gitCmd) {
    Write-Host "  → Installing git via winget..."
    winget install -e --id Git.Git --silent
    # Refresh PATH
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path", "Machine") + ";" +
                [System.Environment]::GetEnvironmentVariable("Path", "User")
}

# ── Step 3: Clone the installer ───────────────────────────────────────────────

Write-Host "  → Downloading Bullang installer..."
if (Test-Path $InstallDir) { Remove-Item -Recurse -Force $InstallDir }
git clone --depth 1 $InstallerRepo $InstallDir

# ── Step 4: Build and run ─────────────────────────────────────────────────────

Write-Host "  → Building installer (this may take a moment)..."
Set-Location $InstallDir
go mod tidy
go build -o bullang-installer.exe .

Write-Host "  → Launching installer..."
Start-Process -Wait -FilePath ".\bullang-installer.exe"

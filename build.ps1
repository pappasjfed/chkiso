<#
.SYNOPSIS
    Build script for chkiso - compiles PowerShell script to executable.

.DESCRIPTION
    This script compiles chkiso.ps1 to chkiso.exe and places it in the bin/ directory.
    It also generates a SHA256 checksum file.

.PARAMETER Clean
    Clean the bin directory before building.

.PARAMETER SkipChecksum
    Skip generating the SHA256 checksum file.

.EXAMPLE
    .\build.ps1
    Builds chkiso.exe in the bin/ directory

.EXAMPLE
    .\build.ps1 -Clean
    Cleans bin/ directory and builds
#>

[CmdletBinding()]
param(
    [switch]$Clean,
    [switch]$SkipChecksum
)

$ErrorActionPreference = 'Stop'

Write-Host "=== chkiso Build Script ===" -ForegroundColor Cyan

# Check if ps2exe is installed
Write-Host "`nChecking for ps2exe module..." -ForegroundColor Yellow
$ps2exeModule = Get-Module -ListAvailable -Name ps2exe
if (-not $ps2exeModule) {
    Write-Host "ps2exe module not found. Installing..." -ForegroundColor Yellow
    try {
        Install-Module -Name ps2exe -Force -Scope CurrentUser -AllowClobber
        Write-Host "✓ ps2exe installed successfully" -ForegroundColor Green
    } catch {
        Write-Error "Failed to install ps2exe: $_"
        Write-Host "`nTo install manually, run: Install-Module -Name ps2exe -Force" -ForegroundColor Yellow
        exit 1
    }
}

Import-Module ps2exe -ErrorAction Stop
Write-Host "✓ ps2exe module loaded" -ForegroundColor Green

# Create or clean bin directory
if ($Clean -and (Test-Path bin)) {
    Write-Host "`nCleaning bin directory..." -ForegroundColor Yellow
    Remove-Item bin/*.exe -ErrorAction SilentlyContinue
    Remove-Item bin/*.sha -ErrorAction SilentlyContinue
    Write-Host "✓ bin directory cleaned" -ForegroundColor Green
}

if (-not (Test-Path bin)) {
    Write-Host "`nCreating bin directory..." -ForegroundColor Yellow
    New-Item -ItemType Directory -Path bin | Out-Null
    Write-Host "✓ bin directory created" -ForegroundColor Green
}

# Compile the executable
Write-Host "`nCompiling chkiso.ps1 to bin/chkiso.exe..." -ForegroundColor Yellow
try {
    ps2exe -inputFile chkiso.ps1 -outputFile bin/chkiso.exe `
        -noConsole:$false -title "chkiso" -version "1.0.0.0" `
        -company "chkiso" -product "chkiso" -copyright "MIT License"
    
    if (-not (Test-Path bin/chkiso.exe)) {
        throw "bin/chkiso.exe was not created"
    }
    
    $fileSize = (Get-Item bin/chkiso.exe).Length
    $fileSizeMB = [math]::Round($fileSize / 1MB, 2)
    Write-Host "✓ Successfully created bin/chkiso.exe ($fileSizeMB MB)" -ForegroundColor Green
    
} catch {
    Write-Error "Failed to compile: $_"
    exit 1
}

# Generate SHA256 checksum
if (-not $SkipChecksum) {
    Write-Host "`nGenerating SHA256 checksum..." -ForegroundColor Yellow
    try {
        $hash = (Get-FileHash -Path bin/chkiso.exe -Algorithm SHA256).Hash.ToLower()
        "$hash  chkiso.exe" | Out-File -FilePath "bin/chkiso.exe.sha" -Encoding utf8NoBOM
        Write-Host "✓ SHA256 checksum saved to bin/chkiso.exe.sha" -ForegroundColor Green
        Write-Host "  Hash: $hash" -ForegroundColor Gray
    } catch {
        Write-Warning "Failed to generate checksum: $_"
    }
}

Write-Host "`n=== Build Complete ===" -ForegroundColor Cyan
Write-Host "`nOutput files:" -ForegroundColor Yellow
Get-ChildItem bin/ | ForEach-Object {
    Write-Host "  - bin/$($_.Name)" -ForegroundColor Gray
}

Write-Host "`nTo test the executable, run:" -ForegroundColor Yellow
Write-Host "  .\bin\chkiso.exe test\test.iso" -ForegroundColor White

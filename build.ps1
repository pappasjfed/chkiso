<#
.SYNOPSIS
    Build script for chkiso tools - compiles PowerShell scripts to executables.

.DESCRIPTION
    This script compiles the PowerShell scripts to executables and places them in the bin/ directory.
    - sha256media.ps1 -> sha256media.exe (hash calculator)
    - verifymedia.ps1 -> verifymedia.exe (content verifier)
    It also generates SHA256 checksum files for each executable.

.PARAMETER Clean
    Clean the bin directory before building.

.PARAMETER SkipChecksum
    Skip generating the SHA256 checksum files.

.EXAMPLE
    .\build.ps1
    Builds all executables in the bin/ directory

.EXAMPLE
    .\build.ps1 -Clean
    Cleans bin/ directory and builds all executables
#>

[CmdletBinding()]
param(
    [switch]$Clean,
    [switch]$SkipChecksum
)

$ErrorActionPreference = 'Stop'

Write-Host "=== chkiso Build Script ===" -ForegroundColor Cyan

# Function to download checkisomd5.exe from GitHub releases
function Get-CheckIsoMd5 {
    Write-Host "`nDownloading checkisomd5.exe..." -ForegroundColor Yellow
    try {
        # Get latest release info from GitHub API
        $releasesUrl = "https://api.github.com/repos/pappasjfed/isomd5sum/releases/latest"
        $release = Invoke-RestMethod -Uri $releasesUrl -ErrorAction Stop
        
        # Find checkisomd5.exe asset
        $asset = $release.assets | Where-Object { $_.name -eq "checkisomd5.exe" } | Select-Object -First 1
        
        if (-not $asset) {
            Write-Warning "checkisomd5.exe not found in latest release, trying to find any release with it..."
            $allReleasesUrl = "https://api.github.com/repos/pappasjfed/isomd5sum/releases"
            $allReleases = Invoke-RestMethod -Uri $allReleasesUrl -ErrorAction Stop
            
            foreach ($rel in $allReleases) {
                $asset = $rel.assets | Where-Object { $_.name -eq "checkisomd5.exe" } | Select-Object -First 1
                if ($asset) {
                    Write-Host "  Found in release: $($rel.tag_name)" -ForegroundColor Gray
                    break
                }
            }
        }
        
        if (-not $asset) {
            Write-Warning "Could not find checkisomd5.exe in any release"
            return $false
        }
        
        $downloadUrl = $asset.browser_download_url
        $outputPath = "bin/checkisomd5.exe"
        
        Write-Host "  Downloading from: $downloadUrl" -ForegroundColor Gray
        Invoke-WebRequest -Uri $downloadUrl -OutFile $outputPath -ErrorAction Stop
        
        if (Test-Path $outputPath) {
            $fileSize = (Get-Item $outputPath).Length
            $fileSizeKB = [math]::Round($fileSize / 1KB, 2)
            Write-Host "✓ Downloaded checkisomd5.exe ($fileSizeKB KB)" -ForegroundColor Green
            return $true
        }
        return $false
    }
    catch {
        Write-Warning "Failed to download checkisomd5.exe: $_"
        return $false
    }
}

# Function to download sha256sum.exe (GnuWin32 CoreUtils)
function Get-Sha256Sum {
    Write-Host "`nDownloading sha256sum.exe..." -ForegroundColor Yellow
    try {
        # GnuWin32 CoreUtils contains sha256sum
        # We'll download the complete package and extract sha256sum.exe
        $coreUtilsZipUrl = "https://sourceforge.net/projects/gnuwin32/files/coreutils/5.3.0/coreutils-5.3.0-bin.zip/download"
        $tempZip = [System.IO.Path]::GetTempFileName() + ".zip"
        $tempExtract = [System.IO.Path]::Combine([System.IO.Path]::GetTempPath(), "coreutils_extract_" + [guid]::NewGuid().ToString())
        
        Write-Host "  Downloading GnuWin32 CoreUtils package..." -ForegroundColor Gray
        Write-Host "  (This may take a moment as it's ~1.5MB)" -ForegroundColor Gray
        
        # Note: SourceForge redirects, so we need to allow redirects
        Invoke-WebRequest -Uri $coreUtilsZipUrl -OutFile $tempZip -MaximumRedirection 5 -ErrorAction Stop
        
        Write-Host "  Extracting sha256sum.exe..." -ForegroundColor Gray
        Expand-Archive -Path $tempZip -DestinationPath $tempExtract -Force
        
        # Find sha256sum.exe in extracted files (typically in bin/ subdirectory)
        $sha256sumExe = Get-ChildItem -Path $tempExtract -Filter "sha256sum.exe" -Recurse | Select-Object -First 1
        
        if ($sha256sumExe) {
            Copy-Item -Path $sha256sumExe.FullName -Destination "bin/sha256sum.exe" -Force
            $fileSize = (Get-Item "bin/sha256sum.exe").Length
            $fileSizeKB = [math]::Round($fileSize / 1KB, 2)
            Write-Host "✓ Downloaded sha256sum.exe ($fileSizeKB KB)" -ForegroundColor Green
            
            # Also copy required DLLs if they exist
            $requiredDlls = @("libiconv2.dll", "libintl3.dll")
            foreach ($dll in $requiredDlls) {
                $dllFile = Get-ChildItem -Path $tempExtract -Filter $dll -Recurse | Select-Object -First 1
                if ($dllFile) {
                    Copy-Item -Path $dllFile.FullName -Destination "bin/$dll" -Force
                    Write-Host "  ✓ Copied dependency: $dll" -ForegroundColor Gray
                }
            }
            
            # Cleanup
            Remove-Item -Path $tempZip -Force -ErrorAction SilentlyContinue
            Remove-Item -Path $tempExtract -Recurse -Force -ErrorAction SilentlyContinue
            return $true
        }
        else {
            Write-Warning "sha256sum.exe not found in extracted package"
            Remove-Item -Path $tempZip -Force -ErrorAction SilentlyContinue
            Remove-Item -Path $tempExtract -Recurse -Force -ErrorAction SilentlyContinue
            return $false
        }
    }
    catch {
        Write-Warning "Failed to download sha256sum.exe: $_"
        Write-Host "  Note: You can manually download from https://sourceforge.net/projects/gnuwin32/files/coreutils/" -ForegroundColor Gray
        Write-Host "  Or install Git for Windows which includes sha256sum" -ForegroundColor Gray
        return $false
    }
}

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

# Array of scripts to compile
$scriptsToCompile = @(
    @{
        Script = "sha256media.ps1"
        Output = "bin/sha256media.exe"
        Title = "sha256media"
        Description = "SHA256 Media Hash Calculator"
    },
    @{
        Script = "verifymedia.ps1"
        Output = "bin/verifymedia.exe"
        Title = "verifymedia"
        Description = "Media Content Verifier"
    }
)

$compiledCount = 0
$failedCompilations = @()

# Compile each executable
foreach ($script in $scriptsToCompile) {
    Write-Host "`nCompiling $($script.Script) to $($script.Output)..." -ForegroundColor Yellow
    try {
        ps2exe -inputFile $script.Script -outputFile $script.Output `
            -noConsole:$false -title $script.Title -version "2.0.0.0" `
            -company "chkiso" -product $script.Title -copyright "MIT License"
        
        if (-not (Test-Path $script.Output)) {
            throw "$($script.Output) was not created"
        }
        
        $fileSize = (Get-Item $script.Output).Length
        $fileSizeMB = [math]::Round($fileSize / 1MB, 2)
        Write-Host "✓ Successfully created $($script.Output) ($fileSizeMB MB)" -ForegroundColor Green
        $compiledCount++
        
        # Generate SHA256 checksum
        if (-not $SkipChecksum) {
            try {
                $hash = (Get-FileHash -Path $script.Output -Algorithm SHA256).Hash.ToLower()
                $exeName = [System.IO.Path]::GetFileName($script.Output)
                "$hash  $exeName" | Out-File -FilePath "$($script.Output).sha" -Encoding utf8NoBOM
                Write-Host "  ✓ SHA256 checksum saved to $($script.Output).sha" -ForegroundColor Gray
            } catch {
                Write-Warning "Failed to generate checksum for $($script.Output): $_"
            }
        }
    } catch {
        Write-Error "Failed to compile $($script.Script): $_"
        $failedCompilations += $script.Script
    }
}

Write-Host "`n=== Build Complete ===" -ForegroundColor Cyan

if ($failedCompilations.Count -gt 0) {
    Write-Host "`nFailed to compile:" -ForegroundColor Red
    foreach ($failed in $failedCompilations) {
        Write-Host "  - $failed" -ForegroundColor Red
    }
}

if ($compiledCount -gt 0) {
    Write-Host "`nSuccessfully compiled $compiledCount executable(s):" -ForegroundColor Green
    Write-Host "`nOutput files:" -ForegroundColor Yellow
    Get-ChildItem bin/ -Filter *.exe | ForEach-Object {
        Write-Host "  - bin/$($_.Name)" -ForegroundColor Gray
    }
    
    if (-not $SkipChecksum) {
        Write-Host "`nChecksum files:" -ForegroundColor Yellow
        Get-ChildItem bin/ -Filter *.sha | ForEach-Object {
            Write-Host "  - bin/$($_.Name)" -ForegroundColor Gray
        }
    }
    
    Write-Host "`nTo test the executables:" -ForegroundColor Yellow
    Write-Host "  .\bin\sha256media.exe test\test.iso" -ForegroundColor White
    Write-Host "  .\bin\verifymedia.exe test\test.iso" -ForegroundColor White
} else {
    Write-Error "No executables were compiled successfully"
    exit 1
}
Get-ChildItem bin/ | ForEach-Object {
    Write-Host "  - bin/$($_.Name)" -ForegroundColor Gray
}

Write-Host "`nTo test the executable, run:" -ForegroundColor Yellow
Write-Host "  .\bin\chkiso.exe test\test.iso" -ForegroundColor White

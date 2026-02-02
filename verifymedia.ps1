<#
.SYNOPSIS
    Verify files on mounted ISO or physical media against hash files.

.DESCRIPTION
    This utility scans a mounted ISO or physical drive for all .sha files
    (also sha256sum.txt, SHA256SUMS) and verifies each file against its hash.
    Hash files should be in Linux sha256sum format: <hash>  <filename>

.PARAMETER Path
    The path to the ISO file or the drive letter to verify (e.g., "D:" or "image.iso").

.PARAMETER Recursive
    Search for hash files recursively (default behavior).

.PARAMETER NoRecursive
    Only search for hash files in the root directory.

.EXAMPLE
    .\verifymedia.ps1 D:\image.iso
    Mounts the ISO and verifies all files against hash files found.

.EXAMPLE
    .\verifymedia.ps1 G:
    Verifies files on drive G: against hash files found.
#>

[CmdletBinding()]
param (
    [Parameter(Mandatory=$true, Position=0)]
    [string]$Path,

    [Parameter(Mandatory=$false)]
    [switch]$NoRecursive
)

# --- Error Tracking ---
$script:hasErrors = $false
$script:mountedDriveLetter = $null

# --- Helper Functions ---

function Get-Sha256sumPath {
    <#
    .SYNOPSIS
        Checks if sha256sum.exe is available and returns its path.
    .OUTPUTS
        String path to sha256sum.exe if found, $null otherwise.
    #>
    # Check for sha256sum.exe in current directory first
    if (Test-Path ".\sha256sum.exe") {
        return (Resolve-Path ".\sha256sum.exe").Path
    }
    # Check if sha256sum.exe is in PATH
    elseif ($sha256sumCmd = Get-Command sha256sum.exe -ErrorAction SilentlyContinue) {
        return $sha256sumCmd.Source
    }
    # Check common Git for Windows installation locations
    else {
        # Try system-wide Git for Windows installation (64-bit)
        $gitPath = "C:\Program Files\Git\usr\bin\sha256sum.exe"
        if (Test-Path $gitPath -ErrorAction SilentlyContinue) {
            return $gitPath
        }
        
        # Try user-local Git for Windows installation
        if ($env:LOCALAPPDATA) {
            $userGitPath = Join-Path $env:LOCALAPPDATA "Programs\Git\usr\bin\sha256sum.exe"
            if (Test-Path $userGitPath -ErrorAction SilentlyContinue) {
                return $userGitPath
            }
        }
        
        # Try 32-bit Git for Windows installation
        $git32Path = "C:\Program Files (x86)\Git\usr\bin\sha256sum.exe"
        if (Test-Path $git32Path -ErrorAction SilentlyContinue) {
            return $git32Path
        }
    }
    return $null
}

function Invoke-Sha256sumUtility {
    <#
    .SYNOPSIS
        Invokes sha256sum.exe utility to calculate hash, with fallback on failure.
    .PARAMETER FilePath
        Path to the file to hash.
    .PARAMETER Sha256sumPath
        Path to the sha256sum.exe utility.
    .OUTPUTS
        Hashtable with 'Success' (bool) and 'Hash' (string) if successful.
    #>
    param(
        [string]$FilePath,
        [string]$Sha256sumPath
    )
    
    try {
        # Run sha256sum and capture output
        $output = & $Sha256sumPath $FilePath 2>&1
        $exitCode = $LASTEXITCODE
        
        if ($exitCode -eq 0 -and $output) {
            # Parse sha256sum output format: "<hash>  <filename>" or "<hash> <filename>"
            $outputStr = $output | Out-String
            if ($outputStr -match '^([a-fA-F0-9]{64})\s+') {
                return @{ Success = $true; Hash = $matches[1].ToLower() }
            }
        }
        
        return @{ Success = $false }
    }
    catch {
        return @{ Success = $false; Error = $_ }
    }
}

function Get-Sha256Hash {
    <#
    .SYNOPSIS
        Calculates SHA256 hash of a file using sha256sum.exe if available, or FileStream (ps2exe compatible).
    .PARAMETER FilePath
        Path to the file to hash.
    .PARAMETER Quiet
        If specified, suppresses the progress message.
    #>
    param(
        [Parameter(Mandatory=$true)]
        [string]$FilePath,
        [switch]$Quiet
    )
    
    # Check if sha256sum.exe utility is available (Windows)
    $sha256sumPath = Get-Sha256sumPath
    
    # If sha256sum.exe is found, try to use it
    if ($sha256sumPath) {
        # Verify file exists before attempting to display its name
        if (-not (Test-Path $FilePath -PathType Leaf)) {
            Write-Error "File not found: $FilePath"
            return $null
        }
        
        $result = Invoke-Sha256sumUtility -FilePath $FilePath -Sha256sumPath $sha256sumPath
        
        if ($result.Success) {
            return $result.Hash
        }
    }
    
    # Use built-in SHA256 calculation if utility not available or failed
    $sha = [System.Security.Cryptography.SHA256]::Create()
    $stream = $null
    
    try {
        # Use FileStream for ps2exe compatibility (Get-FileHash not available in compiled exe)
        $stream = New-Object System.IO.FileStream($FilePath, [System.IO.FileMode]::Open, [System.IO.FileAccess]::Read, [System.IO.FileShare]::Read)
        
        $hashBytes = $sha.ComputeHash($stream)
        return [System.BitConverter]::ToString($hashBytes).Replace("-", "").ToLower()
    }
    finally {
        if ($stream) { $stream.Close() }
        if ($sha) { $sha.Dispose() }
    }
}

function Verify-Contents {
    param ([string]$MountPath)
    
    Write-Host "--- Verifying Contents ---" -ForegroundColor Cyan
    Write-Host "Scanning for hash files in: $MountPath" -ForegroundColor Green
    
    # Find checksum files
    if ($NoRecursive) {
        $checksumFiles = Get-ChildItem -Path $MountPath -Include "*.sha", "sha256sum.txt", "SHA256SUMS" -ErrorAction SilentlyContinue
    } else {
        $checksumFiles = Get-ChildItem -Path $MountPath -Recurse -Include "*.sha", "sha256sum.txt", "SHA256SUMS" -ErrorAction SilentlyContinue
    }
    
    if (-not $checksumFiles) {
        Write-Warning "Could not find any checksum files (*.sha, sha256sum.txt, SHA256SUMS) on the media."
        return
    }

    Write-Host "Found $($checksumFiles.Count) hash file(s)" -ForegroundColor Yellow
    Write-Host ""

    $totalFiles = 0
    $failedFiles = 0
    
    foreach ($checksumFile in $checksumFiles) {
        Write-Host "Processing checksum file: $($checksumFile.FullName)" -ForegroundColor Yellow
        $baseDir = $checksumFile.DirectoryName
        
        Get-Content -Path $checksumFile.FullName | ForEach-Object {
            # Parse sha256sum format: hash  filename
            # Also handle variations with * or ./ prefixes
            if ($_ -match "^([a-fA-F0-9]{64})\s+[\*.\/\\]*(.*)") {
                $totalFiles++
                $expectedHash = $matches[1].ToLower()
                $fileName = $matches[2].Trim()
                
                $filePathOnMedia = Join-Path -Path $baseDir -ChildPath $fileName
                
                if (-not (Test-Path $filePathOnMedia)) {
                    Write-Warning "File not found on media: $fileName (referenced in $($checksumFile.Name))"
                    $failedFiles++
                    return
                }
                
                Write-Host "Verifying: $fileName" -NoNewline
                $calculatedHash = Get-Sha256Hash -FilePath $filePathOnMedia -Quiet
                
                if ($calculatedHash -eq $expectedHash) {
                    Write-Host " -> " -NoNewline
                    Write-Host "OK" -ForegroundColor Green
                } else {
                    Write-Host " -> " -NoNewline
                    Write-Host "FAILED" -ForegroundColor Red
                    $failedFiles++
                    $script:hasErrors = $true
                }
            }
        }
        Write-Host ""
    }
    
    Write-Host "--- Verification Summary ---" -ForegroundColor Cyan
    if ($failedFiles -eq 0 -and $totalFiles -gt 0) {
        Write-Host "Success: All $totalFiles files verified." -ForegroundColor Green
    } elseif ($totalFiles -eq 0) {
        Write-Host "No files to verify." -ForegroundColor Yellow
    } else {
        Write-Host "Failure: $failedFiles out of $totalFiles files failed." -ForegroundColor Red
        $script:hasErrors = $true
    }
}

# --- Main Script Body ---

Write-Host "=== Media Content Verification ===" -ForegroundColor Cyan
Write-Host ""

# Validate path
$isDrive = $false
$driveLetter = ''
$ResolvedPath = $null
$diskImage = $null

if ($Path -match '^([A-Za-z]):\\?$') {
    $driveLetter = $Matches[1]
    try {
        $volume = Get-Volume -DriveLetter $driveLetter -ErrorAction Stop
        $isDrive = $true
        $ResolvedPath = $Path
        $mountPath = "$($driveLetter):\"
        Write-Host "Verifying contents of physical drive at: $mountPath" -ForegroundColor Green
        $script:mountedDriveLetter = $driveLetter
    } catch {
        Write-Error "Drive '$Path' not found or is not ready."
        exit 1
    }
} else {
    try {
        $ResolvedPath = (Resolve-Path -LiteralPath $Path.Trim("`"")).Path
    } catch {
        Write-Error "File not found: $Path"
        exit 1
    }
    
    if (-not (Test-Path $ResolvedPath -PathType Leaf)) {
        Write-Error "Path is not a file: $ResolvedPath"
        exit 1
    }
    
    # Mount ISO
    Write-Host "Mounting ISO..."
    try {
        $diskImage = Mount-DiskImage -ImagePath $ResolvedPath -PassThru
        $driveLetter = ($diskImage | Get-Volume).DriveLetter
        
        if ([string]::IsNullOrWhiteSpace($driveLetter)) {
            Write-Error "Could not get drive letter after mounting ISO."
            exit 1
        }
        
        $mountPath = "${driveLetter}:\"
        Write-Host "ISO mounted at: $mountPath" -ForegroundColor Green
        $script:mountedDriveLetter = $driveLetter
    } catch {
        Write-Error "Failed to mount ISO: $_"
        exit 1
    }
}

Write-Host ""

# Verify contents
try {
    Verify-Contents -MountPath $mountPath
} catch {
    Write-Error "An error occurred during verification: $_"
    $script:hasErrors = $true
} finally {
    # Dismount ISO if we mounted it
    if ($diskImage) {
        Write-Host ""
        Write-Host "Dismounting ISO..." -ForegroundColor Yellow
        try {
            Dismount-DiskImage -ImagePath $ResolvedPath -ErrorAction Stop
            Write-Host "ISO dismounted successfully." -ForegroundColor Green
        } catch {
            Write-Warning "Failed to dismount ISO: $_"
        }
    }
}

Write-Host ""
if ($script:hasErrors) {
    Write-Host "Verification completed with errors." -ForegroundColor Red
    exit 1
} else {
    Write-Host "Verification completed successfully." -ForegroundColor Green
    exit 0
}

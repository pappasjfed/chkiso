<#
.SYNOPSIS
    Calculate SHA256 hash of ISO files or physical drives.

.DESCRIPTION
    This utility calculates the SHA256 hash of an ISO file or physical drive.
    For drives, it uses Win32 device paths to read the raw device.
    Supports both regular PowerShell execution and ps2exe compiled executables.

.PARAMETER Path
    The path to the ISO file or the drive letter to hash (e.g., "D:" or "image.iso").

.PARAMETER Output
    Optional output file to save the hash in sha256sum format.

.EXAMPLE
    .\sha256media.ps1 D:\image.iso
    Calculates the hash of the ISO file.

.EXAMPLE
    .\sha256media.ps1 G:
    Calculates the hash of drive G: (PowerShell script only, not compiled exe).

.EXAMPLE
    .\sha256media.ps1 image.iso -Output image.iso.sha
    Calculates hash and saves to file in sha256sum format.
#>

[CmdletBinding()]
param (
    [Parameter(Mandatory=$true, Position=0)]
    [string]$Path,

    [Parameter(Mandatory=$false)]
    [Alias('Out', 'OutputFile')]
    [string]$Output
)

# --- Error Tracking ---
$script:hasErrors = $false
$script:isCompiledExe = $false

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
        
        if (-not $Quiet) {
            $fileName = (Get-Item $FilePath).Name
            Write-Host "Calculating SHA256 hash for file '$fileName' using sha256sum.exe..."
        }
        
        $result = Invoke-Sha256sumUtility -FilePath $FilePath -Sha256sumPath $sha256sumPath
        
        if ($result.Success) {
            return $result.Hash
        }
        
        # Fall back to built-in on failure
        if (-not $Quiet) {
            Write-Warning "sha256sum.exe utility failed, falling back to built-in hash calculation"
        }
    }
    
    # Use built-in SHA256 calculation if utility not available or failed
    $sha = [System.Security.Cryptography.SHA256]::Create()
    $stream = $null
    
    try {
        if (-not $Quiet) {
            $fileName = (Get-Item $FilePath).Name
            Write-Host "Calculating SHA256 hash for file '$fileName'..."
        }
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

function Get-Sha256FromPath {
    param( [string]$TargetPath, [bool]$IsDrive, [string]$DriveLetter )
    
    # For drives in compiled exe mode, cannot hash raw drives
    if ($IsDrive -and $script:isCompiledExe) {
        Write-Host "`n" -NoNewline
        Write-Host "ERROR: " -ForegroundColor Red -NoNewline
        Write-Host "Cannot hash drive letters with compiled executable" -ForegroundColor Yellow
        Write-Host ""
        Write-Host "Why this doesn't work:" -ForegroundColor Cyan
        Write-Host "  - Compiled executables cannot access Win32 device paths (\\.\$($DriveLetter):)" -ForegroundColor Gray
        Write-Host ""
        Write-Host "Solutions:" -ForegroundColor Cyan
        Write-Host "  1. Use the PowerShell script for drive letters:" -ForegroundColor Yellow
        Write-Host "     powershell -File sha256media.ps1 $($DriveLetter):" -ForegroundColor White
        Write-Host ""
        Write-Host "  2. If the drive contains an ISO file, hash the file directly:" -ForegroundColor Yellow
        Write-Host "     sha256media.exe $($DriveLetter):\path\to\file.iso" -ForegroundColor White
        Write-Host ""
        $script:hasErrors = $true
        exit 1
    }
    
    # For drives (not in compiled exe mode), use built-in calculation with Win32 device paths
    if ($IsDrive) {
        $sha = [System.Security.Cryptography.SHA256]::Create()
        $stream = $null
        
        try {
            Write-Host "Calculating SHA256 hash for drive '$($DriveLetter.ToUpper()):' (this can be slow)..."
            $devicePath = "\\.\${DriveLetter}:"
            # Use FileStream constructor instead of File.OpenRead for Win32 device support
            $stream = New-Object System.IO.FileStream($devicePath, [System.IO.FileMode]::Open, [System.IO.FileAccess]::Read, [System.IO.FileShare]::Read)
            
            $hashBytes = $sha.ComputeHash($stream)
            return [System.BitConverter]::ToString($hashBytes).Replace("-", "").ToLower()
        }
        finally {
            if ($stream) { $stream.Close() }
            if ($sha) { $sha.Dispose() }
        }
    }
    
    # For files, delegate to Get-Sha256Hash
    return Get-Sha256Hash -FilePath $TargetPath
}

# --- Main Script Body ---

Write-Host "=== SHA256 Media Hash Calculator ===" -ForegroundColor Cyan
Write-Host ""

# Detect if running in ps2exe compiled executable
try {
    $currentProcess = [System.Diagnostics.Process]::GetCurrentProcess()
    $processName = $currentProcess.ProcessName
    if ($processName -notmatch '^(powershell|pwsh)$') {
        $script:isCompiledExe = $true
    }
} catch {
    $script:isCompiledExe = $false
}

# Validate path
$isDrive = $false
$driveLetter = ''
$ResolvedPath = $null

if ($Path -match '^([A-Za-z]):\\?$') {
    $driveLetter = $Matches[1]
    try {
        $volume = Get-Volume -DriveLetter $driveLetter -ErrorAction Stop
        $isDrive = $true
        $ResolvedPath = $Path
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
}

# Calculate hash
$hash = Get-Sha256FromPath -TargetPath $ResolvedPath -IsDrive $isDrive -DriveLetter $driveLetter

if ($hash) {
    Write-Host ""
    Write-Host "SHA256: " -NoNewline -ForegroundColor Cyan
    Write-Host $hash -ForegroundColor Green
    
    # Save to output file if specified
    if ($Output) {
        try {
            $fileName = if ($isDrive) { "$($driveLetter):" } else { (Get-Item $ResolvedPath).Name }
            "$hash  $fileName" | Out-File -FilePath $Output -Encoding utf8NoBOM
            Write-Host ""
            Write-Host "Hash saved to: $Output" -ForegroundColor Green
        } catch {
            Write-Warning "Failed to save hash to file: $_"
        }
    }
    
    exit 0
} else {
    Write-Error "Failed to calculate hash"
    exit 1
}

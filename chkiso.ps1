<#
.SYNOPSIS
    Verifies the integrity of an ISO file or a physical disc using one or more methods.

.DESCRIPTION
    This script provides multiple methods for verification. You can run multiple checks in a single command.

    1. Implanted MD5 Check: Use the -MD5 switch to emulate the 'checkisomd5' utility on an ISO or physical drive. If checkisomd5.exe is found in the current directory or PATH, it will be used instead of the built-in check to avoid FIPS restrictions.

    2. External SHA256 Hash String: Use the -Sha256Hash (or its alias -sha256sum) parameter to verify the ISO or physical drive against a provided hash string.

    3. External SHA256 File: Use the -ShaFile parameter to verify the ISO or physical drive against a hash found in a sha256sum-compatible file.

    4. Verify Contents: By default, the script checks the integrity of the files on an ISO or physical drive against any embedded checksum files (*.sha, sha256sum.txt). Use the -NoVerify switch to skip this check.

.PARAMETER Path
    The path to the ISO file or the drive letter of the physical disc to verify (e.g., "D:").

.PARAMETER Sha256Hash
    Specifies the expected SHA256 hash for verification. Alias: sha256sum.

.PARAMETER ShaFile
    Specifies the path to a text file containing SHA256 hashes.

.PARAMETER NoVerify
    A switch that tells the script to skip verifying the hashes of the internal files.

.PARAMETER MD5
    A switch to enable the implanted MD5 check. If checkisomd5.exe is found in the current directory or PATH, it will be used instead of the built-in check to avoid FIPS restrictions.

.PARAMETER Dismount
    A switch that dismounts the ISO or ejects the physical drive after verification.
#>
[CmdletBinding()]
param (
    [Parameter(Mandatory=$true, Position=0, ValueFromPipeline=$true)]
    [string]$Path,

    # FIX: Restored positional parameter
    [Parameter(Mandatory=$false, Position=1)]
    [Alias('sha256sum', 'sha256', 'sha','shaString', 'HashString')]
    [string]$Sha256Hash,

    [Parameter(Mandatory=$false)]
    [Alias('sha256file', 'HashFile')]
    [string]$ShaFile,

    [Parameter(Mandatory=$false)]
    [Alias('ValidateContents', 'CheckContents')]
    [switch]$NoVerify,

    [Parameter(Mandatory=$false)]
    [switch]$MD5,

    [Parameter(Mandatory=$false)]
    [Alias('umount', 'unmount', 'eject')]
    [switch]$Dismount
)

# --- Error Tracking ---
# Track if any errors occurred during verification
$script:hasErrors = $false

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
    
    # For drives, use built-in calculation (sha256sum.exe can't handle Win32 device paths)
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
    
    # For files, check if sha256sum.exe utility is available
    $sha256sumPath = Get-Sha256sumPath
    
    if ($sha256sumPath) {
        Write-Host "Found sha256sum.exe, using it to calculate hash..." -ForegroundColor Green
    }
    
    # Delegate to Get-Sha256Hash for file operations (handles both utility and built-in)
    # Use -Quiet:$false to ensure messages are displayed
    return Get-Sha256Hash -FilePath $TargetPath
}

function Verify-PathAgainstHashString {
    param ([string]$Path, [string]$ExpectedHash, [bool]$IsDrive, [string]$DriveLetter)
    Write-Host "`n--- Verifying Path Against Provided SHA256 Hash ---" -ForegroundColor Cyan
    $ExpectedHash = $ExpectedHash.Trim().ToLower()
    $calculatedHash = Get-Sha256FromPath -TargetPath $Path -IsDrive $IsDrive -DriveLetter $DriveLetter
    Write-Host "  - Expected:   $ExpectedHash"
    Write-Host "  - Calculated: $calculatedHash"
    if ($calculatedHash -eq $ExpectedHash) { 
        Write-Host "Result: SUCCESS - Hashes match." -ForegroundColor Green 
    }
    else { 
        Write-Host "Result: FAILURE - Hashes DO NOT match." -ForegroundColor Red 
        $script:hasErrors = $true
    }
}

function Verify-PathAgainstHashFile {
    param ([string]$Path, [string]$HashFilePath, [bool]$IsDrive, [string]$DriveLetter)
    Write-Host "`n--- Verifying Path Against SHA256 Hash File ---" -ForegroundColor Cyan
    try { $HashFileResolved = (Resolve-Path -LiteralPath $HashFilePath.Trim("`"")).Path }
    catch { 
        Write-Error "Hash file not found: $HashFilePath"
        $script:hasErrors = $true
        return 
    }
    
    $isoFileNamePattern = if ($IsDrive) { "*.iso" } else { (Get-Item -LiteralPath $Path).Name }
    
    # FIX: Corrected Regex
    $pattern = "^([a-fA-F0-9]{64})\s+\*?\s*$([regex]::Escape($isoFileNamePattern))"
    $genericPattern = "^([a-fA-F0-9]{64})\s+\*?\s*.*"

    $matchInfo = Get-Content $HashFileResolved | Select-String -Pattern $pattern | Select-Object -First 1
    if (-not $matchInfo) {
        $matchInfo = Get-Content $HashFileResolved | Select-String -Pattern $genericPattern | Select-Object -First 1
    }

    if (-not $matchInfo) { 
        Write-Error "Could not find a valid SHA256 hash entry in the hash file '$HashFileResolved'."
        $script:hasErrors = $true
        return 
    }
    
    $expectedHash = $matchInfo.Matches[0].Groups[1].Value.ToLower()
    Verify-PathAgainstHashString -Path $Path -ExpectedHash $expectedHash -IsDrive $IsDrive -DriveLetter $DriveLetter
}

function Verify-Contents {
    param ([string]$Path, [bool]$IsDrive, [string]$DriveLetter)
    Write-Host "`n--- Verifying Contents ---" -ForegroundColor Cyan
    $diskImage = $null; $mountPath = ''
    try {
        if ($IsDrive) {
            $mountPath = "$($DriveLetter):\"
            Write-Host "Verifying contents of physical drive at: $mountPath" -ForegroundColor Green
            # Track the drive letter for consistency with ISO mounting and to support proper dismount logic
            $script:mountedDriveLetter = $DriveLetter
        } else {
            Write-Host "Mounting ISO..."
            $diskImage = Mount-DiskImage -ImagePath $Path -PassThru
            $driveLetter = ($diskImage | Get-Volume).DriveLetter
            if ([string]::IsNullOrWhiteSpace($driveLetter)) { 
                Write-Error "Could not get drive letter."
                $script:hasErrors = $true
                return 
            }
            $mountPath = "${driveLetter}:\"
            Write-Host "ISO mounted at: $mountPath" -ForegroundColor Green
            
            # Track the mounted drive letter for proper cleanup
            $script:mountedDriveLetter = $driveLetter
        }
        
        $checksumFiles = Get-ChildItem -Path $mountPath -Recurse -Include "*.sha", "sha256sum.txt", "SHA256SUMS"
        if (-not $checksumFiles) { Write-Warning "Could not find any checksum files (*.sha, sha256sum.txt, etc.) on the media."; return }

        $totalFiles = 0; $failedFiles = 0
        foreach ($checksumFile in $checksumFiles) {
            Write-Host "`nProcessing checksum file: $($checksumFile.FullName)" -ForegroundColor Yellow
            $baseDir = $checksumFile.DirectoryName
            Get-Content -Path $checksumFile.FullName | ForEach-Object {
                # FIX: Corrected Regex
                if ($_ -match "^([a-fA-F0-9]{64})\s+[\*.\/\\]*(.*)") {
                    $totalFiles++; $expectedHash = $matches[1].ToLower(); $fileName = $matches[2].Trim()
                    $filePathOnMedia = Join-Path -Path $baseDir -ChildPath $fileName
                    if (-not (Test-Path $filePathOnMedia)) {
                        Write-Warning "File not found on media: $fileName (referenced in $($checksumFile.Name))"; $failedFiles++; return
                    }
                    Write-Host "Verifying: $fileName" -NoNewline
                    $calculatedHash = Get-Sha256Hash -FilePath $filePathOnMedia -Quiet
                    if ($calculatedHash -eq $expectedHash) { Write-Host " -> OK" -ForegroundColor Green }
                    else { Write-Host " -> FAILED" -ForegroundColor Red; $failedFiles++ }
                }
            }
        }
        Write-Host "`n--- Verification Summary ---" -ForegroundColor Cyan
        if ($failedFiles -eq 0 -and $totalFiles -gt 0) { 
            Write-Host "Success: All $totalFiles files verified." -ForegroundColor Green 
        }
        else { 
            Write-Host "Failure: $failedFiles out of $totalFiles files failed." -ForegroundColor Red
            $script:hasErrors = $true
        }
    } catch { 
        Write-Error "An error occurred: $_"
        $script:hasErrors = $true
    }
    # The finally block is removed from here; dismount/eject is handled at the end of the script.
}

function Verify-ImplantedIsoMd5 {
    param ([string]$Path, [bool]$IsDrive, [string]$DriveLetter)
    Write-Host "`n--- Verifying Implanted ISO MD5 (checkisomd5 compatible) ---" -ForegroundColor Cyan
    
    # First, check if checkisomd5.exe is available
    $checkisomd5Path = $null
    
    # Check current directory first
    if (Test-Path ".\checkisomd5.exe") {
        $checkisomd5Path = (Resolve-Path ".\checkisomd5.exe").Path
        Write-Host "Found checkisomd5.exe in current directory: $checkisomd5Path" -ForegroundColor Green
    }
    else {
        # Check if it's in PATH
        $checkisomd5Cmd = Get-Command checkisomd5.exe -ErrorAction SilentlyContinue
        if ($checkisomd5Cmd) {
            $checkisomd5Path = $checkisomd5Cmd.Source
            Write-Host "Found checkisomd5.exe in PATH: $checkisomd5Path" -ForegroundColor Green
        }
    }
    
    # If checkisomd5.exe is found, use it instead of built-in check
    if ($checkisomd5Path) {
        Write-Host "Using external checkisomd5.exe to avoid FIPS restrictions..."
        try {
            # Verify the executable is accessible
            if (-not (Test-Path $checkisomd5Path -PathType Leaf)) {
                throw "checkisomd5.exe path is not accessible"
            }
            
            $targetPath = if ($IsDrive) { "\\.\${DriveLetter}:" } else { $Path }
            $output = & $checkisomd5Path $targetPath 2>&1
            $exitCode = $LASTEXITCODE
            
            # Display output properly
            if ($output) {
                $output | ForEach-Object { Write-Host $_ }
            }
            
            if ($exitCode -eq 0) {
                Write-Host "`nSUCCESS: Implanted MD5 is valid (verified with checkisomd5.exe)." -ForegroundColor Green
            }
            else {
                Write-Warning "`nFAILURE: Implanted MD5 verification failed (checkisomd5.exe exit code: $exitCode)."
                $script:hasErrors = $true
            }
            return
        }
        catch {
            Write-Warning "Failed to run checkisomd5.exe: $_"
            Write-Host "Falling back to built-in MD5 check..."
        }
    }
    
    # Use built-in check if checkisomd5.exe is not available or failed
    $result = Invoke-ImplantedMd5Check -Path $Path -IsDrive $IsDrive -DriveLetter $DriveLetter
    if ($result) {
        if ($result.VerificationMethod -eq 'FIPS_BLOCKED') {
            Write-Warning "Found implanted MD5 hash: $($result.StoredMD5)"
            Write-Warning "Verification blocked by system FIPS security policy."
        } else {
            $result | Format-List
            if ($result.IsIntegrityOK) { 
                Write-Host "`nSUCCESS: Implanted MD5 is valid." -ForegroundColor Green 
            }
            else { 
                Write-Warning "`nFAILURE: Implanted MD5 does not match calculated hash."
                $script:hasErrors = $true
            }
        }
    } else {
        # If result is null, an error occurred
        $script:hasErrors = $true
    }
}

function Invoke-ImplantedMd5Check {
    param ([string]$Path, [bool]$IsDrive, [string]$DriveLetter)
    $PVD_OFFSET = 32768; $PVD_SIZE = 2048; $APP_USE_OFFSET_IN_PVD = 883; $APP_USE_SIZE = 512; $SECTOR_SIZE = 2048
    $fileStream = $null; $md5 = $null
    try {
        $streamPath = if ($IsDrive) { "\\.\${DriveLetter}:" } else { $Path }
        # Use FileStream constructor instead of File.OpenRead for Win32 device support
        $fileStream = New-Object System.IO.FileStream($streamPath, [System.IO.FileMode]::Open, [System.IO.FileAccess]::Read, [System.IO.FileShare]::Read)
        $fileLength = $fileStream.Length

        $pvdBlock = New-Object byte[] $PVD_SIZE
        $fileStream.Seek($PVD_OFFSET, [System.IO.SeekOrigin]::Begin) | Out-Null
        if ($fileStream.Read($pvdBlock, 0, $PVD_SIZE) -ne $PVD_SIZE) { 
            Write-Error "Could not read PVD."
            $script:hasErrors = $true
            return $null 
        }
        
        $appUseString = [System.Text.Encoding]::ASCII.GetString($pvdBlock, $APP_USE_OFFSET_IN_PVD, $APP_USE_SIZE)
        $md5Match = [regex]::Match($appUseString, 'ISO MD5SUM = ([0-9a-fA-F]{32})')
        if (-not $md5Match.Success) { Write-Warning "FAILURE: No 'ISO MD5SUM' signature found."; return $null }
        $storedHash = $md5Match.Groups[1].Value.ToLower()

        try { $md5 = [System.Security.Cryptography.MD5CryptoServiceProvider]::new() }
        catch [System.InvalidOperationException] {
            if ($_.Exception.Message -like '*FIPS*') { return [PSCustomObject]@{ VerificationMethod = 'FIPS_BLOCKED'; StoredMD5 = $storedHash } }
            else { throw }
        }

        $skipSectors = 0; $skipMatch = [regex]::Match($appUseString, 'SKIPSECTORS\s*=\s*(\d+)'); if ($skipMatch.Success) { $skipSectors = [int]$skipMatch.Groups[1].Value }
        $hashEndOffset = $fileLength - ($skipSectors * $SECTOR_SIZE)
        $implantDataLength = $APP_USE_SIZE  # Clear entire 512-byte Application Use field
        $neutralizedPvd = $pvdBlock.Clone()
        # Fill with spaces (0x20), not zeros - this matches libcheckisomd5 behavior
        for ($i = 0; $i -lt $implantDataLength; $i++) { $neutralizedPvd[$APP_USE_OFFSET_IN_PVD + $i] = 0x20 }

        $fileStream.Seek(0, [System.IO.SeekOrigin]::Begin) | Out-Null; $buffer = New-Object byte[] 65536
        # Part A
        $totalToRead = $PVD_OFFSET
        while ($totalToRead -gt 0) {
            $bytesToRead = [System.Math]::Min($buffer.Length, $totalToRead); $bytesRead = $fileStream.Read($buffer, 0, $bytesToRead)
            if ($bytesRead -eq 0) { break }; $md5.TransformBlock($buffer, 0, $bytesRead, $null, 0) | Out-Null; $totalToRead -= $bytesRead
        }
        # Part B
        $md5.TransformBlock($neutralizedPvd, 0, $neutralizedPvd.Length, $null, 0) | Out-Null
        $fileStream.Seek($PVD_OFFSET + $PVD_SIZE, [System.IO.SeekOrigin]::Begin) | Out-Null
        # Part C
        $totalToRead = $hashEndOffset - $fileStream.Position
        while ($totalToRead -gt 0) {
            $bytesToRead = [System.Math]::Min($buffer.Length, $totalToRead); $bytesRead = $fileStream.Read($buffer, 0, $bytesToRead)
            if ($bytesRead -eq 0) { break }
            $md5.TransformBlock($buffer, 0, $bytesRead, $null, 0) | Out-Null
            $totalToRead -= $bytesRead
        }
        $md5.TransformFinalBlock(@(), 0, 0) | Out-Null
        $calculatedMd5Hex = [System.BitConverter]::ToString($md5.Hash).Replace("-", "").ToLower()
        return [PSCustomObject]@{ VerificationMethod = "ASCII String (checkisomd5 compatible)"; StoredMD5 = $storedHash; CalculatedMD5 = $calculatedMd5Hex; IsIntegrityOK = $storedHash -eq $calculatedMd5Hex }
    } catch { 
        Write-Error "An error occurred during check: $_"
        $script:hasErrors = $true
    }
    finally { if ($fileStream) { $fileStream.Close() }; if ($md5) { $md5.Dispose() } }
    return $null
}

# --- Main Script Body ---

# Track mounted drive letter for proper cleanup
$script:mountedDriveLetter = $null

# FIX: Robust path validation at the start of the script.
$isDrive = $false
$driveLetter = ''
$ResolvedPath = $null # Start with null

if ($Path -match '^([A-Za-z]):\\?$') {
    $driveLetter = $Matches[1]
    try {
        $volume = Get-Volume -DriveLetter $driveLetter -ErrorAction Stop
        # Check if this is an optical drive (CD-ROM, DVD, Blu-ray, etc.)
        # Note: Windows reports most optical drives as 'CD-ROM' including DVD and Blu-ray
        # Some systems may report newer optical drives (like BD-RE) differently
        # If your drive is not recognized, you can pass the ISO file path directly instead
        $acceptedDriveTypes = @('CD-ROM')  # Primary type for optical drives
        
        # Additional check: If drive type is Unknown, we'll accept it if the user explicitly
        # passed a drive letter, assuming they know what they're doing. However, this could
        # cause errors if it's not actually an optical drive.
        if ($volume.DriveType -eq 'Unknown') {
            Write-Warning "Drive type is 'Unknown'. Attempting to treat as optical drive. If you encounter errors, please use the ISO file path directly."
            $isOpticalDrive = $true
        } else {
            $isOpticalDrive = $volume.DriveType -in $acceptedDriveTypes
        }
        
        if ($isOpticalDrive) {
            # Detect if running in ps2exe compiled executable
            $isCompiledExe = $false
            try {
                $currentProcess = [System.Diagnostics.Process]::GetCurrentProcess()
                $processName = $currentProcess.ProcessName
                # If not running as powershell or pwsh, likely a compiled exe
                if ($processName -notmatch '^(powershell|pwsh)$') {
                    $isCompiledExe = $true
                }
            } catch {
                # If we can't determine, assume not compiled
                $isCompiledExe = $false
            }
            
            if ($isCompiledExe) {
                # In compiled exe, we can't use Win32 device paths for drive letters
                # This applies to both mounted ISOs and physical drives
                # Exit with 0 (success) since this is informational guidance, not an error
                Write-Host "`nNote: When using the compiled executable (chkiso.exe), drive letters (e.g., E:) are not supported due to technical limitations with Win32 device paths." -ForegroundColor Yellow
                Write-Host "`nPlease use one of these alternatives:" -ForegroundColor Yellow
                Write-Host "  1. Use the ISO file path directly: chkiso.exe C:\path\to\image.iso" -ForegroundColor Yellow
                Write-Host "  2. Use the PowerShell script instead: powershell -File chkiso.ps1 E:" -ForegroundColor Yellow
                exit 0
            }
            
            # For regular PowerShell with optical drives, treat as a physical drive
            # Note: We skip mounted ISO detection to avoid Get-DiskImage parameter prompts
            # Win32 device path (\\.\X:) will be constructed later when IsDrive is true
            $isDrive = $true
            $ResolvedPath = $Path # For a drive, the path is just the letter (e.g., "E:" or "E:\")
        } else {
            Write-Error "Path '$Path' is a drive, but not an optical drive (CD/DVD/Blu-ray). DriveType detected: $($volume.DriveType)"; exit
        }
    } catch {
        Write-Error "Drive '$Path' not found or is not ready."; exit
    }
} else {
    try {
        $ResolvedPath = (Resolve-Path -LiteralPath $Path.Trim("`"")).Path
    } catch {
        Write-Error "File not found: $Path"; exit
    }
    if (-not (Test-Path $ResolvedPath -PathType Leaf)) {
        Write-Error "Path is not a file: $ResolvedPath"; exit
    }
}

# Execute checks based on provided parameters
if ($PSBoundParameters.ContainsKey('ShaFile')) {
    Verify-PathAgainstHashFile -Path $ResolvedPath -HashFilePath $ShaFile -IsDrive $isDrive -DriveLetter $driveLetter
}
if ($PSBoundParameters.ContainsKey('Sha256Hash')) {
    Verify-PathAgainstHashString -Path $ResolvedPath -ExpectedHash $Sha256Hash -IsDrive $isDrive -DriveLetter $driveLetter
}
# If neither Sha256Hash nor ShaFile is provided, just display the SHA256 sum for informational purposes
if (-not $PSBoundParameters.ContainsKey('Sha256Hash') -and -not $PSBoundParameters.ContainsKey('ShaFile')) {
    Write-Host "`n--- SHA256 Hash (Informational) ---" -ForegroundColor Cyan
    $calculatedHash = Get-Sha256FromPath -TargetPath $ResolvedPath -IsDrive $isDrive -DriveLetter $driveLetter
    Write-Host "SHA256: $calculatedHash" -ForegroundColor Yellow
}
if ($PSBoundParameters.ContainsKey('MD5')) {
    Verify-ImplantedIsoMd5 -Path $ResolvedPath -IsDrive $isDrive -DriveLetter $driveLetter
}
# Run VerifyContents by default unless -NoVerify is specified
if (-not $PSBoundParameters.ContainsKey('NoVerify')) {
    Verify-Contents -Path $ResolvedPath -IsDrive $isDrive -DriveLetter $driveLetter
}

if ($PSBoundParameters.ContainsKey('Dismount')) {
    if ($isDrive) {
        # Eject the drive letter that was explicitly passed in as the Path parameter
        # This is safe because the user explicitly specified this drive
        try {
            Write-Host "`nEjecting drive $driveLetter`:" -ForegroundColor Yellow
            $shell = New-Object -ComObject Shell.Application
            $shell.Namespace(17).ParseName($driveLetter + ":").InvokeVerb("Eject")
        } catch { Write-Error "Failed to eject drive $driveLetter`: . $_" }
    } else {
        # For ISO files, only dismount if we actually mounted it during this script execution
        # This prevents dismounting ISOs that were already mounted before running this script
        if ($script:mountedDriveLetter) {
            $diskImage = Get-DiskImage -ImagePath $ResolvedPath -ErrorAction SilentlyContinue
            if ($diskImage -and $diskImage.Attached) {
                Write-Host "`nDismounting ISO from drive $script:mountedDriveLetter`:..." -ForegroundColor Yellow
                Dismount-DiskImage -ImagePath $ResolvedPath | Out-Null
            } else {
                Write-Warning "ISO is not currently mounted."
            }
        } else {
            Write-Warning "Skipping dismount: ISO was not mounted during VerifyContents. The script only dismounts ISOs it explicitly mounted."
        }
    }
}

# Exit with proper code based on whether errors occurred
if ($script:hasErrors) {
    exit 1
} else {
    exit 0
}

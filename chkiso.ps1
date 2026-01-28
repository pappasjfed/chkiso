<#
.SYNOPSIS
    Verifies the integrity of an ISO file or a physical disc using one or more methods.

.DESCRIPTION
    This script provides multiple methods for verification. You can run multiple checks in a single command.

    1. Implanted MD5 Check: If not disabled with -NoMD5, this emulates the 'checkisomd5' utility on an ISO or physical drive.

    2. External SHA256 Hash String: Use the -Sha256Hash (or its alias -sha256sum) parameter to verify the ISO or physical drive against a provided hash string.

    3. External SHA256 File: Use the -ShaFile parameter to verify the ISO or physical drive against a hash found in a sha256sum-compatible file.

    4. Verify Contents: Use the -VerifyContents switch to check the integrity of the files on an ISO or physical drive against any embedded checksum files (*.sha, sha256sum.txt).

.PARAMETER Path
    The path to the ISO file or the drive letter of the physical disc to verify (e.g., "D:").

.PARAMETER Sha256Hash
    Specifies the expected SHA256 hash for verification. Alias: sha256sum.

.PARAMETER ShaFile
    Specifies the path to a text file containing SHA256 hashes.

.PARAMETER VerifyContents
    A switch that tells the script to verify the hashes of the internal files.

.PARAMETER NoMD5
    A switch to disable the implanted MD5 check.

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
    [switch]$VerifyContents,

    [Parameter(Mandatory=$false)]
    [switch]$NoMD5,

    [Parameter(Mandatory=$false)]
    [Alias('umount', 'unmount', 'eject')]
    [switch]$Dismount
)

# --- Helper Functions ---

function Get-Sha256FromPath {
    param( [string]$TargetPath, [bool]$IsDrive, [string]$DriveLetter )
    if ($IsDrive) {
        Write-Host "Calculating SHA256 hash for drive '$($DriveLetter.ToUpper()):' (this can be slow)..."
        $devicePath = "\\.\$DriveLetter`:"
        $sha = [System.Security.Cryptography.SHA256]::Create()
        try {
            $stream = [System.IO.File]::OpenRead($devicePath)
            $hashBytes = $sha.ComputeHash($stream)
            return [System.BitConverter]::ToString($hashBytes).Replace("-", "").ToLower()
        }
        finally {
            if ($stream) { $stream.Close() }
            if ($sha) { $sha.Dispose() }
        }
    } else {
        Write-Host "Calculating SHA256 hash for file '$((Get-Item $TargetPath).Name)'..."
        return (Get-FileHash -Path $TargetPath -Algorithm SHA256).Hash.ToLower()
    }
}

function Verify-PathAgainstHashString {
    param ([string]$Path, [string]$ExpectedHash, [bool]$IsDrive, [string]$DriveLetter)
    Write-Host "`n--- Verifying Path Against Provided SHA256 Hash ---" -ForegroundColor Cyan
    $ExpectedHash = $ExpectedHash.Trim().ToLower()
    $calculatedHash = Get-Sha256FromPath -TargetPath $Path -IsDrive $IsDrive -DriveLetter $DriveLetter
    Write-Host "  - Expected:   $ExpectedHash"
    Write-Host "  - Calculated: $calculatedHash"
    if ($calculatedHash -eq $ExpectedHash) { Write-Host "Result: SUCCESS - Hashes match." -ForegroundColor Green }
    else { Write-Host "Result: FAILURE - Hashes DO NOT match." -ForegroundColor Red }
}

function Verify-PathAgainstHashFile {
    param ([string]$Path, [string]$HashFilePath, [bool]$IsDrive, [string]$DriveLetter)
    Write-Host "`n--- Verifying Path Against SHA256 Hash File ---" -ForegroundColor Cyan
    try { $HashFileResolved = (Resolve-Path -LiteralPath $HashFilePath.Trim("`"")).Path }
    catch { Write-Error "Hash file not found: $HashFilePath"; return }
    
    $isoFileNamePattern = if ($IsDrive) { "*.iso" } else { (Get-Item -LiteralPath $Path).Name }
    
    # FIX: Corrected Regex
    $pattern = "^([a-fA-F0-9]{64})\s+\*?\s*$([regex]::Escape($isoFileNamePattern))"
    $genericPattern = "^([a-fA-F0-9]{64})\s+\*?\s*.*"

    $matchInfo = Get-Content $HashFileResolved | Select-String -Pattern $pattern | Select-Object -First 1
    if (-not $matchInfo) {
        $matchInfo = Get-Content $HashFileResolved | Select-String -Pattern $genericPattern | Select-Object -First 1
    }

    if (-not $matchInfo) { Write-Error "Could not find a valid SHA256 hash entry in the hash file '$HashFileResolved'."; return }
    
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
            if ([string]::IsNullOrWhiteSpace($driveLetter)) { Write-Error "Could not get drive letter."; return }
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
                    $calculatedHash = (Get-FileHash -Path $filePathOnMedia -Algorithm SHA256).Hash.ToLower()
                    if ($calculatedHash -eq $expectedHash) { Write-Host " -> OK" -ForegroundColor Green }
                    else { Write-Host " -> FAILED" -ForegroundColor Red; $failedFiles++ }
                }
            }
        }
        Write-Host "`n--- Verification Summary ---" -ForegroundColor Cyan
        if ($failedFiles -eq 0 -and $totalFiles -gt 0) { Write-Host "Success: All $totalFiles files verified." -ForegroundColor Green }
        else { Write-Host "Failure: $failedFiles out of $totalFiles files failed." -ForegroundColor Red }
    } catch { Write-Error "An error occurred: $_" }
    # The finally block is removed from here; dismount/eject is handled at the end of the script.
}

function Verify-ImplantedIsoMd5 {
    param ([string]$Path, [bool]$IsDrive, [string]$DriveLetter)
    Write-Host "`n--- Verifying Implanted ISO MD5 (checkisomd5 compatible) ---" -ForegroundColor Cyan
    $result = Invoke-ImplantedMd5Check -Path $Path -IsDrive $IsDrive -DriveLetter $DriveLetter
    if ($result) {
        if ($result.VerificationMethod -eq 'FIPS_BLOCKED') {
            Write-Warning "Found implanted MD5 hash: $($result.StoredMD5)"; Write-Warning "Verification blocked by system FIPS security policy."
        } else {
            $result | Format-List
            if ($result.IsIntegrityOK) { Write-Host "`nSUCCESS: Implanted MD5 is valid." -ForegroundColor Green }
            else { Write-Warning "`nFAILURE: Implanted MD5 does not match calculated hash." }
        }
    }
}

function Invoke-ImplantedMd5Check {
    param ([string]$Path, [bool]$IsDrive, [string]$DriveLetter)
    $PVD_OFFSET = 32768; $PVD_SIZE = 2048; $APP_USE_OFFSET_IN_PVD = 883; $APP_USE_SIZE = 512; $SECTOR_SIZE = 2048
    $fileStream = $null; $md5 = $null
    try {
        $streamPath = if ($IsDrive) { "\\.\$DriveLetter`:" } else { $Path }
        $fileStream = [System.IO.File]::OpenRead($streamPath)
        $fileLength = $fileStream.Length

        $pvdBlock = New-Object byte[] $PVD_SIZE
        $fileStream.Seek($PVD_OFFSET, [System.IO.SeekOrigin]::Begin) | Out-Null
        if ($fileStream.Read($pvdBlock, 0, $PVD_SIZE) -ne $PVD_SIZE) { Write-Error "Could not read PVD."; return $null }
        
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
        $neutralizedPvd = $pvdBlock.Clone(); [System.Array]::Clear($neutralizedPvd, $APP_USE_OFFSET_IN_PVD, $implantDataLength)

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
    } catch { Write-Error "An error occurred during check: $_" }
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
        if ($volume.DriveType -eq 'CD-ROM') {
            $isDrive = $true
            $ResolvedPath = $Path # For a drive, the path is just the letter (e.g., "E:")
        } else {
            Write-Error "Path '$Path' is a drive, but not a CD/DVD drive."; exit
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
if (-not $PSBoundParameters.ContainsKey('NoMD5')) {
    Verify-ImplantedIsoMd5 -Path $ResolvedPath -IsDrive $isDrive -DriveLetter $driveLetter
}
if ($PSBoundParameters.ContainsKey('VerifyContents')) {
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

# What is this?
[![Test](https://github.com/pappasjfed/chkiso/actions/workflows/test.yml/badge.svg)](https://github.com/pappasjfed/chkiso/actions/workflows/test.yml)
[![Build and Release](https://github.com/pappasjfed/chkiso/actions/workflows/build-release.yml/badge.svg)](https://github.com/pappasjfed/chkiso/actions/workflows/build-release.yml)

This is a project that is used to validate ISO images created by standard DSO pipelines.  It checks hashes for ISOs or media.

## Usage

### Basic Usage
By default, the script:
- Displays the SHA256 hash of the ISO/drive
- Verifies internal file integrity against embedded checksum files (*.sha, sha256sum.txt)

```powershell
.\chkiso.ps1 path\to\image.iso
```

### Advanced Options

#### Check against an expected SHA256 hash:
```powershell
.\chkiso.ps1 path\to\image.iso <sha256-hash>
```
**Note**: Content verification runs by default. Use `-NoVerify` to skip it.

#### Check against a hash file:
```powershell
.\chkiso.ps1 path\to\image.iso -ShaFile path\to\hashfile.sha
```
**Note**: Content verification runs by default. Use `-NoVerify` to skip it.

#### Enable implanted MD5 check:
```powershell
.\chkiso.ps1 path\to\image.iso -MD5
```

**Note**: If `checkisomd5.exe` is available in the current directory or PATH, it will be used automatically to avoid FIPS restrictions.

**Note**: SHA256 calculations will automatically use `sha256sum.exe` if available in the current directory or PATH for improved performance. The script will fall back to built-in calculation if the utility is not found.

#### Skip internal file verification:
```powershell
.\chkiso.ps1 path\to\image.iso -NoVerify
```

### Windows Executable
Download the compiled `chkiso.exe` from the [Releases](https://github.com/pappasjfed/chkiso/releases) page and run it:
```cmd
chkiso.exe path\to\image.iso
```

**Important Limitation - Drive Letters**: The compiled executable (`chkiso.exe`) **cannot hash drive letters** directly due to ps2exe limitations with Win32 device paths. This affects both mounted ISOs and physical CD/DVD drives.

**For drive letter hashing, use the PowerShell script instead:**
```powershell
powershell -File chkiso.ps1 G:
```

**Alternative for compiled executable**: Hash the ISO file directly if accessible:
```cmd
chkiso.exe G:\path\to\image.iso
```

**Why this limitation exists:**
- Compiled ps2exe executables cannot access Win32 device paths (`\\.\G:`) needed for raw drive access
- Tools like `sha256sum.exe` can only hash files, not raw drive devices
- The PowerShell script (`chkiso.ps1`) uses native .NET FileStream with Win32 device paths and works with drive letters

## Building

The Windows executable is automatically built and attached to releases via GitHub Actions. 

### Local Build

To build the executable locally:

1. Run the build script:
   ```powershell
   .\build.ps1
   ```
   
   This will:
   - Install ps2exe if needed
   - Compile chkiso.ps1 to bin/chkiso.exe
   - Generate SHA256 checksum (bin/chkiso.exe.sha)

2. Optional: Clean build
   ```powershell
   .\build.ps1 -Clean
   ```

3. Optional: Download utility binaries
   ```powershell
   .\build.ps1 -DownloadUtilities
   ```
   
   This will download helper utilities to bin/:
   - `checkisomd5.exe` - From https://github.com/pappasjfed/isomd5sum/releases
   - `sha256sum.exe` - From GnuWin32 CoreUtils package
   
   These utilities enhance functionality:
   - checkisomd5.exe: Used for MD5 validation (avoids FIPS restrictions)
   - sha256sum.exe: Required for drive letter support in compiled exe

The compiled executable will be in the `bin/` directory.

### Manual Build

If you prefer to build manually:

1. Install ps2exe: `Install-Module -Name ps2exe -Force`
2. Compile: `ps2exe -inputFile chkiso.ps1 -outputFile bin/chkiso.exe -noConsole:$false -title "chkiso" -version "1.0.0.0" -company "chkiso" -product "chkiso" -copyright "MIT License"`

### Code Signing

The executable can be automatically signed if code signing certificates are configured. See [CODE_SIGNING.md](CODE_SIGNING.md) for setup instructions. If certificates are not configured, the workflow will build an unsigned executable.

## Testing

Tests run automatically on pull requests and pushes to main. The test suite validates the executable against `test/test.iso` using multiple verification methods.
